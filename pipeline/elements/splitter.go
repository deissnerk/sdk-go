package elements

import (
	"context"
	"github.com/cloudevents/sdk-go/pipeline"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/pkg/errors"
)

type splitterTaskKey struct {
	key    pipeline.TaskIndex
	subKey pipeline.TaskIndex
}

type TaskSplitter interface {
	Split(origin *pipeline.Task) []*pipeline.TaskAssignment

	// Take the status of all sub-tasks
	Join(status []*pipeline.TaskControl) *pipeline.ProcessorOutput

	IsFinished(status []*pipeline.TaskControl) bool

	pipeline.StartStop
}

type Splitter struct {
	state *SplitterState
	sv    *pipeline.Supervisor
}

func NewSplitter(ts TaskSplitter, id pipeline.ElementId, nextStep pipeline.Runner) *Splitter {
	state := &SplitterState{
		id:       id,
		ts:       ts,
		sw:       pipeline.NewSlidingWindow(pipeline.DefaultWS),
		nextStep: nextStep,
	}

	return &Splitter{
		state: state,
		sv:pipeline.NewSupervisor(state),
	}

}

func (s Splitter) Start() error{
	// The Supervisor will start the state and all related resources
	return s.sv.Start()
}

func (s Splitter) Stop(ctx context.Context) {
	s.sv.Stop(ctx)
}

func (s Splitter) Id() pipeline.ElementId {
	return s.state.id
}

var _ pipeline.Element = (*Splitter)(nil)

//func (s Splitter) SetNextStep(runner pipeline.Runner) {
//	s.state.SetNextStep(runner)
//}

// The default SupervisorState implementation. It maintains a window and sets finishes a task, when all sub-tasks
// are finished.
type SplitterState struct {
	id       pipeline.ElementId
	ts       TaskSplitter
//	maxWSize pipeline.TaskIndex
	//	wsc      WindowStateCallback
	sw       *pipeline.SlidingWindow
	nextStep pipeline.Runner
}

func (st *SplitterState) Id() pipeline.ElementId {
	return st.id
}

func (st *SplitterState) Start() error {
	return st.ts.Start()
}

// Stop() calls Stop() on the TaskSplitter and for the next step to propagate the Stop()
// call through the pipeline. The state is only stopped after the Supervisor was stopped. 
func (st *SplitterState) Stop(ctx context.Context) {
	st.ts.Stop(ctx)
//	st.nextStep.Stop()
}

//func (st *SplitterState) SetNextStep(step pipeline.Runner) {
//	st.nextStep = step
//}

func (st *SplitterState) AddTask(tc *pipeline.TaskContainer, callback chan *pipeline.StatusMessage) {
	tas := st.ts.Split(&tc.Task)

	if len(tas) == 0 {
		tc.SendStatusUpdate(st.id, errors.Errorf("TaskSplitter returned no tasks"), true)
		return
	}
	svt, key := st.sw.AddTask()

	tStat := make([]*pipeline.TaskControl, len(tas))

	for i, ta := range tas {
		tStat[i] = &pipeline.TaskControl{
			Status: pipeline.TaskStatus{
				Id: ta.Runner.Id(),
				//Key: splitterTaskKey{
				//	key:    key,
				//	subKey: pipeline.TaskIndex(i),
				//},
				Result:   nil,
				Finished: false,
			},
			Cancel: ta.Cancel,
		}
	}

	*svt = pipeline.SuperVisorTask{
		Main:  tc,
		TStat: tStat,
	}

	for i, ta := range tas {
		ta.Runner.Push(
			&pipeline.TaskContainer{
				Callback: callback,
				Key: splitterTaskKey{
					key:    key,
					subKey: pipeline.TaskIndex(i),
				},
				Task:    *ta.Task,
				Parent:  tc,
			})
	}
}

func (st *SplitterState) IsIdle() bool {
	return st.sw.IsEmpty()
}

// Returns true, if the window is empty
func (st *SplitterState) UpdateTask(sMsg *pipeline.StatusMessage) bool {
	tKey := sMsg.Key.(splitterTaskKey)
	svt := st.sw.GetSupervisorTask(tKey.key)
	tStat := svt.TStat.([]*pipeline.TaskControl)
	tStat[tKey.subKey].Status = sMsg.Status

	if st.ts.IsFinished(tStat) {
		joinStat := st.ts.Join(tStat)
		if st.nextStep != nil && protocol.IsACK(joinStat.Result) {
			// If there is a next step, push the main task to it and send the result as a status update
			svt.Main.SendStatusUpdate(st.id, joinStat.Result, false)
			st.nextStep.Push(svt.Main.FollowUp(joinStat))
		} else {
			svt.Main.SendStatusUpdate(st.id, joinStat.Result, true)
		}
		return st.sw.RemoveTask(tKey.key)
	}

	return false
}

var _ pipeline.SupervisorState = (*SplitterState)(nil)