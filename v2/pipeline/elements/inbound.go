package elements

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/pipeline"
	"sync"
)

type ReceiveHandler interface {

	pipeline.StartStop

	// HandleResult() is called to handle any status updates regarding the event that
	// was fed into the pipeline. If true is returned, the processing of the event has been
	// completed.
	HandleResult(event binding.Message, ts *pipeline.TaskControl) bool

	// Receive works similar to the Receiver() method of protocol.Receiver,
	// but it should handle protocol specific errors.
	// E.g. in AMQP a connection may be closed from time to time and just has
	// to be reopened again. These protocol specific errors that can be handled
	// graciously SHOULD NOT be reported back to the Inbound
	Receive(ctx context.Context) (*pipeline.Task, error)
}

func NewInbound(handler ReceiveHandler, firstStep pipeline.Runner, parentCtx context.Context,
	id pipeline.ElementId) *Inbound {
	rcvCtx,rcvCancelFn := context.WithCancel(parentCtx)
	i := &Inbound{
		startLock: sync.Mutex{},
		started:   false,
		state: &InboundState{
			id:        nil,
			sw:        nil,
			firstStep: nil,
			rh:        nil,
		},
		sv: pipeline.NewSupervisor(
			&InboundState{
				id:        id,
				sw:        pipeline.NewSlidingWindow(pipeline.DefaultWS),
				firstStep: firstStep,
				rh:        handler,
			}),
		rcvCtx:      rcvCtx,
		rcvCancelFn: rcvCancelFn,
		rcvHandler:  handler,
	}

	return i
}

type Inbound struct {
	startLock   sync.Mutex
	started     bool
	state       *InboundState
	sv          *pipeline.Supervisor
	rcvCtx      context.Context
	rcvCancelFn context.CancelFunc
	rcvHandler  ReceiveHandler
}

func (i *Inbound) Start(wg *sync.WaitGroup) error {
	i.startLock.Lock()
	defer i.startLock.Unlock()
	if i.started == true {
		_, e := fmt.Printf("Already started")
		return e
	}
	wg.Add(1)
	i.sv.Start(wg)

	go func(ctx context.Context) {
		defer wg.Done()
		for {

			if task, err := i.rcvHandler.Receive(ctx); err != nil {
				// TODO Log the error? Perhaps the ReceiveHandler should do that!
				i.Stop()
				// Error handling may be difficult. When to retry and when to cancel?
				// Encapsulate this in ReceiveHandler?
			} else {
				i.sv.Push(&pipeline.TaskContainer{
					Callback: nil,
					Key:      nil,
					Task:     *task,
					Parent:   nil,
					Changes:  nil,
				})
			}
			select {
			case <-ctx.Done():
				return
			default:

			}

		}
	}(i.rcvCtx)
	i.started = true
	return nil

}

func (i *Inbound) Stop() {
	i.rcvCancelFn()
	i.sv.Stop()
}

type InboundState struct {
	id        pipeline.ElementId
	sw        *pipeline.SlidingWindow
	firstStep pipeline.Runner
	rh        ReceiveHandler
}

func (iState *InboundState) AddTask(tc *pipeline.TaskContainer, callback chan *pipeline.StatusMessage) {
	svt, key := iState.sw.AddTask()
	var cancel context.CancelFunc

	// Add cancel to the context
	tc.Task.Context, cancel = context.WithCancel(tc.Task.Context)
	tc.Key = key

	*svt = pipeline.SuperVisorTask{
		Main: tc,
		TStat: pipeline.TaskControl{
			Status: pipeline.TaskStatus{
				Id:       iState.id,
				Result:   nil,
				Finished: false,
			},
			Cancel: cancel,
		},
	}

	iState.firstStep.Push(tc)

}

func (iState *InboundState) UpdateTask(sMsg *pipeline.StatusMessage) bool {
	tKey := sMsg.Key.(pipeline.TaskIndex)
	svt := iState.sw.GetSupervisorTask(tKey)
	tStat := svt.TStat.(pipeline.TaskControl)
	tStat.Status = sMsg.Status
	if iState.rh.HandleResult(svt.Main.Task.Event, &tStat) {
		return iState.sw.RemoveTask(tKey)
	}
	return false
}

func (iState *InboundState) IsIdle() bool {
	return iState.sw.IsEmpty()
}

func (iState *InboundState) Start(wg *sync.WaitGroup) error {
	return nil
}

func (iState *InboundState) Stop() {

	iState.firstStep.Stop()
}
