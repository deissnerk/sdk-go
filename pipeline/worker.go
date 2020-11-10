package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
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
// ProcessorFunc is a type alias to implement a Processor through a function pointer
type ProcessorFunc func(*Task) *ProcessorOutput

func (pf ProcessorFunc) Process(t *Task) *ProcessorOutput {
	return pf(t)
}

type ProcessorConstructor func() (Processor,error)
type ProcessorFuncConstructor func() (ProcessorFunc, error)

type Worker struct {
	id    ElementId
	q     chan *TaskContainer
	p     Processor
	nStep Runner
	stopped chan struct{}
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
		stopped: make(chan struct{}),
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

func (w *Worker) Start() error {

	go func() {
		defer close(w.stopped)

		for tc := range w.q {
			// We could potentially add more sophisticated things like retry handling here
			func() {
				pOut := w.p.Process(&tc.Task)
				if pOut == nil {
					pOut = &ProcessorOutput{}
				}
				if w.nStep != nil && protocol.IsACK(pOut.Result.Error) {
					tc.SendStatusUpdate(w.id, pOut.Result, false)
					tc.AddOutput(pOut)
					w.nStep.Push(tc)
				}else {
					tc.SendStatusUpdate(w.id, pOut.Result, true)
				}
			}()
		}
		//if w.nStep != nil {
		//	w.nStep.Stop()
		//}
	}()

	return nil
}

func (w *Worker) Stop(ctx context.Context) {
	close(w.q)
	select {
	case <-ctx.Done():


	case <-w.stopped:
	}
}
