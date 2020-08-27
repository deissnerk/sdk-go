# Event Pipe

## Intro
_Event Pipe_ defines several components that can be combined in a flexible way
to construct pipelines for processing CloudEvents. An example of such a
pipeline could be an inbound component fetching CloudEvents from an AMQP
link and passing them on to one or multiple routers. The routers in turn
generate a list of endpoints to forward the event to and pass the event to
the according senders. All steps are asynchronous, and a configurable
sliding window provides the means to absorb short bursts.

## Runtime
### Runners

### Supervisors

## Pipeline ELements
### Split/Join

### Router

### Inbound
The *Inbound* can be the first element of a pipeline. Its purpose is to fetch
event messages from outside and feed them into the pipeline.


## Processors



 
## Examples
  
## Future 

### Describe Pipelines in mark-up

### A Runner for WASM Processors
