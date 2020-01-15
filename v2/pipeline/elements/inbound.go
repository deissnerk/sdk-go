package elements

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/pipeline"
	"github.com/cloudevents/sdk-go/pkg/pipeline/config"


	ceamqp "github.com/cloudevents/sdk-go/pkg/bindings/amqp"

	"math"
	"pack.ag/amqp"
	"sync"
	"time"
)

type AMQPRouterSelector struct {
}

func (ars *AMQPRouterSelector) StartRunners() {
	panic("implement me")
}

func (ars *AMQPRouterSelector) StopRunners() {
	panic("implement me")
}

func (ars *AMQPRouterSelector) Split(origin *pipeline.Task, callback chan *pipeline.TaskStatus) []*pipeline.TaskAssignment {
	return nil
}

// We might create a more general InboundProcessor interface later and make this an
// AMQP specific implementation of it.
type Inbound struct {
	id      pipeline.PipelineElementId
	conf    *config.Config
	link    *amqp.Receiver
	ws      uint32
	counter uint32
	results chan *pipeline.TaskStatus
	runner  pipeline.Runner
	ctx     context.Context
	errc    chan error
}

func (i *Inbound) Id() pipeline.PipelineElementId {
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
			m := sMsg.Ref.Task.Msg.(*amqp.Message)
			switch sMsg.Result.Ack {
			case pipeline.Failed:
				_ = m.Reject(&amqp.Error{
					Condition:   amqp.ErrorInternalError,
					Description: sMsg.Result.Err.Error(),
					Info:        nil,
				})
			case pipeline.Stored:
				fallthrough
			case pipeline.Completed:
				m.Accept()
			case pipeline.Retry:
				m.Release()
			}
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
			m, err := i.link.Receive(i.ctx)

			if err != nil {
				i.errc <- err
				return
			} else {

				aMsg := &ceamqp.Message{AMQP:m}

				subCtx, c := context.WithTimeout(i.ctx, time.Second*1000)
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

func NewInbound(ctx context.Context, c *config.Config, src string, opts []amqp.LinkOption) *Inbound {

	r := &Inbound{
		conf: c,
		errc : make(chan error, 1),
		ctx:ctx,
	}

	o := append(opts, amqp.LinkSourceAddress(src))
	if l, err := c.Session.NewReceiver(o...); err == nil {
		r.link = l
	}

	return r
}
