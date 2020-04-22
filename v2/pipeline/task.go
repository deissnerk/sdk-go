package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
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
//
//const (
//	Pending TaskAck = iota
//	Stored
//	Completed
//	Failed
//	Retry
//)

type TaskResult protocol.Result

type TaskStatus struct {
	Id       ElementId
	Ref      *TaskRef
	Result   TaskResult
	Finished bool
}
