# Event Pipe

## Intro
Event Pipe defines several components that can be combined in a flexible way
to construct pipelines for processing CloudEvents. An example of such a
pipeline could be an inbound component fetching CloudEvents from an AMQP
endpoint and passing them on to one or multiple routers. The routers in turn
generate a list of endpoints to forward the event to and pass the event to
the according senders. All steps are asynchronous, and a configurable
transmission window provides the means to absorb short bursts.

## Pipeline Elements

### Processor

### Supervisors

#### Split/Join

#### Router

#### Inbound
The *Inbound* can be the first element of a pipeline. Its purpose is to fetch
event messages from outside and feed them into the pipeline.

### SupervisorState

 
## Examples
  
 
    