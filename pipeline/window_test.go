package pipeline

import (
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func TestSlidingWindow_AddTask(t *testing.T) {
	type fields struct {
		tBuf     []*SuperVisorTask
		wStart   TaskIndex
		wEnd     TaskIndex
		cond     *sync.Cond
		maxWSize TaskIndex
	}
	type args struct {
		svt *SuperVisorTask
	}
	type wants struct {
		wStart TaskIndex
		wEnd   TaskIndex
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{name: "Reset wEnd to 0",
			fields: fields{
				tBuf:     make([]*SuperVisorTask, 10),
				wStart:   1,
				wEnd:     9,
				cond:     sync.NewCond(&sync.Mutex{}),
				maxWSize: 10,
			},
			args: args{&SuperVisorTask{}},
			wants: wants{
				wStart: 1,
				wEnd:   0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := &SlidingWindow{
				tBuf:     tt.fields.tBuf,
				wStart:   tt.fields.wStart,
				wEnd:     tt.fields.wEnd,
				cond:     tt.fields.cond,
				maxWSize: tt.fields.maxWSize,
			}
			sw.AddTask()
			require.Equal(t, tt.wants.wStart, sw.wStart)
			require.Equal(t, tt.wants.wEnd, sw.wEnd)
		})
	}
}

func TestSlidingWindow_RemoveTask(t *testing.T) {

	result := ""

	dummyTask := &SuperVisorTask{
		Main:  &TaskContainer{},
		TStat: nil,
	}
	tBuf := []*SuperVisorTask{
		nil,
		dummyTask,
		nil,
		dummyTask,
		dummyTask,
		nil,
	}
	finalizers := []func(){
		nil,
		nil,
		func() { result = result + "2" },
		nil,
		nil,
		nil,
	}

	type fields struct {
		tBuf       []*SuperVisorTask
		finalizers []func()
		wStart     TaskIndex
		wEnd       TaskIndex
		cond       *sync.Cond
		maxWSize   TaskIndex
	}
	type args struct {
		ti        TaskIndex
		finalizer func()
	}
	type wants struct {
		result string
		empty  bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{name: "Reset wEnd to 0",
			fields: fields{
				tBuf:       tBuf,
				finalizers: finalizers,
				wStart:     1,
				wEnd:       5,
				cond:       sync.NewCond(&sync.Mutex{}),
				maxWSize:   6,
			},
			args: args{1, func() { result = result + "1" }},
			wants: wants{
				result: "12",
				empty:  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := &SlidingWindow{
				tBuf:      tt.fields.tBuf,
				finalizer: tt.fields.finalizers,
				wStart:    tt.fields.wStart,
				wEnd:      tt.fields.wEnd,
				cond:      tt.fields.cond,
				maxWSize:  tt.fields.maxWSize,
			}
			if empty := sw.RemoveTask(tt.args.ti, tt.args.finalizer);
				empty != tt.wants.empty && result != tt.wants.result {
				t.Errorf("RemoveTask() did not work")
			}
		})
	}
}

func TestSlidingWindow_wSize(t *testing.T) {
	type fields struct {
		tBuf     []*SuperVisorTask
		wStart   TaskIndex
		wEnd     TaskIndex
		cond     *sync.Cond
		maxWSize TaskIndex
	}
	tests := []struct {
		name   string
		fields fields
		want   TaskIndex
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := &SlidingWindow{
				tBuf:     tt.fields.tBuf,
				wStart:   tt.fields.wStart,
				wEnd:     tt.fields.wEnd,
				cond:     tt.fields.cond,
				maxWSize: tt.fields.maxWSize,
			}
			if got := sw.wSize(); got != tt.want {
				t.Errorf("wSize() = %v, want %v", got, tt.want)
			}
		})
	}
}
