package main

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	batchv1 "k8s.io/client-go/pkg/apis/batch/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	log "github.com/inconshreveable/log15"
)

type eventType int

const (
	add eventType = iota
	del
	update
)

type event struct {
	Type eventType
	Job  interface{}
}

// Cleaner is a struct to clean pserver.
type Cleaner struct {
	client    *rest.RESTClient
	clientset *kubernetes.Clientset
	ticker    *time.Ticker
	eventCh   chan event
	jobs      map[types.UID]*batchv1.Job
}

/*
// GetTrainerJob gets the trainer job spec.
func (c *Cleaner) getTrainerJob(job *paddlejob.TrainingJob) (*batchv1.Job, error) {
	namespace := job.ObjectMeta.Namespace
	jobname := job.ObjectMeta.Name
	return c.clientset.
		BatchV1().
		Jobs(namespace).
		Get(fmt.Sprintf("%s-trainer", jobname), metav1.GetOptions{})
}
*/

// NewCleaner gets cleaner struct.
func NewCleaner(c *rest.RESTClient, cs *kubernetes.Clientset) *Cleaner {
	return &Cleaner{
		client:    c,
		clientset: cs,
		ticker:    time.NewTicker(time.Second * 5),
		eventCh:   make(chan event),
		jobs:      make(map[types.UID]*batchv1.Job),
	}
}

func (c *Cleaner) startWatch(ctx context.Context) error {
	source := cache.NewListWatchFromClient(
		c.client,
		"Jobs",
		api.NamespaceAll,
		fields.Everything())

	_, informer := cache.NewInformer(
		source,
		&batchv1.Job{},

		// resyncPeriod: Every resyncPeriod, all resources in
		// the cache will retrigger events. Set to 0 to
		// disable the resync.
		0,

		// TrainingJob custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDel,
		})

	go informer.Run(ctx.Done())
	return nil
}

// Run start to watch kubernetes events and do handlers.
func (c *Cleaner) Run(ctx context.Context) error {
	err := c.startWatch(ctx)
	if err != nil {
		return err
	}

	go c.Monitor()

	<-ctx.Done()
	return ctx.Err()
}

func (c *Cleaner) onAdd(obj interface{}) {
	c.eventCh <- event{Type: add, Job: obj}
}

func (c *Cleaner) onDel(obj interface{}) {
	c.eventCh <- event{Type: del, Job: obj}
}

func (c *Cleaner) onUpdate(oldObj, newObj interface{}) {
	c.eventCh <- event{Type: update, Job: newObj}
}

func getTrainerJobName(j *batchv1.Job) string {
	m := j.ObjectMeta.Labels
	if val, ok := m["paddle-job"]; ok {
		return val
	}

	return ""
}

func (c *Cleaner) delPserver(j *batchv1.Job) {
	// del resource
	// del pod
}

func (c *Cleaner) delMaster(j *batchv1.Job) {
	// del resource
	// del pod
}

func (c *Cleaner) clean(j *batchv1.Job) {
	jobname := getTrainerJobName(j)
	if jobname == "" {
		return
	}

	c.delPserver(j)
	c.delMaster(j)
}

// Monitor monitors the cluster paddle-job resource.
func (c *Cleaner) Monitor() {
	for {
		select {
		case <-c.ticker.C:
		case e := <-c.eventCh:
			switch e.Type {
			case add:
				j := e.Job.(*batchv1.Job)
				log.Info(fmt.Sprintf("add jobs namespace:%v name:%v uid:%v",
					j.ObjectMeta.Namespace, j.ObjectMeta.Name, j.ObjectMeta.UID))
				c.jobs[j.UID] = j
			case update: // process only complation
				j := e.Job.(*batchv1.Job)

				// not complete
				if j.Status.CompletionTime == nil {
					break
				}

				// completed already
				if _, ok := c.jobs[j.UID]; !ok {
					break
				}

				log.Info(fmt.Sprintf("complete jobs namespace:%v name:%v uid:%v",
					j.ObjectMeta.Namespace, j.ObjectMeta.Name, j.ObjectMeta.UID))

				c.clean(j)
				delete(c.jobs, j.UID)
			case del:
				j := e.Job.(*batchv1.Job)
				log.Info(fmt.Sprintf("delete jobs namespace:%v name:%v uid:%v",
					j.ObjectMeta.Namespace, j.ObjectMeta.Name, j.ObjectMeta.UID))

				// deleted already
				if _, ok := c.jobs[j.UID]; !ok {
					break
				}

				c.clean(j)
				delete(c.jobs, j.UID)
			default:
				log.Error("unrecognized event", "event", e)
			}
		}
	}
}
