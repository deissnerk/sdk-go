package elements

import (
	"github.com/cloudevents/sdk-go/v2/pipeline"
	"sync"
	"testing"
)

func TestSplitterState_AddTask(t *testing.T) {
	type fields struct {
		id          pipeline.ElementId
		ts          TaskSplitter
		maxWSize    pipeline.TaskIndex
		sw          *pipeline.SlidingWindow
		nextStep    pipeline.Runner
		hasNextStep bool
	}
	type args struct {
		tc       *pipeline.TaskContainer
		callback chan *pipeline.TaskStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &SplitterState{
				id:          tt.fields.id,
				ts:          tt.fields.ts,
				maxWSize:    tt.fields.maxWSize,
				sw:          tt.fields.sw,
				nextStep:    tt.fields.nextStep,
				hasNextStep: tt.fields.hasNextStep,
			}
		})
	}
}

func TestSplitterState_IsIdle(t *testing.T) {
	type fields struct {
		id          pipeline.ElementId
		ts          TaskSplitter
		maxWSize    pipeline.TaskIndex
		sw          *pipeline.SlidingWindow
		nextStep    pipeline.Runner
		hasNextStep bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &SplitterState{
				id:          tt.fields.id,
				ts:          tt.fields.ts,
				maxWSize:    tt.fields.maxWSize,
				sw:          tt.fields.sw,
				nextStep:    tt.fields.nextStep,
				hasNextStep: tt.fields.hasNextStep,
			}
			if got := st.IsIdle(); got != tt.want {
				t.Errorf("IsIdle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitterState_SetNextStep(t *testing.T) {
	type fields struct {
		id          pipeline.ElementId
		ts          TaskSplitter
		maxWSize    pipeline.TaskIndex
		sw          *pipeline.SlidingWindow
		nextStep    pipeline.Runner
		hasNextStep bool
	}
	type args struct {
		step pipeline.Runner
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &SplitterState{
				id:          tt.fields.id,
				ts:          tt.fields.ts,
				maxWSize:    tt.fields.maxWSize,
				sw:          tt.fields.sw,
				nextStep:    tt.fields.nextStep,
				hasNextStep: tt.fields.hasNextStep,
			}
		})
	}
}

func TestSplitterState_Start(t *testing.T) {
	type fields struct {
		id          pipeline.ElementId
		ts          TaskSplitter
		maxWSize    pipeline.TaskIndex
		sw          *pipeline.SlidingWindow
		nextStep    pipeline.Runner
		hasNextStep bool
	}
	type args struct {
		wg *sync.WaitGroup
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &SplitterState{
				id:          tt.fields.id,
				ts:          tt.fields.ts,
				maxWSize:    tt.fields.maxWSize,
				sw:          tt.fields.sw,
				nextStep:    tt.fields.nextStep,
				hasNextStep: tt.fields.hasNextStep,
			}
		})
	}
}

func TestSplitterState_Stop(t *testing.T) {
	type fields struct {
		id          pipeline.ElementId
		ts          TaskSplitter
		maxWSize    pipeline.TaskIndex
		sw          *pipeline.SlidingWindow
		nextStep    pipeline.Runner
		hasNextStep bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &SplitterState{
				id:          tt.fields.id,
				ts:          tt.fields.ts,
				maxWSize:    tt.fields.maxWSize,
				sw:          tt.fields.sw,
				nextStep:    tt.fields.nextStep,
				hasNextStep: tt.fields.hasNextStep,
			}
		})
	}
}

func TestSplitterState_UpdateTask(t *testing.T) {
	type fields struct {
		id          pipeline.ElementId
		ts          TaskSplitter
		maxWSize    pipeline.TaskIndex
		sw          *pipeline.SlidingWindow
		nextStep    pipeline.Runner
		hasNextStep bool
	}
	type args struct {
		sMsg *pipeline.StatusMessage
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &SplitterState{
				id:          tt.fields.id,
				ts:          tt.fields.ts,
				maxWSize:    tt.fields.maxWSize,
				sw:          tt.fields.sw,
				nextStep:    tt.fields.nextStep,
				hasNextStep: tt.fields.hasNextStep,
			}
			if got := st.UpdateTask(tt.args.sMsg); got != tt.want {
				t.Errorf("UpdateTask() = %v, want %v", got, tt.want)
			}
		})
	}
}