// Copyright 2021 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"log"
	"os"

	pdv1 "github.com/paddleflow/paddle-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	c      client.Client
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(pdv1.AddToScheme(scheme))
}

func main() {
	c, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Println("Create client failed")
		os.Exit(1)
	}

	name := "paddlejob-test"
	ns := "paddle-system"

	pdj := &pdv1.PaddleJob{}
	pdj.Namespace = ns
	pdj.Name = name
	pdj.Spec.PS.Template.Spec.Containers = []corev1.Container{}
	pdj.Spec.Worker.Template.Spec.Containers = []corev1.Container{}

	err = c.Create(context.Background(), pdj)
	if err != nil {
		log.Println("Create pdj failed", err)
		os.Exit(1)
	} else {
		log.Println("Create pdj successed")
	}

	pdj2 := pdv1.PaddleJob{}
	err = c.Get(context.Background(), client.ObjectKey{Namespace: ns, Name: name}, &pdj2)
	if err != nil {
		log.Println("Get pdj failed", err)
		os.Exit(1)
	} else {
		log.Println("Get pdj successed", pdj2.Name)
	}

	pdjl := &pdv1.PaddleJobList{}
	err = c.List(context.Background(), pdjl, client.InNamespace(ns))
	if err != nil {
		log.Println("List pdj failed", err)
		os.Exit(1)
	} else {
		log.Println("List pdj successed")
	}

	err = c.Delete(context.Background(), pdj)
	if err != nil {
		log.Println("Delete pdj failed", err)
		os.Exit(1)
	} else {
		log.Println("Delete  pdj successed")
	}

}
