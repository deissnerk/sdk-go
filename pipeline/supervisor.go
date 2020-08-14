package pipeline

import (
	"context"
	"sync"
)

type Supervisor struct {
	state  SupervisorState
	id     ElementId
	q      chan *TaskContainer
	status chan *StatusMessage
}

const (
	DefaultWS TaskIndex = 8
)

type TaskAssignment struct {
	Task   *Task
	Runner Runner
	Cancel context.CancelFunc
}

type SuperVisorTask struct {
	Main *TaskContainer
	//	TStat []*TaskStatus
	TStat interface{}
}

func (s *Supervisor) Push(tc *TaskContainer) {
	select {
	case s.q <- tc:
		return
	case <-tc.Task.Context.Done():
		tc.SendCancelledUpdate()
	}
}

func (s *Supervisor) Start(wg *sync.WaitGroup) {
	s.state.Start(wg)
	s.q = make(chan *TaskContainer, DefaultWS)
	s.status = make(chan *StatusMessage, DefaultWS*8) //8 is just a wild guess right now

	stopComplete := make(chan bool, 1)
	wg.Add(2)

	go func() {
		defer close(stopComplete)
		defer wg.Done()
		defer close(s.status)
		defer s.state.Stop()

		empty := false
		stopped := false
		for {
			if empty && stopped {
				// No incoming events and state is idle
				return
			}
			select {
			case sm := <-s.status:
				empty = s.state.UpdateTask(sm)
			case <-stopComplete:
				stopped = true
				empty = s.state.IsIdle()
			}

		}
	}()

	go func() {
		defer wg.Done()
		for tc := range s.q {
				s.state.AddTask(tc, s.status)
		}
		stopComplete <- true

	}()
}

func (s *Supervisor) Stop() {

	close(s.q)
}

func NewSupervisor(state SupervisorState) *Supervisor {
	s := &Supervisor{
		state: state,
	}
	return s
}

type SupervisorState interface {
	AddTask(tc *TaskContainer, callback chan *StatusMessage)
	UpdateTask(sMsg *StatusMessage) bool
	IsIdle() bool
	StartStop
}
