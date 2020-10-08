package main

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"os"
	"os/signal"
	"pipeline/impl"
	"time"

	"github.com/cloudevents/sdk-go/pipeline"
	"github.com/cloudevents/sdk-go/pipeline/elements"
	"log"
)

func main() {
	// Create AMQP client

	// Close connection and session in the end
	defer func() {
		closeCtx, _ := context.WithTimeout(context.Background(), time.Second*10)
		impl.TearDownAMQP(closeCtx)
	}()

	handler, err := impl.SetupAMQPLink()
	if err != nil {
		log.Fatalf(err.Error())
	}

	sender, err := http.New(http.WithTarget("http://localhost:8080/"))
	if err != nil {
		log.Fatalf(err.Error())
	}

	pb := &pipeline.PipelineBuilder{}
	pb.Then(elements.CreateInbound(handler, context.Background(), "Inbound")).
		Then(elements.Process(func() (pipeline.Processor, error) {
			return &impl.EventEnricher{
				Name:  "step1",
				Value: "EventEnricher",
			}, nil
		}, "Enricher")).
		//		Then(elements.Split(&impl.SampleSplitter{},"SplitterRunner")).
		Then(impl.Split()).
		Then(elements.Process(func() (pipeline.Processor, error) {
			return &impl.CeSender{sender}, nil
		}, "Sender"))
	pipe, err := pb.Build()
	if err == nil {
		pipe.Start()

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		close(c)
		stopCtx, _ := context.WithTimeout(context.Background(), time.Second*10)
		pipe.Stop(stopCtx)
	} else {
		log.Fatalf(err.Error())
	}

}
