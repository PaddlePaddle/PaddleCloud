package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseJob(t *testing.T) {
	testCase := []struct {
		path string
		job  job
	}{
		{
			"../case1-mnist-OFF-20-1-ON-400-round_0/mnist-case1-pass0.log",
			job{1, false, 20},
		},
		{
			"../case1-mnist-ON-20-1-ON-400-round_0/mnist-case1-pass0.log",
			job{1, true, 20},
		},
		{
			"../case2-mnist-OFF-6-1-ON-400-round_1/mnist-case2.log",
			job{2, false, 6},
		},
		{
			"../case2-mnist-ON-6-1-ON-400-round_1/mnist-case2.log",
			job{2, true, 6},
		},
	}

	for _, c := range testCase {
		assert.Equal(t, c.job, parseJob(c.path))
	}
}

func TestParseJobs(t *testing.T) {
	paths := []string{"../case1-mnist-OFF-20-1-ON-400-round_0/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_1/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_2/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_3/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_4/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_5/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_6/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_7/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_8/mnist-case1-pass0.log", "../case1-mnist-OFF-20-1-ON-400-round_9/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_0/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_1/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_2/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_3/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_4/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_5/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_6/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_7/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_8/mnist-case1-pass0.log", "../case1-mnist-ON-20-1-ON-400-round_9/mnist-case1-pass0.log", "../case2-mnist-OFF-6-1-ON-400-round_0/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_1/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_2/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_3/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_4/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_5/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_6/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_7/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_8/mnist-case2.log", "../case2-mnist-OFF-6-1-ON-400-round_9/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_0/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_1/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_2/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_4/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_5/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_6/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_7/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_8/mnist-case2.log", "../case2-mnist-ON-6-1-ON-400-round_9/mnist-case2.log"}

	// we are good if nothing panics
	for _, p := range paths {
		parseJob(p)
	}
}

func TestParseJobCase(t *testing.T) {
	c := parseJobCase("./test_data/case1.txt")
	assert.Equal(t, 3, len(c))
	tc := make([]int, 20)
	for i := range tc {
		tc[i] = 5 + i
	}
	cs := make([]float64, 20)
	for i := range cs {
		cs[i] = float64(i + 1)
	}
	assert.Equal(t, row{
		5, 10.88, 33, 19,
		1, 2, 3, 4,
		tc,
		cs,
	}, c[1])
}
