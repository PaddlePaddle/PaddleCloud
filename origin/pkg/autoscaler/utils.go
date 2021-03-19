/* Copyright (c) 2016 PaddlePaddle Authors All Rights Reserved.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
	 limitations under the License.*/

package autoscaler

import v1 "k8s.io/api/core/v1"

// AddResourceList add another v1.ResourceList to first's inner
// quantity.  v1.ResourceList is equal to map[string]Quantity
func AddResourceList(a v1.ResourceList, b v1.ResourceList) {
	for resname, q := range b {
		v, _ := a[resname]
		v.Add(q)
		a[resname] = v
	}

	return
}
