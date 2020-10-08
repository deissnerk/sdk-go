package pipeline

import (
	"context"
	"fmt"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type TaskIndex uint32

type Task struct {
	Context context.Context
	Event   binding.Message
	Changes []binding.Transformer
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
}

func (t *TaskContainer) SendStatusUpdate(id ElementId, r TaskResult, finished bool) {
	if t.Callback != nil {
		sm := &StatusMessage{
			Key: t.Key,
			Status: TaskStatus{
				Id: id,
				//			Ref:      t,
				Result:   r,
				Finished: finished,
			}}
		fmt.Printf("Send Update %+v\n",sm)
		t.Callback <- sm

	}
}

func (t *TaskContainer) SendCancelledUpdate() {
	t.SendStatusUpdate("", TaskCancelledError{Err: t.Task.Context.Err()}, false)
}

// AddOutput() uses ProcessorOutput to adjust the TaskContainer for the next step
func (t *TaskContainer) AddOutput(output *ProcessorOutput) {
	// Record changes
	if t.Task.Changes != nil {
		t.Task.Changes = append(t.Task.Changes, output.Changes...)
	} else {
		t.Task.Changes = output.Changes
	}

	// If the output contains a new Context, this is used
	if output.FollowUp != nil {
		t.Task.Context = output.FollowUp
	}
}

//func (t *TaskContainer) CollectChanges() []binding.Transformer {
//	if t.Changes != nil {
//		return append(t.getParentChanges(len(t.Changes)), t.Changes...)
//	}
//	return t.getParentChanges(0)
//}
//
//func (t *TaskContainer) getParentChanges(length int) []binding.Transformer {
//	if t.Parent == nil {
//		return make([]binding.Transformer, 0, length)
//	}
//	return append(t.Parent.getParentChanges(length+len(t.Parent.Changes)), t.Parent.Changes...)
//}

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
