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

// Node represent minimum cell to schedule.
type Node interface {
	// the mixed weight of current node.
	Weight() float64

	// get the object to be scheduled.
	Obj() *interface{}
}

// WeightedAccelleratorCFS is a scheduler to schedule jobs/processes to use
// multiple kind of processers, like both CPU and GPU, or mix with FPGA etc.
type WeightedAccelleratorCFS interface {
	AddNode(node *Node) error
	DelNode(node *Node) error

	// tranverse all the nodes and call the callback function.
	Tranverse(callback ...func(*Node)) error

	// Get one node that ready to schedule.
	Get() *Node
}
