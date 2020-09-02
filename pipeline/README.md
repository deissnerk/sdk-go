# Event Pipe

## Intro
_Event Pipe_ defines several components that can be combined in a flexible way
to construct pipelines for processing CloudEvents. An example of such a
pipeline could be an inbound component fetching CloudEvents from an AMQP
link and passing them on to one or multiple targets enriching and validating
the events on the way. All steps are asynchronous, and a configurable
sliding window provides the means to absorb short bursts.

## Pipelines
Pipelines consist of a sequence of processors that process tasks. Each task
is based on a CloudEvent. Processors can return results, trigger external
logic and pass additional information on to the next processor. 

### Tasks
A task encapsulates the information that is passed between event processors
. It contains the event that is processed, a context object, and a list of
 changes preceding processors have already added. For the event the original
  `binding.Message` is used to allow highly optimized flows.   

```
type Task struct {
	Context context.Context
	Event   binding.Message
	Changes []binding.Transformer
}
```

### Processors
A processor takes a `Task` as input and returns a `ProcessorOutput`:
```
type ProcessorOutput struct {
	Result   TaskResult
	FollowUp context.Context
	Changes  []binding.Transformer
}
```
This output can contain a result or error, a context and a list of event
transformers. While the result is passed back to the initiator of the
pipeline, the context is used to call the next processor. 

## Combining Pipelines
Pipelines can be combined using different elements.

### Inbound

The *Inbound* can be the first element of a pipeline. Its purpose is to fetch
event messages from outside and feed them into the pipeline.

### Split/Join

The split/join element can be used to split a task into sub tasks that are
processed in parallel. Of course each sub task can again be processed by a
dedicated pipeline.
After the parallel execution the results can be joined and the next step of
the main pipeline can be triggered. 

### Router

The router is similar to split/join, but it only routes the task to a single
pipeline.

## Runtime

To provide the abstraction of pipelines and processors, a runtime model of
runners and supervisors is used. Runners have a `Push()` method that takes a
`TaskContainer` as input:
```
type TaskContainer struct {
	Callback chan *StatusMessage
	Key      interface{}
	Task     Task
	Parent   *TaskContainer
}
```
The `TaskContainer` contains all information a `Runner` needs to process a
task. There are two implementations of the `Runner` interface.

### Workers

A worker invokes a processor, passes the result to a callback channel and
pushes the task to the next worker in the pipeline runtime.

### Supervisors

Supervisors are used as base of several pipeline elements like `Inbound
`, `Splitter` or `Router`. Supervisors constitute the beginning of a pipeline
. They maintain the state of the tasks that are currently travelling through
the pipeline. The existing implementations of `SupervisorState` apply a
sliding window algorithm.  
 
## Examples
  
## Future 

### Describe Pipelines in mark-up

### A Runner for WASM Processors
