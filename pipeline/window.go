package pipeline

import "sync"

type SlidingWindow struct {
	tBuf      []*SuperVisorTask
	wStart    TaskIndex // The start of the sliding window is the index of the oldest entry in the window
	wEnd      TaskIndex // The end of the sliding window is, where the next entry is added
	cond      *sync.Cond
	maxWSize  TaskIndex
	finalizer []func()
}

func (sw *SlidingWindow) AddTask() (*SuperVisorTask, TaskIndex) {
	sw.cond.L.Lock()
	defer sw.cond.L.Unlock()

	// If the window is full, wait until a slot gets free
	for sw.wSize() >= sw.maxWSize {
		sw.cond.Wait()
	}

	svt := &SuperVisorTask{}
	ti := sw.wEnd

	// As wEnd always points to the next free slot, add the svt here...
	sw.tBuf[ti] = svt

	// ... and increase wEnd afterwards
	if sw.wEnd < sw.maxWSize-1 {
		sw.wEnd++
	} else {
		sw.wEnd = 0
	}
	return svt, ti
}

// RemoveTask removes the task with index ti from the window.
// It returns true, if the window is empty afterwards.
// The finalizers are executed in the order of the tasks
func (sw *SlidingWindow) RemoveTask(ti TaskIndex,finalizer func()) bool {
	defer sw.cond.L.Unlock()
	sw.cond.L.Lock()
	sw.tBuf[ti] = nil
	sw.finalizer[ti] = finalizer

	for i := sw.wStart; sw.tBuf[sw.wStart] == nil && i < sw.maxWSize && sw.wStart != sw.wEnd; i++ {
		sw.finalizer[sw.wStart]()
		sw.finalizer[sw.wStart] = nil
		if sw.wStart < sw.maxWSize-1 {
			sw.wStart++
		} else {
			sw.wStart = 0
		}
	}

	// In case someone is waiting for the window to get free again, send a signal
	sw.cond.Signal()
	if sw.wSize() == 0 {
		return true
	}
	return false
}

func (sw *SlidingWindow) GetSupervisorTask(i TaskIndex) *SuperVisorTask {
	return sw.tBuf[i]
}

// IsEmpty returns true, if the sliding window is empty
// It should not be called frequently, as it has to lock the state for
// other operations
func (sw *SlidingWindow) IsEmpty() bool {
	defer sw.cond.L.Unlock()
	sw.cond.L.Lock()
	e := sw.wSize() == 0
	return e
}

func NewSlidingWindow(maxWSize TaskIndex) *SlidingWindow {
	return &SlidingWindow{
		tBuf:      make([]*SuperVisorTask, maxWSize),
		finalizer: make([]func(), maxWSize),
		wStart:    0,
		wEnd:      0,
		cond:      sync.NewCond(&sync.Mutex{}),
		maxWSize:  maxWSize,
	}
}

// This is not thread safe. The caller needs to lock SlidingWindow.cond.L
func (sw *SlidingWindow) wSize() TaskIndex {
	if sw.wStart <= sw.wEnd {
		return sw.wEnd - sw.wStart
	}
	return (sw.maxWSize - sw.wStart) + sw.wEnd
}