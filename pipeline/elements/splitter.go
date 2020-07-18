package elements

import (
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/pkg/errors"
	pipeline2 "pipeline"
	"sync"
)

type splitterTaskKey struct {
	key    pipeline2.TaskIndex
	subKey pipeline2.TaskIndex
}

type TaskSplitter interface {
	Split(origin *pipeline2.Task) []*pipeline2.TaskAssignment

	// Take the status of all sub-tasks
	Join(status []*pipeline2.TaskControl) *pipeline2.ProcessorOutput

	IsFinished(status []*pipeline2.TaskControl) bool

	pipeline2.StartStop
}

type Splitter struct {
	state *SplitterState
	sv    *pipeline2.Supervisor
	//svWg   *sync.WaitGroup
	//mainWg *sync.WaitGroup
}

func (s Splitter) Start(wg *sync.WaitGroup) {
	// The Supervisor will start the state and all related resources
	s.sv.Start(wg)
}

func (s Splitter) Stop() {
	s.sv.Stop()
}

func (s Splitter) Id() pipeline2.ElementId {
	return s.state.id
}

func (s Splitter) SetNextStep(runner pipeline2.Runner) {
	s.state.SetNextStep(runner)
}

// The default SupervisorState implementation. It maintains a window and sets finishes a task, when all sub-tasks
// are finished.
type SplitterState struct {
	id       pipeline2.ElementId
	ts       TaskSplitter
	maxWSize pipeline2.TaskIndex
	//	wsc      WindowStateCallback
	sw       *pipeline2.SlidingWindow
	nextStep pipeline2.Runner
}

func (st *SplitterState) Id() pipeline2.ElementId {
	return st.id
}

func (st *SplitterState) Start(wg *sync.WaitGroup) {
	st.ts.Start(wg)
}

// Stop() calls Stop() on the TaskSplitter and for the next step to propagate the Stop()
// call through the pipeline. The state is only stopped after the Supervisor was stopped. 
func (st *SplitterState) Stop() {
	st.ts.Stop()
	st.nextStep.Stop()
}

func (st *SplitterState) SetNextStep(step pipeline2.Runner) {
	st.nextStep = step
}

func (st *SplitterState) AddTask(tc *pipeline2.TaskContainer, callback chan *pipeline2.StatusMessage) {
	tas := st.ts.Split(&tc.Task)

	if len(tas) == 0 {
		tc.SendStatusUpdate(st.id, errors.Errorf("TaskSplitter returned no tasks"), true)
		return
	}
	svt, key := st.sw.AddTask()

	tStat := make([]*pipeline2.TaskControl, len(tas))

	for i, ta := range tas {
		tStat[i] = &pipeline2.TaskControl{
			Status: pipeline2.TaskStatus{
				Id: nil,
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

	*svt = pipeline2.SuperVisorTask{
		Main:  tc,
		TStat: tStat,
	}

	for i, ta := range tas {
		ta.Runner.Push(
			&pipeline2.TaskContainer{
				Callback: callback,
				Key: splitterTaskKey{
					key:    key,
					subKey: pipeline2.TaskIndex(i),
				},
				Task:    *ta.Task,
				Parent:  tc,
				Changes: nil,
			})
	}
}

func (st *SplitterState) IsIdle() bool {
	return st.sw.IsEmpty()
}

// Returns true, if the window is empty
func (st *SplitterState) UpdateTask(sMsg *pipeline2.StatusMessage) bool {
	tKey := sMsg.Key.(splitterTaskKey)
	svt := st.sw.GetSupervisorTask(tKey.key)
	tStat := svt.TStat.([]*pipeline2.TaskControl)
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
