package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type TaskIndex uint32

type Task struct {
	Context context.Context
	Event   binding.Message
}

type TaskCancelledError struct {
	Err 	error
}

func (tce TaskCancelledError) Error() string {
	return "Task cancelled: "+tce.Err.Error()
}

// TaskContainer holds all information a Runner needs to execute the Task
type TaskContainer struct {
	Callback chan *StatusMessage
	Key      interface{}
	Task     Task
	Parent   *TaskContainer
	Changes  []binding.Transformer
}

func (t *TaskContainer) SendStatusUpdate(id ElementId, r TaskResult, finished bool) {
	if t.Callback == nil {
		if t.Parent != nil {
			t.Parent.SendStatusUpdate(id, r, finished)
		}
	} else {
		t.Callback <- &StatusMessage{
			Key: t.Key,
			Status: TaskStatus{
				Id: id,
				//			Ref:      t,
				Result:   r,
				Finished: finished,
			},
		}
	}
}

func (t *TaskContainer) SendCancelledUpdate() {
	t.SendStatusUpdate(nil, TaskCancelledError{Err: t.Task.Context.Err()}, false)
}

// FollowUp() returns a new TaskContainer that is a follow-up to the existing one
func (t *TaskContainer) FollowUp(output *ProcessorOutput) *TaskContainer {
	// Record changes
	t.Changes = output.Changes

	followUp := TaskContainer{
		Callback: t.Callback,
		Key:      t.Key,
		Task:     t.Task,
		Parent:   t,
		Changes:  nil, // No changes yet
	}

	// If the output contains a new Context, this is used
	if output.FollowUp != nil {
		followUp.Task.Context = output.FollowUp
	}
	return &followUp
}

func (t *TaskContainer) CollectChanges() []binding.Transformer {
	if t.Changes != nil {
		return append(t.getParentChanges(len(t.Changes)), t.Changes...)
	}
	return t.getParentChanges(0)
}

func (t *TaskContainer) getParentChanges(length int) []binding.Transformer {
	if t.Parent == nil {
		return make([]binding.Transformer, 0, length)
	}
	return append(t.Parent.getParentChanges(length+len(t.Parent.Changes)), t.Parent.Changes...)
}

type TaskResult protocol.Result

type TaskStatus struct {
	Id ElementId
	//	Ref      *TaskContainer
	Result   TaskResult
	Finished bool
}

type StatusMessage struct {
	Key    interface{}
	Status TaskStatus
}

type TaskControl struct {
	Status TaskStatus
	Cancel context.CancelFunc
}
