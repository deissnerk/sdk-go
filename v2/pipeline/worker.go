package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"sync"
)


type ProcessorOutput struct {
	Result   TaskResult
	FollowUp context.Context
}

type Processor interface {
	Process(*TaskRef) ProcessorOutput
}

type Runner interface {
	Element
	Push(*TaskRef)
}

type Worker struct {
	id    ElementId
	q     chan *TaskRef
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
		q:    make(chan *TaskRef, defaultWS),
		stop: make(chan bool, 1),
		p:    p,
	}
}

func (w *Worker) Push(tr *TaskRef) {
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
			case tr, more := <-w.q:
				if !more {
					return
				} else {
					// We could potentially add more sophisticated things like retry handling here
					func() {
						pOut := w.p.Process(tr)
						f := true
						if w.nStep != nil && protocol.IsACK(pOut.Result) {
							f = false
							if pOut.FollowUp != nil {
								tr.Task.Context = pOut.FollowUp
							}
							w.nStep.Push(tr)
						}
						tr.Task.Callback <- &TaskStatus{
							Id:	      w.id,
							Ref:      tr,
							Result:   pOut.Result,
							Finished: f,
						}
					}()
				}
			}
		}
	}()
}

func (w *Worker) Stop() {
	w.stop <- true
}


