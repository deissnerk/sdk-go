package elements

import (
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/pkg/errors"
	"sync"
)
import "github.com/cloudevents/sdk-go/v2/pipeline"

type TaskSplitter interface {
	Split(origin *pipeline.Task, callback chan *pipeline.TaskStatus) []*pipeline.TaskAssignment

	// Take the status of all sub-tasks
	Join(status []*pipeline.TaskStatus) *pipeline.ProcessorOutput

	IsFinished(status []*pipeline.TaskStatus) bool

	pipeline.StartStop
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

func (st *SplitterState) Start(wg *sync.WaitGroup) {
	st.ts.Start(wg)
}

func (st *SplitterState) Stop() {
	st.ts.Stop()
}

func (st *SplitterState) SetNextStep(step pipeline.Runner) {
	st.nextStep = step
	st.hasNextStep = true
}

func (st *SplitterState) AddTask(tr *pipeline.TaskRef, callback chan *pipeline.TaskStatus) {
	tas := st.ts.Split(tr.Task, callback)

	if len(tas) == 0 {
		tr.SendStatusUpdate(st.id, errors.Errorf("TaskSplitter returned no tasks"), true)
		return
	}
	svt, newId := st.sw.AddTask()

	tStat := make([]*pipeline.TaskControl, len(tas))

	p := &pipeline.TaskRef{
		Key:    newId,
		Task:   tr.Task,
		Parent: tr,
	}

	for i, ta := range tas {
		tStat[i] = &pipeline.TaskControl{
			Status: pipeline.TaskStatus{
				Id: nil,
				Ref: &pipeline.TaskRef{
					Key:    pipeline.TaskIndex(i),
					Task:   ta.Task,
					Parent: p,
				},
				Result:   nil,
				Finished: false,
			},
			Cancel: ta.Cancel,
		}
	}

	*svt = pipeline.SuperVisorTask{
		Main:  tr,
		TStat: tStat,
	}

	for i, ta := range tas {
		ta.Runner.Push(tStat[i].Status.Ref)
	}
}

func (st *SplitterState) IsIdle() bool {
	return st.sw.IsEmpty()
}

// Returns true, if the window is empty
func (st *SplitterState) UpdateTask(sMsg *pipeline.TaskStatus) bool {

	svt := st.sw.GetSupervisorTask(sMsg.Ref.Parent.Key)
	tStat := svt.TStat.([]*pipeline.TaskStatus)
	tStat[sMsg.Ref.Key] = sMsg
	// TODO
	// How to pass on the result to the next Task?
	// Add to Context?

	// Idea: Should we move the whole TaskStatus handling into the splitter?
	if st.ts.IsFinished(tStat) {
		joinStat := st.ts.Join(tStat)
		if st.hasNextStep && protocol.IsACK(joinStat.Result){
			// If there is a next step, push the main task to it and send the result as a status update
			st.nextStep.Push(svt.Main)
			svt.Main.SendStatusUpdate(st.id, joinStat.Result, false)
		} else {
			svt.Main.SendStatusUpdate(st.id, joinStat.Result, true)
		}
		return st.sw.RemoveTask(sMsg.Ref.Parent.Key)
	}

	return false
}
