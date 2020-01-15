package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/pkg/pipeline/config"
	"sync"
)

type TaskId uint32

//type CallbackFunc func(s *TaskStatus)

type Task struct {
	Context  context.Context
	Cancel   context.CancelFunc
//	Event    *cloudevents.Event
	Event 	 binding.Message
	Callback chan *TaskStatus
	Msg      interface{}
}

type TaskRef struct {
	Key    TaskId
	Task   *Task
	Parent *TaskRef
}

type TaskAck int

const (
	Pending TaskAck = iota
	Stored
	Completed
	Failed
	Retry
)

type TaskResult struct {
	Ack TaskAck
	Err error
}

type TaskStatus struct {
	Id       ElementId
	Ref      *TaskRef
	Result   TaskResult
	Finished bool
}

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

func NewWorker(p Processor, cfg *config.Config, id ElementId) *Worker {
	return &Worker{
		id:   id,
		q:    make(chan *TaskRef, cfg.WSize),
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
						if w.nStep != nil &&
							tRes.Ack != Failed &&
							tRes.Ack != Retry {
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
