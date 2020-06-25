package elements

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/pkg/binding"
	"github.com/cloudevents/sdk-go/v2/pipeline"
	http2 "github.com/cloudevents/sdk-go/v2/protocol/http"
	"log"
	"net/http"
	"net/url"
)

var _ pipeline.Processor = (*HttpSender)(nil)

type HttpSender struct {
	target    *url.URL
	p http2.Protocol
}

func (s *HttpSender) Process(tr *pipeline.TaskContainer) pipeline.ProcessorOutput {
// Would prefer to use the http binding, but it does not provide access to its
// encoders.
	ev,_,_,err := binding.ToEvent(tr.Task.Event)
	ctx,_,err := s.transport.Send(tr.Task.Context,ev)

	if err != nil{
		return pipeline.ProcessorOutput{
			Ack: pipeline.Failed,
			Err: err,
		}
	}

	rctx := cehttp.TransportContextFrom(ctx)
	res := pipeline.ProcessorOutput{
		Err:      nil,
	}
	switch rctx.StatusCode {
	case 408: fallthrough	// Request timed out
	case 500: fallthrough  	// Internal error
	case 502: fallthrough	// Bad Gateway
	case 503: fallthrough	// Service unavailable
	case 504: 				// Gateway Timeout
		res.Ack = pipeline.Retry // For all these errors a retry might make sense
	case 202:
		res.Ack = pipeline.Stored
	case 200: fallthrough
	case 201: fallthrough
	case 204:
		res.Ack = pipeline.Completed
	default:
		res.Ack = pipeline.Failed

	}
	return res
}

func NewHttpSender(target string, encoding cehttp.Encoding, client *http.Client) (*HttpSender, error) {
	ctx := cloudevents.Context(context.Background())

	p, err := cloudevents.NewHTTP( http2.WithTarget(target))
	if err != nil {
		log.Fatalf("failed to create protocol: %s", err.Error())
	}

	c, err := cloudevents.NewClient(p, cloudevents.WithTimeNow(), cloudevents.WithUUIDs())
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}


	transport, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(target),
		cloudevents.WithEncoding(encoding),
	)
	if err != nil{
		return nil, err
	}

	targetUrl,err := url.Parse(target)
	if err != nil{
		return nil, err
	}

	s := &HttpSender{
		target:    targetUrl,
		transport: transport,
	}

	return s, nil
}
