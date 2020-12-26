package elements

import (
	"context"
	"fmt"
	pipeline2 "github.com/cloudevents/sdk-go/pipeline"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"sync"
)

type ResultHandler interface {
	// HandleResult() is called to handle any status updates regarding the event that
	// was fed into the pipeline. If true is returned, the processing of the event has been
	// completed.
	HandleResult(msg binding.Message, ts *pipeline2.TaskStatus) bool
}

// The ReceiverHandler is also a protocol.Receiver, but handling of communication errors
// that require protocol specific logic or retries should be done inside Receive(). The Inbound
// will stop inbound processing on error.
type Receiver protocol.Receiver

func NewInbound(handler ResultHandler,
	receiver Receiver,
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
		receiver:    receiver,
		handler:     handler,
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
	receiver    Receiver
	handler     ResultHandler

	stopped chan struct{}
}

func (i *Inbound) Push(tc *pipeline2.TaskContainer) {
	i.sv.Push(tc)
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

	if i.receiver != nil {
		go func(ctx context.Context) {
			defer close(i.stopped)
			for {

				if msg, err := i.receiver.Receive(ctx); err != nil {
					// Log the error? Perhaps the ReceiveHandler should do that!

					// Error handling may be difficult. When to retry and when to cancel?
					// Encapsulate this in ReceiveHandler?
				} else {
					task, err := pipeline2.NewAccessMetadataTask(context.TODO(), msg)
					if err != nil {
						i.handler.HandleResult(msg,
							&pipeline2.TaskStatus{
								Result:   pipeline2.TaskResult{Result: err},
								Finished: true,
								Id:       i.id})
					}

					i.sv.Push(pipeline2.NewRootContainer(task))
				}
				select {
				case <-ctx.Done():
					return
				default:

				}

			}
		}(i.rcvCtx)
	}
	i.started = true
	return nil

}

func (i *Inbound) Stop(ctx context.Context) {
	i.startLock.Lock()
	defer i.startLock.Unlock()
	if i.started {
		if i.receiver != nil {
			i.rcvCancelFn()
			select {
			case <-i.stopped:
			case <-ctx.Done():
			}
		}
		i.sv.Stop(ctx)
		i.started = false
	}
}

type InboundState struct {
	id        pipeline2.ElementId
	sw        *pipeline2.SlidingWindow
	firstStep pipeline2.Runner
	rh        ResultHandler
}

func (iState *InboundState) AddTask(tc *pipeline2.TaskContainer, callback chan *pipeline2.StatusMessage) {
	svt, key := iState.sw.AddTask()

	*svt = pipeline2.SuperVisorTask{
		Main: tc,
		TStat: &pipeline2.TaskStatus{
			Id:       iState.id,
			Result:   pipeline2.TaskResult{},
			Finished: false,
		},
	}

	iState.firstStep.Push(tc.NewChild(callback, key, tc.Task))

}

func (iState *InboundState) UpdateTask(sMsg *pipeline2.StatusMessage) bool {
	tKey := sMsg.Key.(pipeline2.TaskIndex)
	svt := iState.sw.GetSupervisorTask(tKey)
	tStat := svt.TStat.(*pipeline2.TaskStatus)
	tStat = &sMsg.Status
	if iState.rh.HandleResult(svt.Main.GetWrappedMessage(), tStat) {
		// HandleResult() takes care of the result at the very beginning of the Task. No more actions
		// needed after that.
		return iState.sw.RemoveTask(tKey, func() {})
	}
	return false
}

func (iState *InboundState) IsIdle() bool {
	return iState.sw.IsEmpty()
}

func (iState *InboundState) Start() error {
	if startStop, ok := iState.rh.(pipeline2.StartStop); ok {
		return startStop.Start()
	}
	return nil
}

func (iState *InboundState) Stop(ctx context.Context) {
	if startStop, ok := iState.rh.(pipeline2.StartStop); ok {
		startStop.Stop(ctx)
	}
}
