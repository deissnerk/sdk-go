package elements

import (
	"context"
	"github.com/cloudevents/sdk-go/pipeline"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/pkg/errors"
)

/*type splitterTaskKey struct {
	key    pipeline.TaskIndex
	subKey pipeline.TaskIndex
}

type splitterTaskStatus struct {
	j Joiner
	tc []*pipeline.TaskControl
}*/

type Router interface {
	Route(origin *pipeline.Task) *TaskAssignment

	pipeline.StartStop
}

type RouterConstructor func() (Router, error)

type RouterRunner struct {
	state *RouterState
	sv    *pipeline.Supervisor
}

func (s *RouterRunner) Push(tc *pipeline.TaskContainer) {
	s.sv.Push(tc)
}

func NewRouterRunner(r Router, id pipeline.ElementId, nextStep pipeline.Runner) (*RouterRunner, error) {
	state := &RouterState{
		id: id,

		sw:       pipeline.NewSlidingWindow(pipeline.DefaultWS),
		nextStep: nextStep,
	}

	return &RouterRunner{
		state: state,
		sv:    pipeline.NewSupervisor(state),
	}, nil
}

func (s RouterRunner) Start() error {
	// The Supervisor will start the state and all related resources
	return s.sv.Start()
}

func (s RouterRunner) Stop(ctx context.Context) {
	s.sv.Stop(ctx)
}

func (s RouterRunner) Id() pipeline.ElementId {
	return s.state.id
}

var _ pipeline.Runner = (*RouterRunner)(nil)

//func (s SplitterRunner) SetNextStep(runner pipeline.Runner) {
//	s.state.SetNextStep(runner)
//}

// The default SupervisorState implementation. It maintains a window and sets finishes a task, when all sub-tasks
// are finished.
type RouterState struct {
	id pipeline.ElementId
	ts Router
	//	maxWSize pipeline.TaskIndex
	//	wsc      WindowStateCallback
	sw       *pipeline.SlidingWindow
	nextStep pipeline.Runner
}

func (st *RouterState) Id() pipeline.ElementId {
	return st.id
}

func (st *RouterState) Start() error {
	return st.ts.Start()
}

// Stop() calls Stop() on the Splitter and for the next step to propagate the Stop()
// call through the pipeline. The state is only stopped after the Supervisor was stopped.
func (st *RouterState) Stop(ctx context.Context) {
	st.ts.Stop(ctx)
	//	st.nextStep.Stop()
}

func (st *RouterState) AddTask(tc *pipeline.TaskContainer, callback chan *pipeline.StatusMessage) {
	ta := st.ts.Route(tc.Task)

	if ta == nil {
		tc.SendStatusUpdate(st.id, pipeline.TaskResult{
			Error:  errors.Errorf("Router returned no route"),
			Result: nil,
		}, true)
		return
	}
	svt, key := st.sw.AddTask()

	*svt = pipeline.SuperVisorTask{
		Main: tc,
		TStat: &pipeline.TaskStatus{
			Finished: false,
		},
	}

	ta.Runner.Push(tc.NewChild(callback,key,ta.Task))
}

func (st *RouterState) IsIdle() bool {
	return st.sw.IsEmpty()
}

// Returns true, if the window is empty
func (st *RouterState) UpdateTask(sMsg *pipeline.StatusMessage) bool {
	tKey := sMsg.Key.(pipeline.TaskIndex)
	svt := st.sw.GetSupervisorTask(tKey)
	tStat := svt.TStat.(*pipeline.TaskStatus)
	tStat = &sMsg.Status

	if tStat.Finished {
		if st.nextStep != nil && protocol.IsACK(tStat.Result.Error) {
			// If there is a next step, push the main task to it and send the result as a status update
			svt.Main.SendStatusUpdate(st.id, tStat.Result, false)
			svt.Main.AddOutput(&pipeline.ProcessorOutput{
				Result:   sMsg.Status.Result,
				FollowUp: nil,
				Changes:  nil,
			})
			st.nextStep.Push(svt.Main)
		} else {
			svt.Main.SendStatusUpdate(st.id, joinStat.Result, true)
		}
		return st.sw.RemoveTask(tKey.key)
	}

	return false
}

var _ pipeline.SupervisorState = (*SplitterState)(nil)
