package pipeline

import "sync"

type StartStop interface {
	Start(wg *sync.WaitGroup) error
	Stop()
}
