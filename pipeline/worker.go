package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"sync"
)

var _ Runner = (*Worker)(nil)

type ProcessorOutput struct {
	Result   TaskResult
	FollowUp context.Context
	Changes  []binding.Transformer
}

// TODO Needs to be closed/stopped?
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
	p     Processor
	nStep Runner
}

func (w *Worker) Id() ElementId {
	return w.id
}

func NewWorker(p Processor, id ElementId, nextStep Runner) *Worker {
	return &Worker{
		id:    id,
		q:     make(chan *TaskContainer, DefaultWS),
		p:     p,
		nStep: nextStep,
	}
}

func (w *Worker) Push(tc *TaskContainer) {
	select {
	case w.q <- tc:
		return
	case <-tc.Task.Context.Done():
		tc.SendCancelledUpdate()
	}
}

func (w *Worker) Then(runner Runner) Element {
	w.nStep = runner
	return runner
}

func (w *Worker) Start(wg *sync.WaitGroup) error {
	wg.Add(1)
	go func() {
		defer wg.Done()

		for tc := range w.q {
			// We could potentially add more sophisticated things like retry handling here
			func() {
				pOut := w.p.Process(&tc.Task)
				f := true
				if w.nStep != nil && protocol.IsACK(pOut.Result) {
					f = false
					w.nStep.Push(tc.FollowUp(pOut))
				}
				tc.SendStatusUpdate(w.id, pOut.Result, f)
			}()
		}
		if w.nStep != nil {
			w.nStep.Stop()
		}
	}()

	return nil
}

func (w *Worker) Stop() {
	close(w.q)
}
