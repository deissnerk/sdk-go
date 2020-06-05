package main

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/v2/pipeline"
	"github.com/cloudevents/sdk-go/v2/pipeline/config"
	"github.com/cloudevents/sdk-go/v2/pipeline/elements"
	"log"
	"net/http"
	"pack.ag/amqp"
	"time"
)

const (
	addr = "amqp://localhost"
)

func main() {

	// Create AMQP client
	client, err := amqp.Dial(addr,
		amqp.ConnSASLPlain("artemis", "simetraehcapa"),
		amqp.ConnIdleTimeout(0))
	if err != nil {
		log.Panicf(err.Error())
	}
	session, err := client.NewSession()
	if err != nil {
		log.Panicf(err.Error())
	}

	// Close connection and session in the end
	defer func() {
		if closeCtx, err := context.WithTimeout(context.Background(), time.Second*10); err != nil {
			session.Close(closeCtx)
		}
		client.Close()
	}()

	cfg := config.Config{
		WSize:   8,
		Session: session,
		Client:  client,
	}

	senders := make([]*elements.HttpSender, 6)
	httpClient := &http.Client{}
	for i, _ := range senders {
		s, err := elements.NewHttpSender("http://localhost:8080", cloudevents.HTTPBinaryV03, httpClient)
		if err == nil {
			senders[i] = s
		}
	}

	rRouterId := "RandomRouter"
	pb := pipeline.NewPipelineBuilder(&cfg)

	pipe := pb.
		StartWithInbound(pipeline.NewInbound(context.Background(), &cfg, "testtopic", nil)).
		SplitWith(NewRandomRouter(senders, &cfg), &rRouterId)
	pipe.Start()
	pipe.Wait()

}
