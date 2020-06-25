package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"sync"
)


type ProcessorOutput struct {
	Result   TaskResult
	FollowUp context.Context
	Changes  []binding.Transformer
}

type Processor interface {
	Process(*Task) *ProcessorOutput
}

type Runner interface {
	Element
	Push(*TaskContainer)
}

type Worker struct {
	id    ElementId
	q     chan *TaskContainer
	stop  chan bool
	p     Processor
	nStep Runner
}

func (w *Worker) Id() ElementId {
	return w.id
}

func NewWorker(p Processor, id ElementId) *Worker {
	return &Worker{
		id:   id,
		q:    make(chan *TaskContainer, defaultWS),
		stop: make(chan bool, 1),
		p:    p,
	}
}

func (w *Worker) Push(tr *TaskContainer) {
	w.q <- tr
}

func (w *Worker) SetNextStep(runner Runner) {
	w.nStep = runner
}

func (w *Worker) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-w.stop:
				return
			case tc, more := <-w.q:
				if !more {
					return
				} else {
					// We could potentially add more sophisticated things like retry handling here
					func() {
						pOut := w.p.Process(&tc.Task)
						f := true
						if w.nStep != nil && protocol.IsACK(pOut.Result) {
							f = false
							w.nStep.Push(tc.FollowUp(pOut))
						}
						tc.SendStatusUpdate(w.id,pOut.Result,f)
					}()
				}
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.stop <- true
}


