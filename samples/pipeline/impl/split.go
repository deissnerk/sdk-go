package impl

import (
	"context"
	"github.com/cloudevents/sdk-go/pipeline"
)

type SampleSplitter struct {
	pipeA *pipeline.Pipeline
	pipeB *pipeline.Pipeline
}

func (s SampleSplitter) Split(origin *pipeline.Task) []*pipeline.TaskAssignment {
	ret := make([]*pipeline.TaskAssignment,2)
	ret[0] = &pipeline.TaskAssignment{
		Task:   &pipeline.Task{
			Context: origin.Context,
			Event:   origin.Event,
			Changes: nil,
		},
		Runner: nil,
		Cancel: nil,
	}
	return ret
}

func (s SampleSplitter) Join(status []*pipeline.TaskControl) *pipeline.ProcessorOutput {
	panic("implement me")
}

func (s SampleSplitter) IsFinished(status []*pipeline.TaskControl) bool {
	panic("implement me")
}

func (s SampleSplitter) Start() error {
	if err:=s.pipeA.Start(); err != nil {
		return err
	}
	return s.pipeB.Start()

}

func (s SampleSplitter) Stop(ctx context.Context) {
	panic("implement me")
}


