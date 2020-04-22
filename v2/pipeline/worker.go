package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"sync"
)


type Processor interface {
	Process(*TaskRef) TaskResult
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
// When do we need Cancel()?
						defer tr.Task.Cancel()
						tRes := w.p.Process(tr)
						f := true
						if w.nStep != nil && protocol.IsACK(tRes) {
							f = false
							w.nStep.Push(tr)
						}
						tr.Task.Callback <- &TaskStatus{
							Id:	      w.id,
							Ref:      tr,
							Result:   tRes,
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

func NewTask(t *Task, status chan *TaskStatus) *Task {
	nt := *t
	nt.Context, nt.Cancel = context.WithCancel(t.Context)
	nt.Callback = status
	return &nt
}
