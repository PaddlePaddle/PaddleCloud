package operator_test

import (
	"testing"
	"time"

	"github.com/PaddlePaddle/cloud/go/controller"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := operator.New(nil)
	assert.NotNil(t, c)
}

type myFreeResource struct {
}

func (f *myFreeResource) GPU() int {
	panic("not implemented")
}

func (f *myFreeResource) CPU() float64 {
	panic("not implemented")
}

func (f *myFreeResource) Mem() float64 {
	panic("not implemented")
}

func TestMonitor(t *testing.T) {
	c := operator.New(&myFreeResource{})
	ch := make(chan struct{})

	go func() {
		c.Monitor(nil)
		close(ch)
	}()

	time.Sleep(10 * time.Millisecond)

	select {
	case <-ch:
		t.Fatal("monitor should be blocked")
	default:
	}
}
