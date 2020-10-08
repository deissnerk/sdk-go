package impl

import (
	"github.com/cloudevents/sdk-go/pipeline"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type EventEnricher struct{
	Name string
	Value string
}

func (ee *EventEnricher) Process(t *pipeline.Task) *pipeline.ProcessorOutput {
	//	binding.TransformerFunc()
	tr := transformer.AddExtension(ee.Name, ee.Value)
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

type Counter struct{}




