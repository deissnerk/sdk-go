# Event Pipe

## Intro
Event Pipe defines several components that can be combined in a flexible way
to construct pipelines for processing CloudEvents. An example of such a
pipeline could be an inbound component that fetches CloudEvents from an AMQP
endpoint and passes them on to one or multiple routers. The routers in turn
generate a list of endpoints to forward the event to and pass the event to
the according senders. All steps are asynchronous, and a configurable
transmission window provides the means to absorb short bursts.

## The Role of CloudEvents
 
## Pipeline Elements

### Inbound
The *Inbound* can be the first element of a pipeline. Its purpose is to fetch
event messages from outside and feed them into the pipeline. The first
inbound that has been implemented, is an AMQP inbound that opens a receiving
link and forwards messages into the next step.

### Processor

### Splitter
 

  
 
    