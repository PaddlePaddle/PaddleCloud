package operator_test

import (
	"testing"

	"github.com/PaddlePaddle/cloud/go/controller"
	"github.com/stretchr/testify/assert"
)

func TestNeedGPU(t *testing.T) {
	var s operator.Config
	assert.False(t, s.NeedGPU())

	s.Spec.Trainer.Resources.Limits.GPU = 1
	assert.True(t, s.NeedGPU())
}

func TestElastic(t *testing.T) {
	var s operator.Config
	assert.False(t, s.Elastic())

	s.Spec.Trainer.MinInstance = 1
	s.Spec.Trainer.MaxInstance = 2
	assert.True(t, s.Elastic())
}
