package elements

import (
	"context"
	"github.com/cloudevents/sdk-go/pipeline"
)

// TODO Add ReceiveHandlerConstructor
func CreateInbound(handler ResultHandler,receiver Receiver, parentCtx context.Context,id pipeline.ElementId) pipeline.ElementConstructor{
	return func(nextStep pipeline.Runner) (pipeline.Element,error) {
		return NewInbound(handler,receiver,parentCtx,id, nextStep),nil
	}
}

func Process(pc pipeline.ProcessorConstructor, id pipeline.ElementId) pipeline.ElementConstructor {
	return func(nextStep pipeline.Runner) (pipeline.Element,error) {
		p,err := pc()
		if err != nil {
			return nil, err
		}
		return pipeline.NewWorker(p,id,nextStep),nil
	}
}

func Use(pc pipeline.ProcessorFuncConstructor, id pipeline.ElementId) pipeline.ElementConstructor {
	return func(nextStep pipeline.Runner) (pipeline.Element,error) {
		p,err := pc()
		if err != nil {
			return nil, err
		}
		return pipeline.NewWorker(p,id,nextStep),nil
	}
}

func Split(sc SplitterConstructor, id pipeline.ElementId) pipeline.ElementConstructor{
	return func(nextStep pipeline.Runner) (pipeline.Element,error) {
		sc,err := sc()
		if err != nil {
			return nil, err
		}
		return NewSplitterRunner(sc,id,nextStep)
	}
}

