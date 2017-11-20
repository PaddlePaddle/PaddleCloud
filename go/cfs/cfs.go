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

package cfs

// We can import a RB Tree implementation from here in initial
// implement, and refine it later

// import (
//   "github.com/sakeven/RbTree"
// )

// PrioLevel is enum type indicating priority levels.
type PrioLevel int

const (
	// Experiement priority level
	Experiement PrioLevel = 0
	// Offline job
	Offline = 3
	// Normal jobs
	Normal = 7
	// Production level job
	Production = 11
)

// Node is the atimic schedule unit for the scheduler.
type Node interface {
	// GetPrio returns the current priority level.
	GetPrio() PrioLevel
	// SetPrio set the node priority level directly.
	SetPrio(prio PrioLevel)

	// MaxInstances returns the desired max parallelism of the job.
	MaxInstances() int
	// MinInstances returns the minimal parallelism the job can be running.
	MinInstances() int
	// ResourceScore returns resource score of a single pod. It's
	// caculated by sum(weight*ResourceValue).
	ResourceScore() int64

	// Expected returns expected parallelism (how much pods) to run for
	// current scheduling step.
	Expected() int64
	// Running returns the current parrallelism of the node.
	// If Running == 0 means the job is waiting for resources.
	Running() int64

	// Obj returns inner scheduling unit.
	Obj() *interface{}
}

// GpuPriorityCFS is a scheduler to schedule jobs/processes to use
// multiple kind of processers, like both CPU and GPU, or mix with FPGA etc.
type GpuPriorityCFS interface {
	// AddNode insert new node to the scheduler.
	AddNode(node *Node) error
	// DelNode remove the completed node from scheduler.
	DelNode(node *Node) error
	// GetLeftMost return the smallest valued node in the scheduler's tree.
	GetLeftMost() *Node
	// GetRightMost return the maximum valued node in the scheduler's tree.
	GetRightMost() *Node
	// Len return number of nodes in the scheduler.
	Len() int

	// Tranverse go thought every nodes in the scheduler.
	Tranverse(callback ...func(*Node)) error
}
