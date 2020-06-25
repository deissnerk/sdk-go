package pipeline

import (
	"context"
	v2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"math"
	"pack.ag/amqp"
	"sync"
	"time"
)

type InboundHandler interface {
	HandleResult(ts *TaskStatus) error
	Receive(ctx context.Context) (binding.Message,error)
	Close(ctx context.Context) (err error)
}

// We might create a more general InboundProcessor interface later and make this an
// AMQP specific implementation of it.
type Inbound struct {
	stop      chan bool
	id        ElementId
	ws        uint32
	counter   uint32
	results   chan *TaskStatus
	runner    Runner
	errc      chan error
	ih        InboundHandler
	createCtx func(parent context.Context) (context.Context,context.CancelFunc)
}

func (i *Inbound) Id() ElementId {
	return i.id
}

func (i *Inbound) SetNextStep(runner Runner) {
	i.runner = runner
}

func (i *Inbound) Stop() {
	panic("implement me")
}

func (i *Inbound) SetFollowUp(runner Runner) {
	i.runner = runner
}

func (i *Inbound) Start(wg *sync.WaitGroup) {
	i.results = make(chan *TaskStatus, i.ws)
	wg.Add(2)
	go func() {
		defer wg.Done()
		for sMsg := range i.results {
			_ = i.ih.HandleResult(sMsg)
		}
	}()

	go func() {
		defer wg.Done()
		defer func() {
			if closeCtx, err := context.WithTimeout(i.ctx, time.Second*10); err != nil {
				_ = i.ih.Close(closeCtx)
			}
		}()

		for {
			ev,err := i.ih.Receive(i.ctx)
			if err != nil {
				i.errc <- err
				return
			} else {
				t := NewTask(ev,i.results,
					func()(context.Context,context.CancelFunc) {return i.createCtx(i.ctx)})
				i.runner.Push(&TaskContainer{
					Key:    i.nextId(),
					Task:   t,
					Parent: nil,
				})
			}


		}
	}()
}

func (i *Inbound) nextId() TaskIndex {
	if i.counter == math.MaxUint32 {
		i.counter = 0
	} else {
		i.counter++
	}
	return TaskIndex(i.counter)
}

//func NewInbound(ctx context.Context, c *config.Config, src string, opts []amqp.LinkOption) *Inbound {
func NewInbound(id ElementId,
	ctx context.Context,
	ih InboundHandler,
	r Runner,
	createCtx func(context.Context)(context.Context,context.CancelFunc)) *Inbound {

	i := &Inbound{
		id:        id,
		ws:        0,
		counter:   0,
		results:   nil,
		runner:    r,
		errc:      make(chan error, 1),
		ih:        ih,
		stop:       make (chan bool, 1),
		createCtx: createCtx,
	}

	//r := &Inbound{
	//	conf: c,
	//	errc : make(chan error, 1),
	//	ctx:ctx,
	//}
	//
	//o := append(opts, amqp.LinkSourceAddress(src))
	//if l, err := c.Session.NewReceiver(o...); err == nil {
	//	r.link = l
	//}
	//
	//return r
	return i
}

type ReceiveHandler struct {
	rcv 	protocol.Receiver
}

func NewReceiveHandler(rcv protocol.Receiver) *ReceiveHandler {
	return &ReceiveHandler{
		rcv:rcv,
	}
}

func (ah *ReceiveHandler) Close(ctx context.Context) (err error) {
	if rcvCloser,ok := ah.rcv.(protocol.ReceiveCloser); ok{
		return rcvCloser.Close(ctx)
	}
	
	// Nothing has to be closed
	return nil
}

func (ah *ReceiveHandler) HandleResult(ts *TaskStatus) error {
	m := ts.Ref.Task.Msg.(*amqp.Message)
    if v2.IsNACK(ts.Result) {
		_ = m.Reject(&amqp.Error{
			Condition:   amqp.ErrorInternalError,
			Description: ts.Result.Error(),
			Info:        nil,
		})
	} else if ts.Finished {
		m.Accept()
	}

	return ts.Result
}

func (ah *ReceiveHandler) ReceiveTask(rcvCtx context.Context) (binding.Message, error) {
	return ah.rcv.Receive(rcvCtx)
}




