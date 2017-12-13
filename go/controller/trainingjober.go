/* Copyright (c) 2016 PaddlePaddle Authors All Rights Reserve.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
	 limitations under the License. */

package controller

import (
	"fmt"
	"time"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	log "github.com/inconshreveable/log15"
)

const (
	defaultLoopNum = 3
	defaultDur     = 1 * time.Second
)

// TrainingJober mananges TraingJobs.
type TrainingJober struct {
	cluster *Cluster
}

// NewTrainingJober create a TrainingJober.
func NewTrainingJober(c *Cluster) *TrainingJober {
	return &TrainingJober{
		cluster: c,
	}
}

func (c *TrainingJober) cleanupPserver(namespace, jobname string) error {
	name := jobname + "-pserver"
	err := c.cluster.DeleteReplicaSet(namespace, name)
	if err != nil {
		return fmt.Errorf("delete pserver namespace:%s name:%s error:%v",
			namespace, name, err)
	}

	log.Error(fmt.Sprintf("delete pserver namespace:%s name:%s",
		namespace, name))
	return nil
}

func (c *TrainingJober) cleanupMaster(namespace, jobname string) error {
	name := jobname + "-master"
	err := c.cluster.DeleteReplicaSet(namespace, name)
	if err != nil {
		return fmt.Errorf("delete master namespace:%s name:%s error:%v",
			namespace, name, err)
	}

	log.Error(fmt.Sprintf("delete master namespace:%s name:%s",
		namespace, name))
	return nil
}

func (c *TrainingJober) cleanupTrainer(namespace, jobname string) error {
	name := jobname + "-trainer"
	err := c.cluster.DeleteTrainerJob(namespace, name)
	if err != nil {
		return fmt.Errorf("delete trainerjob namespace:%v name:%v error:%v",
			namespace, name, err)
	}

	log.Error(fmt.Sprintf("delete trainerjob namespace:%s name:%s",
		namespace, name))
	return nil
}

func (c *TrainingJober) createMaster(job *paddlejob.TrainingJob) error {
	var parser DefaultJobParser
	m := parser.ParseToMaster(job)

	_, err := c.cluster.CreateReplicaSet(m)
	if err != nil {
		e := fmt.Errorf("create master namespace:%v  name:%v error:%v",
			job.ObjectMeta.Namespace, job.ObjectMeta.Name, err)
		log.Error(e.Error())
		return e
	}

	return nil
}

func (c *TrainingJober) createPserver(job *paddlejob.TrainingJob) error {
	var parser DefaultJobParser
	p := parser.ParseToPserver(job)

	_, err := c.cluster.CreateReplicaSet(p)
	if err != nil {
		e := fmt.Errorf("create pserver namespace:%v  name:%v error:%v",
			job.ObjectMeta.Namespace, job.ObjectMeta.Name, err)
		log.Error(e.Error())
		return e
	}
	return nil
}

func (c *TrainingJober) createTrainer(job *paddlejob.TrainingJob) error {
	var parser DefaultJobParser
	t := parser.ParseToTrainer(job)

	_, err := c.cluster.CreateJob(t)
	if err != nil {
		e := fmt.Errorf("create trainerjob namespace:%v  name:%v error:%v",
			job.ObjectMeta.Namespace, job.ObjectMeta.Name, err)
		log.Error(e.Error())
		return e
	}

	return nil
}

// Complete clears master and pserver resources.
func (c *TrainingJober) Complete(job *paddlejob.TrainingJob) {
	c.cleanupPserver(job.ObjectMeta.Namespace,
		job.ObjectMeta.Name)

	c.cleanupMaster(job.ObjectMeta.Namespace,
		job.ObjectMeta.Name)
}

// Destroy destroys resource and pods.
func (c *TrainingJober) Destroy(job *paddlejob.TrainingJob) {
	c.Complete(job)

	c.cleanupTrainer(job.ObjectMeta.Namespace,
		job.ObjectMeta.Name)
}

func (c *TrainingJober) checkAndCreate(job *paddlejob.TrainingJob) error {
	tname := job.ObjectMeta.Name + "-trainer"
	mname := job.ObjectMeta.Name + "-master"
	pname := job.ObjectMeta.Name + "-pserver"
	namespace := job.ObjectMeta.Namespace

	t, terr := c.cluster.GetTrainerJobByName(namespace, tname)
	m, merr := c.cluster.GetReplicaSet(namespace, mname)
	p, perr := c.cluster.GetReplicaSet(namespace, pname)

	if terr != nil ||
		merr != nil ||
		perr != nil {
		err := fmt.Errorf("trainerjob_err:%v master_err:%v pserver_err:%v",
			terr, merr, perr)
		log.Error(err.Error())
		return err
	}

	if m == nil {
		if err := c.createMaster(job); err != nil {
			return fmt.Errorf("namespace:%v create master:%v error:%v",
				namespace, mname, err)
		}
	}

	if t == nil {
		if err := c.createTrainer(job); err != nil {
			if m == nil {
				c.cleanupMaster(namespace, mname)
			}
			return fmt.Errorf("namespace:%v create trainer:%v error:%v",
				namespace, tname, err)
		}
	}

	if p == nil {
		if m == nil {
			c.cleanupMaster(namespace, mname)
		}

		if t == nil {
			c.cleanupTrainer(namespace, tname)
		}

		if err := c.createPserver(job); err != nil {
			return fmt.Errorf("namespace:%v create pserver:%v error:%v",
				namespace, pname, err)
		}
	}

	return nil
}

// Ensure try to make sure trainer, pserver, master exists.
func (c *TrainingJober) Ensure(job *paddlejob.TrainingJob) error {
	var err error
	for i := 0; i < defaultLoopNum; i++ {
		err = c.checkAndCreate(job)
		if err == nil {
			return nil
		}
		time.Sleep(defaultDur)
	}

	return err
}
