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

package edl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := newAutoscaler(nil)
	assert.NotNil(t, c)
}

func TestMonitor(t *testing.T) {
	c := newAutoscaler(nil)
	ch := make(chan struct{})

	go func() {
		c.Monitor()
		close(ch)
	}()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-ch:
		t.Fatal("monitor should be blocked")
	default:
	}
}
