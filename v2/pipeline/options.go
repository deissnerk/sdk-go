package pipeline

type StateOption func(*Supervisor) error

func WithWindowSize(ws uint32) StateOption {
	return func(sv *Supervisor) error {
		sv.maxWSize = ws
		return nil
	}
}
