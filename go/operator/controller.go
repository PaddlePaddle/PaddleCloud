package operator

import (
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultLoopDur = time.Second
)

// FreeResource indicates the free resource available.
type FreeResource interface {
	GPU() int
	CPU() float64
	Mem() float64
}

// EventType is the type of the spec event.
type EventType int

// Spec event types
const (
	Add EventType = iota
	Delete
)

// ConfigEvent is an event happened to the specs.
type ConfigEvent struct {
	Type   EventType
	Config Config
}

type job struct {
	Config      Config
	CurInstance int
}

func (j job) Fulfullment() float64 {
	minInstance := j.Config.Spec.Trainer.MinInstance
	maxInstance := j.Config.Spec.Trainer.MaxInstance

	if minInstance == maxInstance {
		return 1
	}

	curInstance := j.CurInstance
	return float64(curInstance-minInstance) / float64(maxInstance-minInstance)
}

// Controller controls a training job.
type Controller struct {
	ticker *time.Ticker
	free   FreeResource
	jobs   map[string]job
}

// New creates a new controller.
func New(free FreeResource, options ...func(*Controller)) *Controller {
	c := &Controller{
		free:   free,
		ticker: time.NewTicker(defaultLoopDur),
		jobs:   make(map[string]job),
	}
	for _, option := range options {
		option(c)
	}
	return c
}

type jobs []job

func (j jobs) Len() int {
	return len(j)
}

func (j jobs) Less(a int, b int) bool {
	scoreA := j[a].Fulfullment()
	scoreB := j[b].Fulfullment()

	if scoreA == scoreB {
		resA := j[a].Config.Spec.Trainer.Resources
		resB := j[b].Config.Spec.Trainer.Resources
		if resA.Limits.GPU == resB.Limits.GPU {
			if resA.Requests.CPU == resB.Requests.CPU {
				return resA.Requests.Mem < resB.Requests.Mem
			}
			return resA.Requests.CPU < resB.Requests.CPU
		}
		return resA.Limits.GPU < resB.Limits.GPU
	}
	return scoreA < scoreB
}

func (j jobs) Swap(a int, b int) {
	j[a], j[b] = j[b], j[a]
}

// sortedElasticJobs return the names of sorted jobs by fulfillment
// and tiebreakers in ascending order.
func (c *Controller) sortedElasticJobs() []string {
	var js jobs
	for _, v := range c.jobs {
		if !v.Config.Elastic() {
			continue
		}

		js = append(js, v)
	}
	sort.Sort(js)
	var result []string
	for _, v := range js {
		result = append(result, v.Config.MetaData.Name)
	}
	return result
}

func (c *Controller) dynamicScaling() {
	// TODO(helin)
}

// Monitor schedules and scales the training jobs.
func (c *Controller) Monitor(event <-chan ConfigEvent) {
	for {
		select {
		case <-c.ticker.C:
		case e := <-event:
			switch e.Type {
			case Add:
				log.Debugf("Add spec: %s", e.Config.MetaData.Name)
			case Delete:
				log.Debugf("Delete spec: %s", e.Config.MetaData.Name)
			}
		}
		c.dynamicScaling()
	}
}
