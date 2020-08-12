package main

import (
	"context"
	amqp2 "github.com/cloudevents/sdk-go/protocol/amqp/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"os"
	"os/signal"
	"time"

	"github.com/cloudevents/sdk-go/pipeline"
	"github.com/cloudevents/sdk-go/pipeline/elements"
	"log"

	"github.com/Azure/go-amqp"
)

const (
	addr = "amqp://localhost"
)

type AMQPReceiveHandler struct {
	receiver *amqp.Receiver
}

func (ah *AMQPReceiveHandler) HandleResult(event binding.Message, tc *pipeline.TaskControl) bool {
	if tc.Status.Result != nil {
		log.Printf("Element: %s, Result: %s\n", tc.Status.Id, tc.Status.Result.Error())
	}
	if tc.Status.Finished {

		msg := event.(*amqp2.Message)
		if protocol.IsNACK(tc.Status.Result) {
			msg.AMQP.Reject(&amqp.Error{
				Condition:   amqp.ErrorInternalError,
				Description: tc.Status.Result.Error(),
				Info:        nil,
			})
		} else {
			msg.AMQP.Accept()
		}
	}
	return false
}

func (ah *AMQPReceiveHandler) Receive(ctx context.Context) (*pipeline.Task, error) {
	m, err := ah.receiver.Receive(ctx)

	if err != nil {
		return nil, err
	}

	return &pipeline.Task{
		Context: context.TODO(),
		Event:   amqp2.NewMessage(m),
		Changes: nil,
	}, nil
}

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

	receiver, err := session.NewReceiver(amqp.LinkSourceAddress("test"))
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

	handler := &AMQPReceiveHandler{
		receiver: receiver,
	}
	sender,err := http.New(http.WithTarget("http://localhost:8090/"))
	pb := &pipeline.PipelineBuilder{}
	pb.Then(elements.CreateInbound(handler, context.Background(), "Inbound")).
		Then(elements.Process(&eventEnricher{}, "Enricher")).
		Then(elements.Process(&CeSender{sender}, "Sender"))

	pipe := pb.Build()
	pipe.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<- c
	close(c)
	pipe.Stop()

}

type eventEnricher struct{}

func (ee *eventEnricher) Process(t *pipeline.Task) *pipeline.ProcessorOutput {
	//	binding.TransformerFunc()
	tr := transformer.AddExtension("myextension", "bla")
	ts := make([]binding.Transformer, 1)
	ts[0] = tr

	return &pipeline.ProcessorOutput{
		Changes: ts,
	}
}

type CeSender struct {
	protocol.Sender
}

func (ces *CeSender) Process(t *pipeline.Task) *pipeline.ProcessorOutput {
	return &pipeline.ProcessorOutput{
		Result:   ces.Send(t.Context, t.Event, t.Changes...),
		FollowUp: nil,
		Changes:  nil,
	}
}
