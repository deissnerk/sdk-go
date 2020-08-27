package elements

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/binding"
	pipeline2 "github.com/cloudevents/sdk-go/pipeline"
	"sync"
)

type ReceiveHandler interface {
	// HandleResult() is called to handle any status updates regarding the event that
	// was fed into the pipeline. If true is returned, the processing of the event has been
	// completed.
	HandleResult(event binding.Message, ts *pipeline2.TaskControl) bool

	// Receive works similar to the Receiver() method of protocol.Receiver,
	// but it should handle protocol specific errors.
	// E.g. in AMQP a connection may be closed from time to time and just has
	// to be reopened again. These protocol specific errors that can be handled
	// graciously SHOULD NOT be reported back to the Inbound
	Receive(ctx context.Context) (*pipeline2.Task, error)
}

func NewInbound(handler ReceiveHandler,
	parentCtx context.Context,
	id pipeline2.ElementId,
	firstStep pipeline2.Runner) *Inbound {

	rcvCtx, rcvCancelFn := context.WithCancel(parentCtx)
	i := &Inbound{
		startLock: sync.Mutex{},
		started:   false,
		id:        id,
		sv: pipeline2.NewSupervisor(
			&InboundState{
				id:        id,
				sw:        pipeline2.NewSlidingWindow(pipeline2.DefaultWS),
				firstStep: firstStep,
				rh:        handler,
			}),
		rcvCtx:      rcvCtx,
		rcvCancelFn: rcvCancelFn,
		rcvHandler:  handler,
		stopped:     make(chan struct{}),
	}

	return i
}

type Inbound struct {
	startLock sync.Mutex
	started   bool
	//	state       *InboundState
	id          pipeline2.ElementId
	sv          *pipeline2.Supervisor
	rcvCtx      context.Context
	rcvCancelFn context.CancelFunc
	rcvHandler  ReceiveHandler
	stopped     chan struct{}
}

func (i *Inbound) Id() pipeline2.ElementId {
	return i.id
}

func (i *Inbound) Start() error {
	i.startLock.Lock()
	defer i.startLock.Unlock()
	if i.started == true {
		_, e := fmt.Printf("Already started")
		return e
	}

	i.sv.Start()

	go func(ctx context.Context) {
		defer close(i.stopped)
		for {

			if task, err := i.rcvHandler.Receive(ctx); err != nil {
				// Log the error? Perhaps the ReceiveHandler should do that!

				// Error handling may be difficult. When to retry and when to cancel?
				// Encapsulate this in ReceiveHandler?
			} else {
				i.sv.Push(&pipeline2.TaskContainer{
					Callback: nil,
					Key:      nil,
					Task:     *task,
					Parent:   nil,
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

func (i *Inbound) Stop(ctx context.Context) {
	i.startLock.Lock()
	defer i.startLock.Unlock()
	if i.started {
		i.rcvCancelFn()
		select {
		case <- i.stopped:
			case <- ctx.Done():
		}
		i.sv.Stop(ctx)
		i.started = false
	}
}

type InboundState struct {
	id        pipeline2.ElementId
	sw        *pipeline2.SlidingWindow
	firstStep pipeline2.Runner
	rh        ReceiveHandler
}

func (iState *InboundState) AddTask(tc *pipeline2.TaskContainer, callback chan *pipeline2.StatusMessage) {
	svt, key := iState.sw.AddTask()
	var cancel context.CancelFunc

	// Add cancel to the context
	tc.Task.Context, cancel = context.WithCancel(tc.Task.Context)
	tc.Key = key
	tc.Callback = callback

	*svt = pipeline2.SuperVisorTask{
		Main: tc,
		TStat: pipeline2.TaskControl{
			Status: pipeline2.TaskStatus{
				Id:       iState.id,
				Result:   nil,
				Finished: false,
			},
			Cancel: cancel,
		},
	}

	iState.firstStep.Push(tc)

}

func (iState *InboundState) UpdateTask(sMsg *pipeline2.StatusMessage) bool {
	tKey := sMsg.Key.(pipeline2.TaskIndex)
	svt := iState.sw.GetSupervisorTask(tKey)
	tStat := svt.TStat.(pipeline2.TaskControl)
	tStat.Status = sMsg.Status
	if iState.rh.HandleResult(svt.Main.Task.Event, &tStat) {
		return iState.sw.RemoveTask(tKey)
	}
	return false
}

func (iState *InboundState) IsIdle() bool {
	return iState.sw.IsEmpty()
}

func (iState *InboundState) Start() error {
	return nil
}

func (iState *InboundState) Stop(ctx context.Context) {

	//	iState.firstStep.Stop()
}
