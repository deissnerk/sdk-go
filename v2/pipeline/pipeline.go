package pipeline

import (
	"sync"
)

type ElementId *string

type Element interface {
	Start(wg *sync.WaitGroup)
	Stop()
	Id() ElementId
	SetNextStep(runner Runner)
}

type Pipeline struct {
	steps []Element
	wg    *sync.WaitGroup
}

func (p *Pipeline) Wait() {
	p.wg.Wait()
}

func (p *Pipeline) Start() {
	for _, s := range p.steps {
		s.Start(p.wg)
	}
}

type PipelineBuilder struct {
	steps []Element
	wg    *sync.WaitGroup
}

func NewPipelineBuilder() *PipelineBuilder {
	return &PipelineBuilder{
		wg:  &sync.WaitGroup{}}
}

func (pb *PipelineBuilder) StartWithInbound(pe Element) *PipelineBuilder {
	pb.steps = make([]Element, 1)
	pb.steps[0] = pe
	return pb
}

func (pb *PipelineBuilder) StartWithProcessor(p Processor, id ElementId) *PipelineBuilder {
	pb.steps = make([]Element, 1)
	pb.steps[0] = NewWorker(p, id)
	return pb

}

func (pb *PipelineBuilder) ContinueWith(p Processor, id ElementId) *PipelineBuilder {
	nw := NewWorker(p, id)
	pb.steps[len(pb.steps)-1].SetNextStep(nw)
	pb.steps = append(pb.steps, nw)
	return pb
}

func (pb *PipelineBuilder) EndWith(p Processor, id ElementId) *Pipeline {
	nw := NewWorker(p, id)
	pb.steps[len(pb.steps)-1].SetNextStep(nw)
	pSteps := make([]Element, len(pb.steps)+1)
	for i, s := range pb.steps {
		pSteps[i] = s
	}
	pSteps[len(pSteps)-1] = nw

	return &Pipeline{
		steps: pSteps,
		wg:    pb.wg,
	}
}

func (pb *PipelineBuilder) SplitWith(ts TaskSplitter, id ElementId) *Pipeline {
	sv := NewSupervisor(ts, id)
	pb.steps[len(pb.steps)-1].SetNextStep(sv)
	pSteps := make([]Element, len(pb.steps)+1)
	for i, s := range pb.steps {
		pSteps[i] = s
	}
	pSteps[len(pSteps)-1] = sv

	return &Pipeline{
		steps: pSteps,
		wg:    pb.wg,
	}
}
