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
	defaultWS TaskIndex = 8
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

func (s *Supervisor) Push(tr *TaskContainer) {
	s.q <- tr
}

func (s *Supervisor) Start(wg *sync.WaitGroup) {
	s.state.Start(wg)
	s.q = make(chan *TaskContainer, defaultWS)
	s.status = make(chan *StatusMessage, defaultWS*8) //8 is just a wild guess right now

	stopComplete := make(chan bool, 1)
	wg.Add(2)

	// Can this goroutine be the same as in Inbound?
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
				break
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
	//TODO Create channels when starting and close them when stopping
	close(s.q)
}

func (s *Supervisor) initChannels() {
}

func NewSupervisor(state SupervisorState) *Supervisor {
	s := &Supervisor{
		state:  state,
	}
	return s
}

type SupervisorState interface {
	AddTask(tr *TaskContainer, callback chan *StatusMessage)
	UpdateTask(sMsg *StatusMessage) bool
	SetNextStep(runner Runner)
	IsIdle() bool
	StartStop
}
