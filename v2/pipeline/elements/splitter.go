package elements

import (
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/pkg/errors"
	"sync"
)
import "github.com/cloudevents/sdk-go/v2/pipeline"

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
	state  *SplitterState
	sv     *pipeline.Supervisor
	svWg   *sync.WaitGroup
	mainWg *sync.WaitGroup
}

func (s Splitter) Start(wg *sync.WaitGroup) {
// The Supervisor will start the state and all related resources
	s.sv.Start(wg)
}

func (s Splitter) Stop() {
	s.sv.Stop()
}

func (s Splitter) Id() pipeline.ElementId {
	return s.state.id
}

func (s Splitter) SetNextStep(runner pipeline.Runner) {
	s.state.SetNextStep(runner)
}

// The default SupervisorState implementation. It maintains a window and sets finishes a task, when all sub-tasks
// are finished.
type SplitterState struct {
	id       pipeline.ElementId
	ts       TaskSplitter
	maxWSize pipeline.TaskIndex
	//	wsc      WindowStateCallback
	sw          *pipeline.SlidingWindow
	nextStep    pipeline.Runner
	hasNextStep bool
}

func (st *SplitterState) Id() pipeline.ElementId {
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

func (st *SplitterState) SetNextStep(step pipeline.Runner) {
	st.nextStep = step
	st.hasNextStep = true
}

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
				Changes: nil,
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
		if st.hasNextStep && protocol.IsACK(joinStat.Result) {
			// If there is a next step, push the main task to it and send the result as a status update
			st.nextStep.Push(svt.Main.FollowUp(joinStat))
			svt.Main.SendStatusUpdate(st.id, joinStat.Result, false)
		} else {
			svt.Main.SendStatusUpdate(st.id, joinStat.Result, true)
		}
		return st.sw.RemoveTask(tKey.key)
	}

	return false
}
