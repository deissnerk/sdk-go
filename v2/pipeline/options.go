package pipeline

type SupervisorOption func(*Supervisor) error

/*func WithWindowSize(ws TaskIndex) SupervisorOption {
	return func(sv *Supervisor) error {
		sv.sw = NewSlidingWindow(ws)
		return nil
	}
}
*/
func WithSplitter(ts TaskSplitter, maxWSize TaskIndex) SupervisorOption {
	return func(sv *Supervisor) error {
		sv.state = &SplitterState{
			ts:       ts,
			maxWSize: maxWSize,
			wsc:      sv,
			sw:       NewSlidingWindow(maxWSize),
		}
		return nil
	}
}
