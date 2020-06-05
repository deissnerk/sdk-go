package pipeline

import (
	"context"
	"sync"
)

//type ResultFunc func(t *Task, ts *ProcessorOutput)

type Supervisor struct {
	state  SupervisorState
	id     ElementId
	q      chan *TaskRef
	status chan *TaskStatus
	stop   chan bool
}

func (s *Supervisor) SendStatusUpdate(ch chan *TaskStatus, ts *TaskStatus) {
	ts.Id = s.id
	ch <- ts
}

const (
	defaultWS TaskIndex = 8
)

func (s *Supervisor) SetNextStep(runner Runner) {
	s.state.SetNextStep(runner)
}

func (s *Supervisor) Id() ElementId {
	return s.id
}


type TaskAssignment struct {
	Task   *Task
	Runner Runner
	Cancel context.CancelFunc
}

type SuperVisorTask struct {
	Main *TaskRef
	//	TStat []*TaskStatus
	TStat interface{}
}

func (s *Supervisor) Push(tr *TaskRef) {
	s.q <- tr
}

func (s *Supervisor) Start(wg *sync.WaitGroup) {
	s.state.Start(wg)

	stopComplete := make(chan bool, 1)
	wg.Add(2)

	// Can this goroutine be the same as in Inbound?
	go func() {
		defer close(stopComplete)
		defer wg.Done()
		empty := false
		stopped := false
		for {
			if empty && stopped {
				break
			}
			select {
			case <-s.stop:
				// Stop accepting new Tasks
				close(s.q)
				// Should we cancel all tasks in addition?

			case ts := <-s.status:
				empty = s.state.UpdateTask(ts)
			case <-stopComplete:
				stopped = true
				empty = s.state.IsIdle()
			}

		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			case tr, more := <-s.q:
				if !more {
					stopComplete <- true
					break
				}
				s.state.AddTask(tr, s.status)
			}
		}
	}()
}

func (s *Supervisor) Stop() {
	s.state.Stop()
}

func NewSupervisor(ctx context.Context, id ElementId, state SupervisorState) *Supervisor {
	s := &Supervisor{
		q:      make(chan *TaskRef, defaultWS),
		status: make(chan *TaskStatus, defaultWS*8), //8 is just a wild guess right now
		stop:   make(chan bool, 1),
		state:  state,
		id:     id,
	}
	//for _, o := range opts {
	//	o(s)
	//}
	//
	//// A StateOption could have created its own SupervisorState implementation.
	//if s.state == nil {
	//	s.state = &SplitterState{
	//		tBuffer:  make([]*SuperVisorTask, defaultWS),
	//		wStart:   0,
	//		wEnd:     0,
	//		wsc:      s,
	//		cond:     sync.NewCond(&sync.Mutex{}),
	//		maxWSize: defaultWS,
	//	}
	//}
	return s
}

// This interface describes how the Supervisor interacts with the state of sliding window
// An implementation, that waits for all sub-tasks to successfully complete, is used by default.
type SupervisorState interface {
	AddTask(tr *TaskRef, callback chan *TaskStatus)
	UpdateTask(sMsg *TaskStatus) bool
	SetNextStep(runner Runner)
	IsIdle() bool
	StartStop
}
