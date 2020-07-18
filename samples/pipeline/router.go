package main

import (
	"fmt"
	"github.com/cloudevents/sdk-go/v2/pipeline"
	"github.com/cloudevents/sdk-go/v2/pipeline/config"
	"github.com/cloudevents/sdk-go/v2/pipeline/elements"
	"math/rand"
	"sync"
)

var _ pipeline.TaskSplitter = (*RandomRouter)(nil)

type RandomRouter struct {
	senders []pipeline.Runner
}

func (d *RandomRouter) StartRunners(wg *sync.WaitGroup) {
	for _, r := range d.senders {
		r.Start(wg)
	}
}

func (d *RandomRouter) StopRunners() {
	for _, r := range d.senders {
		r.Stop()
	}
}

func (d *RandomRouter) Split(origin *pipeline.Task, callback chan *pipeline.TaskStatus) []*pipeline.TaskAssignment {
	tas := make([]*pipeline.TaskAssignment, len(d.senders))
	i := 0
	for _, w := range d.senders {
		if rand.Intn(100) >= 50 {
			tas[i] = &pipeline.TaskAssignment{
				Task:   pipeline.NewTask(origin, callback),
				Runner: w,
			}
			i++
		}
	}
	if i == 0 {
		return tas[0:0]
	}
	return tas[0 : i-1]
}

func NewRandomRouter(senders []*elements.HttpSender, cfg *config.Config) *RandomRouter {
	dr := &RandomRouter{
		senders: make([]pipeline.Runner, len(senders)),
	}

	for i, s := range senders {
		id := fmt.Sprintf("sender %i", i)
		dr.senders[i] = pipeline.NewWorker(s,cfg,&id)
	}

	return dr
}
