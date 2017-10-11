package operator

import (
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	defaultLoopDur = time.Second
)

// Cluster represents the cluster managment system such as Kubernetes.
type Cluster interface {
	// Free resources, must reflect the resources consumed by the
	// jobs created by SubmitJob that are still pending, or the
	// resource release by the job deletion by DeleteJob that are
	// still pending.
	FreeGPU() int
	FreeCPU() float64
	FreeMem() float64

	SubmitJob(Config) error
	DeleteJob(Config) error
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
	ticker  *time.Ticker
	cluster Cluster
	jobs    map[string]job
}

// New creates a new controller.
func New(cluster Cluster, options ...func(*Controller)) *Controller {
	c := &Controller{
		cluster: cluster,
		ticker:  time.NewTicker(defaultLoopDur),
		jobs:    make(map[string]job),
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

// elastic job filter.
func elastic(j job) bool {
	return j.Config.Elastic()
}

// gpu job filter.
func gpu(j job) bool {
	return j.Config.Spec.Trainer.Resources.Limits.GPU > 0
}

// sortedElasticJobs return the names of sorted jobs by fulfillment
// and tiebreakers in ascending order.
func (c *Controller) sortedJobs(filters ...func(job) bool) []string {
	var js jobs
nextJob:
	for _, v := range c.jobs {
		for _, f := range filters {
			if !f(v) {
				continue nextJob
			}
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
				log.Debugf("Add config: %s", e.Config.MetaData.Name)
				_, ok := c.jobs[e.Config.MetaData.Name]
				if ok {
					log.Errorf("The config %s to add already exists.", e.Config.MetaData.Name)
					continue
				}
				c.jobs[e.Config.MetaData.Name] = job{Config: e.Config}
			case Delete:
				log.Debugf("Delete config: %s", e.Config.MetaData.Name)
				j, ok := c.jobs[e.Config.MetaData.Name]
				if !ok {
					log.Errorf("Could not find the config %s to delete.", e.Config.MetaData.Name)
					continue
				}
				delete(c.jobs, e.Config.MetaData.Name)
				go func(j job) {
					err := c.cluster.DeleteJob(j.Config)
					if err != nil {
						log.Errorf("error on delete job: %v", err)
					}
				}(j)
			}
		}
		c.dynamicScaling()
	}
}
