package pipeline

import (
	"github.com/pkg/errors"
	"sync"
)

//type ResultFunc func(t *Task, ts *TaskResult)

type Supervisor struct {
	ws       WindowState
	id       ElementId
	q        chan *TaskRef
	status   chan *TaskStatus
	stop     chan bool
	maxWSize uint32
	tg       TaskSplitter
}

func (s *Supervisor) SendStatusUpdate(ch chan *TaskStatus, ts *TaskStatus) {
	ts.Id = s.id
	ch <- ts
}

func (s *Supervisor) PushNextStep(tr *TaskRef) {
	panic("implement me")
}



const(
	defaultWS uint32 = 8
)

func (s *Supervisor) SetNextStep(runner Runner) {
//TODO
	panic("implement me")
}

func (s *Supervisor) Id() ElementId {
	return s.id
}

type TaskSplitter interface {
	Split(origin *Task, callback chan *TaskStatus) []*TaskAssignment
	// Add result aggregation?
	StartRunners(wg *sync.WaitGroup)
	StopRunners()
}

type TaskAssignment struct {
	Task   *Task
	Runner Runner
}

type SuperVisorTask struct {
	main  *TaskRef
	tStat []*TaskStatus
}

// The default WindowState implementation. It maintains a window and sets finishes a task, when all sub-tasks
// are finished.
type svWindowState struct {
	tBuffer []*SuperVisorTask
	wStart  uint32
	wEnd    uint32
	wsc      WindowStateCallback
	cond    *sync.Cond
	maxWSize uint32
}

func (s *Supervisor) Push(tr *TaskRef) {
	s.q <- tr
}

func (st *svWindowState) AddTask(tr *TaskRef, tas []*TaskAssignment) {
	st.cond.L.Lock()
	defer st.cond.L.Unlock()

	for st.wSize() >= st.maxWSize {
		st.cond.Wait()
	}

	//	tas := st.sv.tg.Split(tr.Task, st.sv.status)

	if len(tas) == 0 {
		st.wsc.SendStatusUpdate(tr.Task.Callback,&TaskStatus{
			Ref: tr,
			Result: TaskResult{
				Ack: Failed,
				Err: errors.Errorf("TaskSplitter returned no tasks"),
			},
			Finished: true,
		})
	}
	newId := TaskId(st.wEnd)

	tStat := make([]*TaskStatus, len(tas))
	p := &TaskRef{
		Key:    newId,
		Task:   tr.Task,
		Parent: tr,
	}

	for i, ta := range tas {
		tStat[i] = &TaskStatus{
			Ref: &TaskRef{
				Key:    TaskId(i),
				Task:   ta.Task,
				Parent: p,
			},
			Result: initialStatus(),
		}
	}

	svt := &SuperVisorTask{
		main:  tr,
		tStat: tStat,
	}

	st.tBuffer[newId] = svt

	for i, ta := range tas {
		ta.Runner.Push(tStat[i].Ref)
	}

	if st.wEnd < st.maxWSize-1 {
		st.wEnd++
	} else {
		st.wEnd = 0
	}
}

// This is not thread safe. The caller needs to lock svWindowState.cond.L
func (st *svWindowState) RemoveTask(id TaskId) {
	i := uint32(id)
	st.tBuffer[i] = nil

	for i := st.wStart; st.tBuffer[st.wStart] == nil && i < st.maxWSize && st.wStart != st.wEnd; i++ {
		if st.wStart < st.maxWSize-1 {
			st.wStart++
		} else {
			st.wStart = 0
		}
	}
}

// Returns true, if the window is empty
func (st *svWindowState) UpdateTask(sMsg *TaskStatus) bool {
	sendStatus := true

	svt := st.tBuffer[sMsg.Ref.Parent.Key]
	oldAck := svt.tStat[sMsg.Ref.Key].Result.Ack
	svt.tStat[sMsg.Ref.Key].Result = sMsg.Result
	svt.tStat[sMsg.Ref.Key].Finished = sMsg.Finished
	if oldAck != sMsg.Result.Ack {
		sendStatus = true
	}

// Add call to callback to start next step
	for _, ts := range svt.tStat {
		if !ts.Finished {
			if sendStatus {
// The task is not finished, but the Ack result has changed. Therefore send a status update
				st.wsc.SendStatusUpdate(	svt.main.Task.Callback ,&TaskStatus{
					Ref:    svt.main,
					Result: st.aggregateResult(svt.tStat),
					Finished:false,
				})
			}
			return false // There is still work to do
		}
	}
	st.wsc.SendStatusUpdate(	svt.main.Task.Callback ,&TaskStatus{
		Ref:    svt.main,
		Result: st.aggregateResult(svt.tStat),
		Finished:true,
	})

	st.wsc.PushNextStep(svt.main)
	defer st.cond.L.Unlock()
	st.cond.L.Lock()

	// All Workers finished
	st.RemoveTask(sMsg.Ref.Parent.Key)
	if st.wSize() == 0 {
		return true
	}
	st.cond.Signal()
	return false
}

func (st *svWindowState) aggregateResult(ts []*TaskStatus) TaskResult {
	agg := TaskResult{
		Ack: Pending,
		Err: nil,
	}

	pending, stored, completed, failed, retry := 0, 0, 0, 0, 0

	for _, rs := range ts {
		switch rs.Result.Ack {
		case Pending:
			pending++
		case Stored:
			stored++
		case Completed:
			completed++
		case Retry:
			retry++
		case Failed:
			failed++
		}
	}

	// And now the tough decisions

	// If anything failed, the parent task failed
	if failed > 0 {
		agg.Ack = Failed
		agg.Err = errors.New("At least one sub action failed.")
		return agg
	}

	if stored + completed == len(ts){
		if stored == 0{
			agg.Ack = Completed
		}else {
			agg.Ack = Stored
		}
	} else if retry == len(ts){
		agg.Ack = Retry
	}
//TODO What about partial Retry?
	return agg
}

// This is not thread safe. The caller needs to lock svWindowState.cond.L
func (st *svWindowState) wSize() uint32 {
	if st.wStart <= st.wEnd {
		return st.wEnd - st.wStart + 1
	}
	return (st.maxWSize - st.wStart) + st.wEnd + 1
}

func initialStatus() TaskResult {
	return TaskResult{
		Ack: Pending,
		Err: nil,
	}
}

func (s *Supervisor) Start(wg *sync.WaitGroup) {
	s.tg.StartRunners(wg)

	stopComplete := make(chan bool, 1)
	wg.Add(2)
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
				empty = s.ws.UpdateTask(ts)
			case <-stopComplete:
				stopped = true
			}

		}
	}()

	go func() {
		defer wg.Done()
		for {
			tr, more := <-s.q
			if !more {
				stopComplete <- true
				break
			}

			s.ws.AddTask(tr, s.tg.Split(tr.Task,s.status))

		}
	}()
}

func (s *Supervisor) Stop() {
	s.tg.StopRunners()
}

func NewSupervisor(tg TaskSplitter,id ElementId, opts ...StateOption) *Supervisor {


	s := &Supervisor{
		q:        make(chan *TaskRef, defaultWS),
		status:   make(chan *TaskStatus, defaultWS*8), //8 is just a wild guess right now
		stop:     nil,
		tg:       tg,
		ws: 	  nil,
		id:		  id,
	}
	for _,o := range opts{
		o(s)
	}

// A StateOption could have created its own WindowState implementation.
	if s.ws == nil{
		s.ws = &svWindowState{
			tBuffer: make([]*SuperVisorTask, s.maxWSize),
			wStart:  0,
			wEnd:    0,
			wsc:     s,
			cond:    sync.NewCond(&sync.Mutex{}),
			maxWSize: defaultWS,
		}
	}
	return s
}

type WindowState interface {
	AddTask(tr *TaskRef, tas []*TaskAssignment)
	RemoveTask(id TaskId)
	UpdateTask(sMsg *TaskStatus) bool
}

type WindowStateCallback interface{
	PushNextStep(tr *TaskRef)
	SendStatusUpdate(ch chan *TaskStatus, s *TaskStatus)
}
