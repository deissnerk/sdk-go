package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type TaskIndex uint32

type Task struct {
	Context context.Context
	//	Event    *cloudevents.Event
	Event    binding.Message
	Callback chan *TaskStatus
}

// TODO
// Better term?
type TaskRef struct {
	Key    TaskIndex
	Task   *Task
	Parent *TaskRef
}

func (t *TaskRef) SendStatusUpdate(id ElementId, r TaskResult, finished bool) {
	t.Task.Callback <- &TaskStatus{
		Id:       id,
		Ref:      t,
		Result:   r,
		Finished: finished,
	}
}

type TaskResult protocol.Result

type TaskStatus struct {
	Id       ElementId
	Ref      *TaskRef
	Result   TaskResult
	Finished bool
}

type TaskControl struct {
	Status TaskStatus
	Cancel context.CancelFunc
}
