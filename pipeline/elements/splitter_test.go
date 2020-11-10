package elements

import (
	"context"
	"github.com/cloudevents/sdk-go/pipeline"
	"reflect"
	"testing"
)

func TestNewSplitterRunner(t *testing.T) {
	type args struct {
		ts       Splitter
		id       pipeline.ElementId
		nextStep pipeline.Runner
	}
	tests := []struct {
		name    string
		args    args
		want    *SplitterRunner
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSplitterRunner(tt.args.ts, tt.args.id, tt.args.nextStep)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSplitterRunner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSplitterRunner() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitterRunner_Id(t *testing.T) {
	type fields struct {
		state *SplitterState
		sv    *pipeline.Supervisor
	}
	tests := []struct {
		name   string
		fields fields
		want   pipeline.ElementId
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SplitterRunner{
				state: tt.fields.state,
				sv:    tt.fields.sv,
			}
			if got := s.Id(); got != tt.want {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitterRunner_Push(t *testing.T) {
	type fields struct {
		state *SplitterState
		sv    *pipeline.Supervisor
	}
	type args struct {
		tc *pipeline.TaskContainer
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
			s := &SplitterRunner{
				state: tt.fields.state,
				sv:    tt.fields.sv,
			}
			s.Id()
		})
	}
}

func TestSplitterRunner_Start(t *testing.T) {
	type fields struct {
		state *SplitterState
		sv    *pipeline.Supervisor
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := SplitterRunner{
				state: tt.fields.state,
				sv:    tt.fields.sv,
			}
			if err := s.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSplitterRunner_Stop(t *testing.T) {
	type fields struct {
		state *SplitterState
		sv    *pipeline.Supervisor
	}
	type args struct {
		ctx context.Context
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
			s := SplitterRunner{
				state: tt.fields.state,
				sv:    tt.fields.sv,
			}
			s.Id()
		})
	}
}

func TestSplitterState_AddTask(t *testing.T) {
	type fields struct {
		id       pipeline.ElementId
		ts       Splitter
		sw       *pipeline.SlidingWindow
		nextStep pipeline.Runner
	}
	type args struct {
		tc       *pipeline.TaskContainer
		callback chan *pipeline.StatusMessage
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
				id:       tt.fields.id,
				ts:       tt.fields.ts,
				sw:       tt.fields.sw,
				nextStep: tt.fields.nextStep,
			}
			st.Id()
		})
	}
}

func TestSplitterState_Id(t *testing.T) {
	type fields struct {
		id       pipeline.ElementId
		ts       Splitter
		sw       *pipeline.SlidingWindow
		nextStep pipeline.Runner
	}
	tests := []struct {
		name   string
		fields fields
		want   pipeline.ElementId
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &SplitterState{
				id:       tt.fields.id,
				ts:       tt.fields.ts,
				sw:       tt.fields.sw,
				nextStep: tt.fields.nextStep,
			}
			if got := st.Id(); got != tt.want {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitterState_IsIdle(t *testing.T) {
	type fields struct {
		id       pipeline.ElementId
		ts       Splitter
		sw       *pipeline.SlidingWindow
		nextStep pipeline.Runner
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
				id:       tt.fields.id,
				ts:       tt.fields.ts,
				sw:       tt.fields.sw,
				nextStep: tt.fields.nextStep,
			}
			if got := st.IsIdle(); got != tt.want {
				t.Errorf("IsIdle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSplitterState_Start(t *testing.T) {
	type fields struct {
		id       pipeline.ElementId
		ts       Splitter
		sw       *pipeline.SlidingWindow
		nextStep pipeline.Runner
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &SplitterState{
				id:       tt.fields.id,
				ts:       tt.fields.ts,
				sw:       tt.fields.sw,
				nextStep: tt.fields.nextStep,
			}
			if err := st.Start(); (err != nil) != tt.wantErr {
				t.Errorf("Start() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSplitterState_Stop(t *testing.T) {
	type fields struct {
		id       pipeline.ElementId
		ts       Splitter
		sw       *pipeline.SlidingWindow
		nextStep pipeline.Runner
	}
	type args struct {
		ctx context.Context
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
				id:       tt.fields.id,
				ts:       tt.fields.ts,
				sw:       tt.fields.sw,
				nextStep: tt.fields.nextStep,
			}
			st.Id()
		})
	}
}

func TestSplitterState_UpdateTask(t *testing.T) {
	type fields struct {
		id       pipeline.ElementId
		ts       Splitter
		sw       *pipeline.SlidingWindow
		nextStep pipeline.Runner
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
				id:       tt.fields.id,
				ts:       tt.fields.ts,
				sw:       tt.fields.sw,
				nextStep: tt.fields.nextStep,
			}
			if got := st.UpdateTask(tt.args.sMsg); got != tt.want {
				t.Errorf("UpdateTask() = %v, want %v", got, tt.want)
			}
		})
	}
}