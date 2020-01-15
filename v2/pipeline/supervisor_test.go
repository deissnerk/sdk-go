package pipeline

import (
	"github.com/cloudevents/sdk-go/pkg/pipeline/config"
	"reflect"
	"sync"
	"testing"
)

func TestNewSupervisor(t *testing.T) {
	type args struct {
		tg  TaskSplitter
		cfg *config.Config
		id  ElementId
	}
	tests := []struct {
		name string
		args args
		want *Supervisor
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSupervisor(tt.args.tg, tt.args.cfg, tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSupervisor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupervisor_Id(t *testing.T) {
	type fields struct {
		id       ElementId
		q        chan *TaskRef
		status   chan *TaskStatus
		stop     chan bool
		maxWSize uint32
		tg       TaskSplitter
	}
	tests := []struct {
		name   string
		fields fields
		want   ElementId
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				id:       tt.fields.id,
				q:        tt.fields.q,
				status:   tt.fields.status,
				stop:     tt.fields.stop,
				maxWSize: tt.fields.maxWSize,
				tg:       tt.fields.tg,
			}
			if got := s.Id(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Id() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupervisor_Push(t *testing.T) {
	type fields struct {
		id       ElementId
		q        chan *TaskRef
		status   chan *TaskStatus
		stop     chan bool
		maxWSize uint32
		tg       TaskSplitter
	}
	type args struct {
		tr *TaskRef
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
			s := &Supervisor{
				id:       tt.fields.id,
				q:        tt.fields.q,
				status:   tt.fields.status,
				stop:     tt.fields.stop,
				maxWSize: tt.fields.maxWSize,
				tg:       tt.fields.tg,
			}
		})
	}
}

func TestSupervisor_Start(t *testing.T) {
	type fields struct {
		id       ElementId
		q        chan *TaskRef
		status   chan *TaskStatus
		stop     chan bool
		maxWSize uint32
		tg       TaskSplitter
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
			s := &Supervisor{
				id:       tt.fields.id,
				q:        tt.fields.q,
				status:   tt.fields.status,
				stop:     tt.fields.stop,
				maxWSize: tt.fields.maxWSize,
				tg:       tt.fields.tg,
			}
		})
	}
}

func TestSupervisor_Stop(t *testing.T) {
	type fields struct {
		id       ElementId
		q        chan *TaskRef
		status   chan *TaskStatus
		stop     chan bool
		maxWSize uint32
		tg       TaskSplitter
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Supervisor{
				id:       tt.fields.id,
				q:        tt.fields.q,
				status:   tt.fields.status,
				stop:     tt.fields.stop,
				maxWSize: tt.fields.maxWSize,
				tg:       tt.fields.tg,
			}
		})
	}
}

func Test_initialStatus(t *testing.T) {
	tests := []struct {
		name string
		want TaskResult
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := initialStatus(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initialStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_svWindowState_addTask(t *testing.T) {
	type fields struct {
		tBuffer []*SuperVisorTask
		wStart  uint32
		wEnd    uint32
		sv      *Supervisor
		cond    *sync.Cond
	}
	type args struct {
		tr *TaskRef
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
			st := &svWindowState{
				tBuffer: tt.fields.tBuffer,
				wStart:  tt.fields.wStart,
				wEnd:    tt.fields.wEnd,
				sv:      tt.fields.sv,
				cond:    tt.fields.cond,
			}
			if got := st.AddTask(tt.args.tr); got != tt.want {
				t.Errorf("addTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_svWindowState_aggregateResult(t *testing.T) {
	type fields struct {
		tBuffer []*SuperVisorTask
		wStart  uint32
		wEnd    uint32
		sv      *Supervisor
		cond    *sync.Cond
	}
	type args struct {
		ts []*TaskStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   TaskResult
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &svWindowState{
				tBuffer: tt.fields.tBuffer,
				wStart:  tt.fields.wStart,
				wEnd:    tt.fields.wEnd,
				sv:      tt.fields.sv,
				cond:    tt.fields.cond,
			}
			if got := st.aggregateResult(tt.args.ts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("aggregateResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_svWindowState_removeTask(t *testing.T) {
	type fields struct {
		tBuffer []*SuperVisorTask
		wStart  uint32
		wEnd    uint32
		sv      *Supervisor
		cond    *sync.Cond
	}
	type args struct {
		id TaskId
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
			st := &svWindowState{
				tBuffer: tt.fields.tBuffer,
				wStart:  tt.fields.wStart,
				wEnd:    tt.fields.wEnd,
				sv:      tt.fields.sv,
				cond:    tt.fields.cond,
			}
		})
	}
}

func Test_svWindowState_updateTask(t *testing.T) {
	type fields struct {
		tBuffer []*SuperVisorTask
		wStart  uint32
		wEnd    uint32
		sv      *Supervisor
		cond    *sync.Cond
	}
	type args struct {
		sMsg *TaskStatus
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
			st := &svWindowState{
				tBuffer: tt.fields.tBuffer,
				wStart:  tt.fields.wStart,
				wEnd:    tt.fields.wEnd,
				sv:      tt.fields.sv,
				cond:    tt.fields.cond,
			}
			if got := st.UpdateTask(tt.args.sMsg); got != tt.want {
				t.Errorf("updateTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_svWindowState_wSize(t *testing.T) {
	type fields struct {
		tBuffer []*SuperVisorTask
		wStart  uint32
		wEnd    uint32
		sv      *Supervisor
		cond    *sync.Cond
	}
	tests := []struct {
		name   string
		fields fields
		want   uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &svWindowState{
				tBuffer: tt.fields.tBuffer,
				wStart:  tt.fields.wStart,
				wEnd:    tt.fields.wEnd,
				sv:      tt.fields.sv,
				cond:    tt.fields.cond,
			}
			if got := st.wSize(); got != tt.want {
				t.Errorf("wSize() = %v, want %v", got, tt.want)
			}
		})
	}
}