// Copyright (c) 2021 PaddlePaddle Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package constants

import (
	"knative.dev/serving/pkg/apis/autoscaling"
)

// PaddleService Key
const (
	PaddleService               = "paddleService"
	PaddleServiceDefaultPodName = "http1"
)

// PaddleService configuration name and namespce
const (
	PaddleServiceConfigName      = "paddleservice-config"
	PaddleServiceConfigNamespace = "paddleservice-system"
)

// PaddleService resource defaults
var (
	PaddleServiceDefaultCPU                               = "0.2"
	PaddleServiceDefaultMemory                            = "512Mi"
	PaddleServiceDefaultMinScale                          = 0 // 0 if scale-to-zero is desired
	PaddleServiceDefaultMaxScale                          = 0 // 0 means limitless
	PaddleServiceDefaultTimeout                     int64 = 300
	PaddleServiceDefaultScalingClass                      = autoscaling.KPA // kpa or hpa
	PaddleServiceDefaultScalingMetric                     = "concurrency"   // concurrency, rps or cpu (hpa required)
	PaddleServiceDefaultScalingTarget                     = 100
	PaddleServiceDefaultTargetUtilizationPercentage       = "70"
	PaddleServiceDefaultWindow                            = "60s"
	PaddleServiceDefaultPanicWindow                       = "10" // percentage of StableWindow
	PaddleServiceDefaultPanicThreshold                    = "200"
	PaddleServivceDefaultTrafficPercents                  = 50
)

var (
	ReadinessInitialDelaySeconds int32 = 60
	ReadinessFailureThreshold    int32 = 3
	ReadinessPeriodSeconds       int32 = 10
	ReadinessTimeoutSeconds      int32 = 180
	SuccessThreshold             int32 = 1
	LivenessInitialDelaySeconds  int32 = 60
	LivenessFailureThreshold     int32 = 3
	LivenessPeriodSeconds        int32 = 10
)

var (
	ServiceAnnotationsList = []string{
		autoscaling.MinScaleAnnotationKey,
		autoscaling.MaxScaleAnnotationKey,
		autoscaling.ClassAnnotationKey,
		autoscaling.MetricAnnotationKey,
		autoscaling.TargetAnnotationKey,
		autoscaling.TargetUtilizationPercentageKey,
		autoscaling.WindowAnnotationKey,
		autoscaling.PanicWindowPercentageAnnotationKey,
		autoscaling.PanicThresholdPercentageAnnotationKey,
		"kubectl.kubernetes.io/last-applied-configuration",
	}
)

func DefaultServiceName(name string) string {
	return name + "-default"
}

func CanaryServiceName(name string) string {
	return name + "-canary"
}
