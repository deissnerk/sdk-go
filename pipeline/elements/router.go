package elements

//import (
//	"context"
//	"github.com/cloudevents/sdk-go/pipeline"
//	"github.com/cloudevents/sdk-go/v2/protocol"
//	"github.com/pkg/errors"
//)
//
///*type splitterTaskKey struct {
//	key    pipeline.TaskIndex
//	subKey pipeline.TaskIndex
//}
//
//type splitterTaskStatus struct {
//	j Joiner
//	tc []*pipeline.TaskControl
//}*/
//
//type Router interface {
//	Route(origin *pipeline.Task) ([]*TaskAssignment,Joiner)
//
//	pipeline.StartStop
//}
//
//type RouterConstructor func() (Router,error)
//
//type RouterRunner struct {
//	state *SplitterState
//	sv    *pipeline.Supervisor
//}
//
//func (s *RouterRunner) Push(tc *pipeline.TaskContainer) {
//	s.sv.Push(tc)
//}
//
//func NewRouterRunner(r Router, id pipeline.ElementId, nextStep pipeline.Runner) (*SplitterRunner,error) {
//	state := &RouterState{
//		id:       id,
//
//		sw:       pipeline.NewSlidingWindow(pipeline.DefaultWS),
//		nextStep: nextStep,
//	}
//
//	return &SplitterRunner{
//		state: state,
//		sv:pipeline.NewSupervisor(state),
//	},nil
//}
//
//func (s RouterRunner) Start() error{
//	// The Supervisor will start the state and all related resources
//	return s.sv.Start()
//}
//
//func (s RouterRunner) Stop(ctx context.Context) {
//	s.sv.Stop(ctx)
//}
//
//func (s RouterRunner) Id() pipeline.ElementId {
//	return s.state.id
//}
//
//var _ pipeline.Runner = (*RouterRunner)(nil)
//
////func (s SplitterRunner) SetNextStep(runner pipeline.Runner) {
////	s.state.SetNextStep(runner)
////}
//
//// The default SupervisorState implementation. It maintains a window and sets finishes a task, when all sub-tasks
//// are finished.
//type RouterState struct {
//	id pipeline.ElementId
//	ts Router
//	//	maxWSize pipeline.TaskIndex
//	//	wsc      WindowStateCallback
//	sw       *pipeline.SlidingWindow
//	nextStep pipeline.Runner
//}
//
//func (st *RouterState) Id() pipeline.ElementId {
//	return st.id
//}
//
//func (st *RouterState) Start() error {
//	return st.ts.Start()
//}
//
//// Stop() calls Stop() on the Splitter and for the next step to propagate the Stop()
//// call through the pipeline. The state is only stopped after the Supervisor was stopped.
//func (st *RouterState) Stop(ctx context.Context) {
//	st.ts.Stop(ctx)
//	//	st.nextStep.Stop()
//}
//
////func (st *SplitterState) SetNextStep(step pipeline.Runner) {
////	st.nextStep = step
////}
//
//func (st *RouterState) AddTask(tc *pipeline.TaskContainer, callback chan *pipeline.StatusMessage) {
//	tas,j := st.ts.Split(&tc.Task)
//
//	if len(tas) == 0 {
//		tc.SendStatusUpdate(st.id, pipeline.TaskResult{
//			Error:  errors.Errorf("Splitter returned no tasks"),
//			Result: nil,
//		}, true)
//		return
//	}
//	svt, key := st.sw.AddTask()
//
//	stc := make([]*pipeline.TaskControl, len(tas))
//
//	for i, ta := range tas {
//		id := ta.Runner.Id()
//		stc[i] = &pipeline.TaskControl{
//			Status: pipeline.TaskStatus{
//				Id: id,
//				//Key: splitterTaskKey{
//				//	key:    key,
//				//	subKey: pipeline.TaskIndex(i),
//				//},
//				Finished: false,
//			},
//			Cancel: ta.Cancel,
//		}
//	}
//
//	*svt = pipeline.SuperVisorTask{
//		Main:  tc,
//		TStat: &splitterTaskStatus{
//			j:  j,
//			tc: stc,
//		},
//	}
//
//	for i, ta := range tas {
//		ta.Runner.Push(
//			&pipeline.TaskContainer{
//				Callback: callback,
//				Key: splitterTaskKey{
//					key:    key,
//					subKey: pipeline.TaskIndex(i),
//				},
//				Task:    *ta.Task,
//				Parent:  tc,
//			})
//	}
//}
//
//func (st *RouterState) IsIdle() bool {
//	return st.sw.IsEmpty()
//}
//
//// Returns true, if the window is empty
//func (st *RouterState) UpdateTask(sMsg *pipeline.StatusMessage) bool {
//	tKey := sMsg.Key.(splitterTaskKey)
//	svt := st.sw.GetSupervisorTask(tKey.key)
//	tStat := svt.TStat.(*splitterTaskStatus)
//	tStat.tc[tKey.subKey].Status = sMsg.Status
//
//	if tStat.j.IsFinished(tStat.tc) {
//		joinStat := tStat.j.Join(tStat.tc)
//		if st.nextStep != nil && protocol.IsACK(joinStat.Result.Error) {
//			// If there is a next step, push the main task to it and send the result as a status update
//			svt.Main.SendStatusUpdate(st.id, joinStat.Result, false)
//			svt.Main.AddOutput(joinStat)
//			st.nextStep.Push(svt.Main)
//		} else {
//			svt.Main.SendStatusUpdate(st.id, joinStat.Result, true)
//		}
//		return st.sw.RemoveTask(tKey.key)
//	}
//
//	return false
//}
//
//var _ pipeline.SupervisorState = (*SplitterState)(nil)
