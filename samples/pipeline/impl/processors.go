package impl

import (
	"github.com/cloudevents/sdk-go/pipeline"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

type SourceLengthCalculator struct {
	
}

func (ac *SourceLengthCalculator) Process(t *pipeline.Task) *pipeline.ProcessorOutput{
 
	mr,ok := t.Event.(binding.MessageMetadataReader)
	if !ok {
		e,err := binding.ToEvent(t.Context,t.Event,nil)
		if err == nil {
			mr = binding.ToMessage(e).(binding.MessageMetadataReader)
		} else {
			return &pipeline.ProcessorOutput{
				Result:   pipeline.TaskResult{
					Error:  err,
					Result: nil,
				},
				FollowUp: nil,
				Changes:  nil,
			}
		}
	}

	_,attr := mr.GetAttribute(spec.Source)
	src := attr.(string)
	return &pipeline.ProcessorOutput{
		Result:   pipeline.TaskResult{
			Error:  nil,
			Result: len(src),
		},
		FollowUp: nil,
		Changes:  nil,
	}

	
}

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
		Result:   pipeline.TaskResult{
			Error:  ces.Send(t.Context, t.Event, t.Changes...),
			Result: nil,
		},
		FollowUp: nil,
		Changes:  nil,
	}
}

type Counter struct{}




