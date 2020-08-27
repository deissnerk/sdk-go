package impl

import (
	"context"
	"github.com/Azure/go-amqp"
	"github.com/cloudevents/sdk-go/pipeline"
	amqp2 "github.com/cloudevents/sdk-go/protocol/amqp/v2"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"log"
)

const (
	addr = "amqp://localhost"
)

var (
	client *amqp.Client
	session *amqp.Session
	receiver *amqp.Receiver
	handler *AMQPReceiveHandler
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

func SetupAMQPLink() (*AMQPReceiveHandler,error){
	var err error
	client, err = amqp.Dial(addr,
		amqp.ConnSASLPlain("artemis", "simetraehcapa"),
		amqp.ConnIdleTimeout(0))
	if err != nil {
		return nil,err
	}
	session, err = client.NewSession()
	if err != nil {
		return nil,err
	}

	receiver, err = session.NewReceiver(amqp.LinkSourceAddress("test"))
	if err != nil {
		return nil,err
	}

	handler = &AMQPReceiveHandler{
		receiver: receiver,
	}
	return handler,nil
}

func TearDownAMQP(ctx context.Context){
	handler.receiver.Close(ctx)
	session.Close(ctx)
	client.Close()
}