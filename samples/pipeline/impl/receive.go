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
	client   *amqp.Client
	session  *amqp.Session
	receiver *amqp.Receiver
	handler  *AMQPReceiveHandler
)

type AMQPReceiveHandler struct {
	receiver *amqp.Receiver
}

func (ah *AMQPReceiveHandler) HandleResult(event binding.Message, tc *pipeline.TaskControl) bool {
	if tc.Status.Result.Error != nil {
		log.Printf("Element: %s, Error: %s\n", tc.Status.Id, tc.Status.Result.Error.Error())
	}
	if tc.Status.Finished {

		msg := event.(*amqp2.Message)
		if protocol.IsNACK(tc.Status.Result.Error) {
			msg.AMQP.Reject(&amqp.Error{
				Condition:   amqp.ErrorInternalError,
				Description: tc.Status.Result.Error.Error(),
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

func SetupAMQPLink() (*AMQPReceiveHandler, error) {
	var err error
	client, err = amqp.Dial(addr,
		amqp.ConnSASLPlain("artemis", "simetraehcapa"),
		amqp.ConnIdleTimeout(0))
	if err != nil {
		return nil, err
	}
	session, err = client.NewSession()
	if err != nil {
		return nil, err
	}

	receiver, err = session.NewReceiver(amqp.LinkSourceAddress("test"))

	if err != nil {
		return nil, err
	}

	handler = &AMQPReceiveHandler{
		receiver: receiver,
	}
	return handler, nil
}

func TearDownAMQP(ctx context.Context) {
	handler.receiver.Close(ctx)
	session.Close(ctx)
	client.Close()
}

type SdkReceiver struct {
	opener   protocol.Opener
	receiver protocol.Receiver
	inboundCancel context.CancelFunc
}

func (sr *SdkReceiver) Start() error {
	var ctx context.Context
	ctx,sr.inboundCancel = context.WithCancel(context.Background())
	go sr.opener.OpenInbound(ctx)
	return nil
}

func (sr *SdkReceiver) Stop(ctx context.Context) {
	sr.inboundCancel()
	if c,ok := sr.receiver.(protocol.Closer);ok {
		c.Close(ctx)
	}
}

func (sr *SdkReceiver) HandleResult(event binding.Message, ts *pipeline.TaskControl) bool {
	if ts.Status.Finished {
		event.Finish(ts.Status.Result.Error)
		return true
	}
	return false
}

func (sr *SdkReceiver) Receive(ctx context.Context) (*pipeline.Task, error) {
	m,err := sr.receiver.Receive(ctx)
	return &pipeline.Task{
		Context: context.TODO(),
		Event:   m,
		Changes: nil,
	},err
}

func NewSdkReceiver(o protocol.Opener, r protocol.Receiver) *SdkReceiver {
	return &SdkReceiver{
		opener:   o,
		receiver: r,
	}
}
