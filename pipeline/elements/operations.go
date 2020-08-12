package elements

import (
	"context"
	"github.com/cloudevents/sdk-go/pipeline"
)

func CreateInbound(handler ReceiveHandler,parentCtx context.Context,id pipeline.ElementId) pipeline.ElementConstructor{
	return func(nextStep pipeline.Runner) pipeline.Element {
		return NewInbound(handler,parentCtx,id, nextStep)

	}
}

func Process(p pipeline.Processor, id pipeline.ElementId) pipeline.ElementConstructor {
	return func(nextStep pipeline.Runner) pipeline.Element {
		return pipeline.NewWorker(p,id,nextStep)
	}
}