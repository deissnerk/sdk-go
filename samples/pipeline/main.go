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

	var httpPipe, amqpPipe *pipeline.Pipeline

	// Close connection and session in the end
	defer func() {
		closeCtx, _ := context.WithTimeout(context.Background(), time.Second*10)
		impl.TearDownAMQP(closeCtx)
	}()

	handler, amqpErr := impl.SetupAMQPLink()

	sender, httpErr := http.New(http.WithTarget("http://localhost:8080/"))
	if httpErr != nil && amqpErr != nil {
		log.Fatalf("Unable to init HTTP and AMQP!")
	}
	if httpErr == nil {

		httpProtocol, err := http.New(http.WithPort(8083))
		if err != nil {
			log.Fatalf(err.Error())
		}
		httpRcvHandler := impl.NewSdkReceiver(httpProtocol, httpProtocol)

		httpPb := &pipeline.PipelineBuilder{}
		httpPb.Then(elements.CreateInbound(httpRcvHandler, context.Background(), "Inbound")).
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
		httpPipe, httpErr = httpPb.Build()
		if httpErr == nil {
			httpErr = httpPipe.Start()
		}
	}

	if amqpErr == nil {
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
		amqpPipe, amqpErr = pb.Build()
		if amqpErr == nil {
			amqpErr = amqpPipe.Start()
		}
	}

	// If something started successfully, wait for OS signal
	if amqpErr == nil || httpErr == nil {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		close(c)
		stopCtx, _ := context.WithTimeout(context.Background(), time.Second*10)
		amqpPipe.Stop(stopCtx)
		httpPipe.Stop(stopCtx)
	} else {
		log.Fatalf("Nothing worked!")
	}

}
