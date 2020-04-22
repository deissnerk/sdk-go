package elements

import (
	"context"
	v2 "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/pipeline"

	ceamqp "github.com/cloudevents/sdk-go/v2/protocol/amqp"

	"math"
	"pack.ag/amqp"
	"sync"
	"time"
)

type InboundHandler interface {
	HandleResult(ts *pipeline.TaskStatus) error
	ReceiveTask(ctx context.Context) (*pipeline.TaskRef,error)
}

// We might create a more general InboundProcessor interface later and make this an
// AMQP specific implementation of it.
type Inbound struct {
	id      pipeline.ElementId
	ws      uint32
	counter uint32
	results chan *pipeline.TaskStatus
	runner  pipeline.Runner
	errc    chan error
	ih 		InboundHandler
	ctx		context.Context
}

func (i *Inbound) Id() pipeline.ElementId {
	return i.id
}

func (i *Inbound) SetNextStep(runner pipeline.Runner) {
	i.runner = runner
}

func (i *Inbound) Stop() {
	panic("implement me")
}

func (i *Inbound) SetFollowUp(runner pipeline.Runner) {
	i.runner = runner
}

func (i *Inbound) Start(wg *sync.WaitGroup) {
	i.results = make(chan *pipeline.TaskStatus, i.ws)
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
				_ = i.link.Close(closeCtx)
			}
		}()

		for {
			i.ih.ReceiveTask()

			t := &pipeline.Task{
				Context:  subCtx,
				Cancel:   c,
				Event:    aMsg,
				Callback: i.results,
				Msg:      m,
			}

			i.runner.Push(&pipeline.TaskRef{
				Key:    i.nextId(),
				Task:   t,
				Parent: nil,
			})

		}
	}()

}

func (i *Inbound) nextId() pipeline.TaskId {
	if i.counter == math.MaxUint32 {
		i.counter = 0
	} else {
		i.counter++
	}
	return pipeline.TaskId(i.counter)
}

//func NewInbound(ctx context.Context, c *config.Config, src string, opts []amqp.LinkOption) *Inbound {
func NewInbound(ih InboundHandler, r pipeline.Runner) *Inbound {



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
}

type AMQPHandler struct {
	link    *amqp.Receiver
	results chan *pipeline.TaskStatus
	errc    chan error

}

func (ah *AMQPHandler) HandleResult(ts *pipeline.TaskStatus) error {
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

func (ah *AMQPHandler) ReceiveTask(ctx context.Context) (*pipeline.TaskRef, error) {
	m, err := ah.link.Receive(i.ctx)

	if err != nil {
		i.errc <- err
		return
	} else {

		aMsg := &ceamqp.Message{AMQP:m}

		subCtx, c := context.WithTimeout(i.ctx, time.Second*1000)
	}
}




