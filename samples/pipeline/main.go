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

var (
)


func main() {
	// Create AMQP client

	// Close connection and session in the end
	defer func() {
		closeCtx, _ := context.WithTimeout(context.Background(), time.Second*10)
		impl.TearDownAMQP(closeCtx)
	}()

	handler,err := impl.SetupAMQPLink()
	if err != nil {
		log.Fatalf(err.Error())
	}

	sender,err := http.New(http.WithTarget("http://localhost:8090/"))
	if err != nil{
		log.Fatalf(err.Error())
	}

	pb := &pipeline.PipelineBuilder{}
	pb.Then(elements.CreateInbound(handler, context.Background(), "Inbound")).
		Then(elements.Process(&impl.EventEnricher{}, "Enricher")).
		Then(elements.Split(&impl.SampleSplitter{},"Splitter")).
		Then(elements.Process(&impl.CeSender{sender}, "Sender"))

	pipe := pb.Build()
	pipe.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<- c
	close(c)
	stopCtx,_ := context.WithTimeout(context.Background(),time.Second*10)
	pipe.Stop(stopCtx)

}

