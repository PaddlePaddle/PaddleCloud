package main

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/fields"
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
	complete
)

type event struct {
	Type eventType
	//Job  *paddlejob.TrainingJob
}
type Cleaner struct {
	client    *rest.RESTClient
	clientset *kubernetes.Clientset
	ticker    *time.Ticker
	eventCh   chan event
	jobs      map[string]string
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

func NewCleaner(c *rest.RESTClient, cs *kubernetes.Clientset) *Cleaner {
	return &Cleaner{
		client:    c,
		clientset: cs,
		ticker:    time.NewTicker(time.Second * 5),
		eventCh:   make(chan event),
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

// OnAdd notifies the autoscaler that a job has been added.
func (c *Cleaner) onAdd(obj interface{}) {
	c.eventCh <- event{Type: add}
}

// OnDel notifies the autoscaler that a job has been deleted.
func (c *Cleaner) onDel(obj interface{}) {
	c.eventCh <- event{Type: del}
}

// Monitor monitors the cluster resources and training jobs in a loop,
// scales the training jobs according to the cluster resource.
func (c *Cleaner) Monitor() {
	for {
		select {
		case <-c.ticker.C:
		case e := <-c.eventCh:
			log.Debug("monitor received event", "event", e)
			switch e.Type {
			case add:
				//var tj *batchv1.Job
				//var err error
				//c.jobs[e.Job.ObjectMeta.Name] = e.Job.ObjectMeta.Name
				fmt.Println("add %v", e)
			case del:
				// TODO(helin): delete all created
				// resources (e.g., trainer Job,
				// pserver Replica Set) when we
				// schedules the resources.
				//delete(a.jobs, e.Job.ObjectMeta.Name)
				fmt.Println("delete %v", e)
			//case complete:?
			default:
				log.Error("unrecognized event", "event", e)
			}
		}
	}
}
