package main

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// TODO(gongwb): filer only paddle-job
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

func cleanupReplicaSets(client *kubernetes.Clientset,
	namespace string, l metav1.ListOptions) error {
	rsList, err := client.ExtensionsV1beta1().ReplicaSets(namespace).List(l)
	if err != nil {
		return err
	}

	for _, rs := range rsList.Items {
		err := client.ExtensionsV1beta1().ReplicaSets(namespace).Delete(rs.ObjectMeta.Name, nil)
		if err != nil {
			log.Error(fmt.Sprintf("delete rs namespace:%v  rsname:%v err:%v", namespace, rs.Name, err))
		}

		log.Info(fmt.Sprintf("delete rs namespace:%v  rsname:%v", namespace, rs.Name))

	}
	return nil
}

func cleanupPods(client *kubernetes.Clientset,
	namespace string, l metav1.ListOptions) error {
	podList, err := client.CoreV1().Pods(namespace).List(l)
	if err != nil {
		return err
	}

	for _, pod := range podList.Items {
		err := client.CoreV1().Pods(namespace).Delete(pod.ObjectMeta.Name, nil)
		if err != nil {
			log.Error(fmt.Sprintf("delete pod namespace:%v  podname:%v err:%v", namespace, pod.Name, err))
		}

		log.Info(fmt.Sprintf("delete pod namespace:%v  podname:%v", namespace, pod.Name))
	}
	return nil
}

func (c *Cleaner) cleanupPserver(namespace, jobname string) {
	cleanupReplicaSets(c.clientset, namespace,
		metav1.ListOptions{LabelSelector: "paddle-job-pserver=" + jobname})
	log.Info(fmt.Sprintf("delete pserver replicaset namespace:%s jobname:%s", namespace, jobname))

	// wait to delete replicaset
	time.Sleep(1 * time.Second)

	cleanupPods(c.clientset, namespace,
		metav1.ListOptions{LabelSelector: "paddle-job-pserver=" + jobname})
	log.Info(fmt.Sprintf("delete pserver pods namespace:%s jobname:%s", namespace, jobname))
}

func (c *Cleaner) cleanup(j *batchv1.Job) {
	jobname := getTrainerJobName(j)
	if jobname == "" {
		return
	}

	c.cleanupPserver(j.ObjectMeta.Namespace, jobname)
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
				// get only paddle-job, it's not the best method.
				if getTrainerJobName(j) == "" {
					break
				}

				log.Info(fmt.Sprintf("add jobs namespace:%v name:%v uid:%v",
					j.ObjectMeta.Namespace, j.ObjectMeta.Name, j.ObjectMeta.UID))
				c.jobs[j.UID] = j
			case update: // process only complation
				j := e.Job.(*batchv1.Job)

				// not complete
				if j.Status.CompletionTime == nil {
					break
				}

				// if not in controll or completed already
				if _, ok := c.jobs[j.UID]; !ok {
					break
				}

				log.Info(fmt.Sprintf("complete jobs namespace:%v name:%v uid:%v",
					j.ObjectMeta.Namespace, j.ObjectMeta.Name, j.ObjectMeta.UID))

				c.cleanup(j)
				delete(c.jobs, j.UID)
			case del:
				j := e.Job.(*batchv1.Job)
				log.Info(fmt.Sprintf("delete jobs namespace:%v name:%v uid:%v",
					j.ObjectMeta.Namespace, j.ObjectMeta.Name, j.ObjectMeta.UID))

				// deleted already
				if _, ok := c.jobs[j.UID]; !ok {
					break
				}

				c.cleanup(j)
				delete(c.jobs, j.UID)
			default:
				log.Error("unrecognized event", "event", e)
			}
		}
	}
}
