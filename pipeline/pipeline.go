package pipeline

import (
	"context"
	"fmt"
	"time"
)

type ElementId string

type Element interface {
	StartStop
	Id() ElementId
}

type Runner interface {
	Element
	Push(*TaskContainer)
}

type Pipeline struct {
	id ElementId
	reverseElements []Element
}

func (p *Pipeline) Id() ElementId{
	return p.id
}

func (p *Pipeline) Start() error {
	for i, s := range p.reverseElements {
		if err := s.Start(); err != nil {
			// If the start of the current step fails, stop again the following steps, that were already started successfully.
			if i > 0 {
				ctx, _ := context.WithTimeout(context.Background(), time.Second)
				p.Stop(ctx)
			}
			return err
		}
	}
	return nil
}

// Drain drains the pipeline by shutting off inbound tasks and stopping each step
// as soon as it contains no more pending tasks
func (p *Pipeline) Stop(ctx context.Context) {
	for i := len(p.reverseElements); i >= 0; i-- {
		p.reverseElements[i].Stop(ctx)
	}
}

type RunnerPipeline struct {
	r Runner
	*Pipeline
}

func (rp *RunnerPipeline) Push(tc *TaskContainer) {
	rp.r.Push(tc)
}

var _ Runner = (*RunnerPipeline)(nil)

type ElementConstructor func(nextStep Runner) (Element, error)

type constructorListEntry struct {
	construct ElementConstructor
	before    *constructorListEntry
}

type PipelineBuilder struct {
	lastConstructor *constructorListEntry
	length          int
}

func (pb *PipelineBuilder) Then(construct ElementConstructor) *PipelineBuilder {
	if pb.lastConstructor == nil {
		pb.lastConstructor = &constructorListEntry{
			construct: construct,
			before:    nil,
		}
		pb.length = 1
	} else {
		newEnd := &constructorListEntry{
			construct: construct,
			before:    pb.lastConstructor,
		}
		pb.lastConstructor = newEnd
		pb.length++
	}
	return pb
}

func (pb *PipelineBuilder) Build() (*Pipeline,error) {
	reverseElements := make([]Element, pb.length)

	next,err := pb.lastConstructor.construct(nil)
	if err != nil{
		return nil,err
	}
	reverseElements[0] = next
	for entry, i := pb.lastConstructor.before, 1; entry != nil; entry, i = entry.before, i+1 {
		var err error
		r,ok := next.(Runner)
		if ok {
			next,err = entry.construct(r)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("Element %s is not a Runner", next.Id())
		}
		reverseElements[i] = next
	}

	return &Pipeline{
		reverseElements: reverseElements,
	},nil
}

func (pb *PipelineBuilder) BuildRunnerPipeline() (*RunnerPipeline,error) {
	p,err := pb.Build()
	if err != nil {
		return nil,err
	}
	r, ok := p.reverseElements[len(p.reverseElements)-1].(Runner)
	if ok {
		return &RunnerPipeline{
			r:        r,
			Pipeline: p,
		}, nil
	}
	return nil, fmt.Errorf("Pipeline %s is not a Runner", p.Id())
}