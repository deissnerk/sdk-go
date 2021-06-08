package pipeline

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type TaskIndex uint32

type Task struct {
	ctx context.Context
	//Event   binding.Message
	changes []binding.Transformer
	mr      binding.MessageMetadataReader
	msg     binding.Message
	cancel  context.CancelFunc
	output  ProcessorOutput
}

func (t *Task) Context() context.Context {
	return t.ctx
}

func (t *Task) Changes() []binding.Transformer {
	return t.changes
}

func (t *Task) ReadEncoding() binding.Encoding {
	return t.msg.ReadEncoding()
}

func (t *Task) ReadStructured(ctx context.Context, writer binding.StructuredWriter) error {
	return t.msg.ReadStructured(ctx, writer)
}

// This is currently not thread safe. It should only be called once in the whole pipeline
func (t *Task) ReadBinary(ctx context.Context, writer binding.BinaryWriter) error {
	return t.msg.ReadBinary(ctx, writer)
}

// It is not guaranteed that this can be called multiple times
// A more sophisticated approach, e.g. with ref count, has not been implemented
// to reduce performance overhead.
// An alternative implementation of Task with different performance trade-offs could be added later
func (t *Task) Finish(err error) error {
	// TODO Idea: Collect output and use it in tc.NextStep() to generate new tc + task. finished is set by container to state, when the pipeline is finished
	t.output.Result = TaskResult{
		Error:  err,
		Result: nil,
	}
	return nil
	//	return t.msg.Finish(err)
}

func (t *Task) FinishWithResult(err error, result interface{}) error {
	t.output.Result = TaskResult{
		Error:  err,
		Result: result,
	}
	return nil
}

func (t *Task) GetAttribute(attributeKind spec.Kind) (spec.Attribute, interface{}) {
	return t.mr.GetAttribute(attributeKind)
}

func (t *Task) GetExtension(name string) interface{} {
	return t.mr.GetExtension(name)
}

func (t *Task) GetWrappedMessage() binding.Message {
	return t.msg
}

func (t *Task) SetChanges(transformers []binding.Transformer) {
	t.output.Changes = transformers
}

func (t *Task) SetFollowUpContext(ctx context.Context) {
	t.output.FollowUp = ctx
}

// NewAccessMetadataTask creates a Task that is optimized for cases, when task processing is
// only aiming at the event context. If the message is binary encoded, this means that data
// does not have to be buffered, as it is only read once in the end.
// Messages with structured encoding are converted to event messages, as parsing is needed
// to access the event context.
func NewAccessMetadataTask(ctx context.Context, m binding.Message) (*Task, error) {
	t := &Task{}
	t.ctx, t.cancel = context.WithCancel(ctx)

	var ok bool
	if t.mr, ok = m.(binding.MessageMetadataReader); ok {
		t.msg = m
	} else {
		ev, err := binding.ToEvent(ctx, m, nil)
		if err != nil {
			return nil, err
		}
		t.msg = binding.ToMessage(ev)
		t.mr = t.msg.(binding.MessageMetadataReader)
	}
	return t, nil
}

func (t *Task) NewSubTask() *Task {
	ctx, cancel := context.WithCancel(t.ctx)
	return &Task{
		ctx:     ctx,
		changes: nil,
		// Should we use t.mr and t.msg or just t?
		// If we use t, this would enable recursive calls, e.g. to Finish()
		mr:     t.mr,
		msg:    t.msg,
		cancel: cancel,
	}
}

type TaskCancelledError struct {
	Err error
}

func (tce TaskCancelledError) Error() string {
	return "Task cancelled: " + tce.Err.Error()
}

// TaskContainer holds all information a Runner needs to execute the Task
type TaskContainer struct {
	owner    ElementId
	callback chan *StatusMessage
	key      interface{}
	*Task
	parent *TaskContainer
}

func NewRootContainer(t *Task) *TaskContainer {
	// A root container just contains the task.
	// Everything is nil.
	tc := &TaskContainer{
		Task: t,
	}

	return tc
}

func (t *TaskContainer) NewChild(owner ElementId, callback chan *StatusMessage, key interface{}, task *Task) *TaskContainer {
	return &TaskContainer{
		owner:    owner,
		callback: callback,
		key:      key,
		Task:     task,
		parent:   t,
	}
}

func (t *TaskContainer) NextStep() *TaskContainer {

}

// A task can only be cancelled through its container
func (t *TaskContainer) Cancel() {
	t.cancel()
}

func (t *TaskContainer) Key() interface{} {
	return t.key
}

func (t *TaskContainer) SendStatusUpdate() {
	if t.callback != nil {
		sm := &StatusMessage{
			Key:    t.key,
			Id:     t.owner,
// TODO What about Finished?
			Status: t.status,
		}
		t.callback <- sm
	}
}

func (t *TaskContainer) sendCancelledUpdate() {
	t.status.Result = TaskResult{
		Error:  TaskCancelledError{Err: t.Task.ctx.Err()},
		Result: nil,
	}
	t.status.Finished = false
	t.SendStatusUpdate()
}

// Idea
// States of a Task: ready -> busy(bound to element) -> stepFinished -> next iteration

// AddOutput() uses ProcessorOutput to adjust the TaskContainer for the next step
// TODO There is a certain redundancy with Task.AddChanges() etc.
// Perhaps on container level it would be better to have a "wrap up" function?
// It could determine, if the real Finish() should be called
// Is ProcessorOutput still needed, if we collect everything in the Task?
func (t *TaskContainer) AddOutput(output *ProcessorOutput) {
	// Record changes
	if t.Task.changes != nil {
		t.Task.changes = append(t.Task.changes, output.Changes...)
	} else {
		t.Task.changes = output.Changes
	}

	// If the output contains a new Context, this is used
	if output.FollowUp != nil {
		t.Task.ctx = output.FollowUp
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

// This is all quite strange. protocol.Receipt and protocol.Error are not clearly defined.
type TaskResult struct {
	Error  protocol.Result
	Result interface{}
}

type TaskStatus struct {
	Result   TaskResult
	Finished bool
}

type StatusMessage struct {
	Id     ElementId
	Key    interface{}
	Status TaskStatus
}

type TaskControl struct {
	Status TaskStatus
	Cancel context.CancelFunc
}
