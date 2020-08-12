package pipeline

import (
	"sync"
)

type ElementId string

type Element interface {
	StartStop
	Id() ElementId
}

type Pipeline struct {
	Element
	reverseElements []Element
	wg              *sync.WaitGroup
}

func (p *Pipeline) Start() {
	for _, s := range p.reverseElements {
		s.Start(p.wg)
	}
}

// Drain drains the pipeline by shutting off inbound tasks and stopping each step
// as soon as it contains no more pending tasks
func (p *Pipeline) Stop() {
	p.reverseElements[len(p.reverseElements)-1].Stop()
	p.wg.Wait()
}

type ElementConstructor func(nextStep Runner) Element


type constructorListEntry struct {
	construct ElementConstructor
	before    *constructorListEntry
}

type PipelineBuilder struct {
	lastConstructor *constructorListEntry
	length             int
}

func (pb *PipelineBuilder) Then(construct ElementConstructor) *PipelineBuilder{
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

func (pb *PipelineBuilder) Build() *Pipeline {
	reverseElements := make([]Element, pb.length)

	next := pb.lastConstructor.construct(nil)
	reverseElements[0] = next
	for entry, i := pb.lastConstructor.before, 1; entry != nil; entry, i = entry.before, i+1 {
		next = entry.construct(next.(Runner))
		reverseElements[i] = next
	}

	return &Pipeline{
		reverseElements: reverseElements,
		wg:              &sync.WaitGroup{},
	}
}
