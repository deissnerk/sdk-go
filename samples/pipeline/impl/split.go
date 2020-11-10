package impl

import (
	"context"
	"github.com/cloudevents/sdk-go/pipeline"
	"github.com/cloudevents/sdk-go/pipeline/elements"
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/transformer"
)

type SampleSplitter struct {
	pipeA *pipeline.RunnerPipeline
	pipeB *pipeline.RunnerPipeline
}

func (s SampleSplitter) Split(origin *pipeline.Task) ([]*elements.TaskAssignment,elements.Joiner) {
	ret := make([]*elements.TaskAssignment, 2)

	ctxA, cancelA := context.WithCancel(origin.Context)
	ret[0] = &elements.TaskAssignment{
		Task: &pipeline.Task{
			Context: ctxA,
			Event:   origin.Event,
			Changes: nil,
		},
		Runner: s.pipeA,
		Cancel: cancelA,
	}

	ctxB, cancelB := context.WithCancel(origin.Context)
	ret[1] = &elements.TaskAssignment{
		Task: &pipeline.Task{
			Context: ctxB,
			Event:   origin.Event,
			Changes: nil,
		},
		Runner: s.pipeB,
		Cancel: cancelB,
	}

	return ret,&sampleJoiner{}
}

type sampleJoiner struct {

}

func (j *sampleJoiner) Join(status []*pipeline.TaskControl) *pipeline.ProcessorOutput {
	tr := transformer.AddExtension("sourcelength", status[1].Status.Result.Result)
	ts := make([]binding.Transformer, 1)
	ts[0] = tr

	return &pipeline.ProcessorOutput{
		Result: pipeline.TaskResult{},
		FollowUp: nil,
		Changes:  ts,
	}

}

func (j *sampleJoiner) IsFinished(status []*pipeline.TaskControl) bool {
	for _,s := range status {
		if !s.Status.Finished {
			return false
		}
	}
	return true
}

func (s SampleSplitter) Start() error {
	if err := s.pipeA.Start(); err != nil {
		return err
	}
	return s.pipeB.Start()

}

func (s SampleSplitter) Stop(ctx context.Context) {
	s.pipeA.Stop(ctx)
	s.pipeB.Stop(ctx)
}

func Split() pipeline.ElementConstructor {
	pbA := pipeline.PipelineBuilder{}
	pbA.Then(elements.Use(func() (pipeline.ProcessorFunc,error) {
		counter := 0
		return func(t *pipeline.Task) (*pipeline.ProcessorOutput) {
			counter++
			return &pipeline.ProcessorOutput{
				Result:   pipeline.TaskResult{
					Error:  nil,
					Result: counter,
				},
				FollowUp: nil,
				Changes:  nil,
			}
		}, nil
	},"pipeA, step 1"))

	pbB := pipeline.PipelineBuilder{}
	pbB.Then(elements.Process(func() (pipeline.Processor,error){
		return &SourceLengthCalculator{},nil
	},"pipeB, step 1"))

	return elements.Split(func() (elements.Splitter, error) {

		pipeA, err := pbA.BuildRunnerPipeline()
		if err != nil {
			return nil, err
		}
		pipeB, err := pbB.BuildRunnerPipeline()
		if err != nil {
			return nil, err
		}

		return &SampleSplitter{
			pipeA: pipeA,
			pipeB: pipeB,
		}, nil
	}, "SplitterRunner")
}
