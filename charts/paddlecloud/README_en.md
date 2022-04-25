# PaddleCloud Helm Charts

PaddleCloud aims to provide a set of easy-to-use cloud components based on PaddlePaddle and related kits to meet customers' business cloud requirements. In order to get through the whole process from training to deployment, the model training component paddlejob, the model inference component serving, and the sample caching component sampleset for acceleration have been developed. The components provide users with almost zero-based experience tutorials and easy-to-use programming interfaces. 
For more details please refer to [PaddleCloud](https://github.com/PaddlePaddle/PaddleCloud).

### Prerequisites

* Kubernetes, 1.8 <= version <= 2.1
* kubectl
* helm

If you do not have a Kubernetes environment, you can refer to [microk8s official documentation](https://microk8s.io/docs/getting-started) for installation. If you use macOS system, or encounter installation problems, you can refer to the document [macOS install microk8s](./docs/macOS_install_microk8s.md).

### Installation

We assume that you have installed the kubernates cluster environment and you can access the cluster through command such as **helm** and **kubectl**. Otherwise, please refer to the more detailed [installation tutorial](./docs/tutorials/Installation_en.md) for help. If you deploy components in the production environment or have custom installation requirements, please also refer to [Installation Tutorial](./docs/tutorials/Installation_en.md).

Add and update helm's charts repositories,

```bash
$ helm repo add paddlecloud https://paddleflow-public.hkg.bcebos.com/charts
$ helm repo update
```

These commands deploy PaddleCloud on the Kubernetes cluster in the default configuration. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

Install all components and all dependencies using helm.(Optional)

```bash
$ helm install pdc paddlecloud/paddlecloud --set tags.all-dep=true --namespace paddlecloud --create-namespace
```

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `pdc` helm release:

```console
$ helm uninstall pdc
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

### Global parameters

| Name                      | Description                                     | Value |
| ------------------------- | ----------------------------------------------- | ----- |
| `global.imageRegistry`    | Global Docker image registry                    | `""`  |
| `global.imagePullSecrets` | Global Docker registry secret names as an array | `[]`  |
| `global.storageClass`     | Global StorageClass for Persistent Volume(s)    | `""`  |
| `global.redis.password`   | Global redis password, available when enabled redis (overrides `redis.auth.password`) | `""`  |


### Common parameters

| Name                     | Description                                                                             | Value           |
| ------------------------ | --------------------------------------------------------------------------------------- | --------------- |
| `kubeVersion`            | Override Kubernetes version                                                             | `""`            |
| `nameOverride`           | String to partially override common.names.fullname (will maintain the release name)     | `""`            |
| `fullnameOverride`       | String to fully override common.names.fullname                                          | `""`            |
| `clusterDomain`          | Kubernetes Cluster Domain                                                               | `cluster.local` |
| `commonLabels`           | Labels to add to all deployed objects                                                   | `{}`            |
| `commonAnnotations`      | Annotations to add to all deployed objects                                              | `{}`            |
| `diagnosticMode.enabled` | Enable diagnostic mode (all probes will be disabled and the command will be overridden) | `false`         |
| `diagnosticMode.command` | Command to override all containers in the the deployment(s)/daemonset(s)                | `["sleep"]`     |
| `diagnosticMode.args`    | Args to override all containers in the the deployment(s)/daemonset(s)                   | `["infinity"]`  |


### PaddleJob controller parameters

| Name                                        | Description                                                                                                              | Value                  |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ---------------------- |
| `paddlejob.enabled`                         | Enable PaddleJob controller when install                                                                                 | `true`                 |
| `paddlejob.image.registry`                  | PaddleJob controller image registry                                                                                      | `registry.baidubce.com`|
| `paddlejob.image.repository`                | PaddleJob controller repository                                                                     | `paddleflow-public/paddlecloud/paddlejob`   |
| `paddlejob.image.tag`                       | PaddleJob controller image tag (immutable tags are recommended)                                                          | `v0.4.0`               |
| `paddlejob.image.pullPolicy`                | PaddleJob controller image pull policy                                                                                   | `IfNotPresent`         |
| `paddlejob.image.pullSecrets`               | PaddleJob controller image pull secrets                                                                                  | `[]`                   |
| `paddlejob.command`                         | Override controller default command                                                                                      | `[]`                   |
| `paddlejob.args`                            | Override controller default args                                                                                         | `[]`                   |
| `paddlejob.resources.limits`                | The resources limits for the container containers                                                                        | `{}`                   |
| `paddlejob.resources.requests`              | The requested resources for the container containers                                                                     | `{}`                   |
| `paddlejob.podAffinityPreset`               | Pod affinity preset. Ignored if `paddlejob.affinity` is set. Allowed values: `soft` or `hard`                            | `""`                   |
| `paddlejob.podAntiAffinityPreset`           | Pod anti-affinity preset. Ignored if `paddlejob.affinity` is set. Allowed values: `soft` or `hard`                       | `soft`                 |
| `paddlejob.nodeAffinityPreset.type`         | Node affinity preset type. Ignored if `paddlejob.affinity` is set. Allowed values: `soft` or `hard`                      | `""`                   |
| `paddlejob.nodeAffinityPreset.key`          | Node label key to match. Ignored if `paddlejob.affinity` is set                                                          | `""`                   |
| `paddlejob.nodeAffinityPreset.values`       | Node label values to match. Ignored if `paddlejob.affinity` is set                                                       | `[]`                   |
| `paddlejob.affinity`                        | Affinity for pod assignment.                                                                                             | `{}`                   |
| `paddlejob.nodeSelector`                    | Node labels for pod assignment.                                                                                          | `{}`                   |
| `paddlejob.tolerations`                     | Tolerations for pod assignment.                                                                                          | `[]`                   |


### SampleSet controller parameters

| Name                                        | Description                                                                                                              | Value                  |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ---------------------- |
| `sampleset.enabled`                         | Enable SampleSet controller when install                                                                                 | `true`                 |
| `sampleset.image.registry`                  | SampleSet controller image registry                                                                                      | `registry.baidubce.com`|
| `sampleset.image.repository`                | SampleSet controller repository                                                                     | `paddleflow-public/paddlecloud/sampleset`   |
| `sampleset.image.tag`                       | SampleSet controller image tag (immutable tags are recommended)                                                          | `v0.4.0`               |
| `sampleset.image.pullPolicy`                | SampleSet controller image pull policy                                                                                   | `IfNotPresent`         |
| `sampleset.image.pullSecrets`               | SampleSet controller image pull secrets                                                                                  | `[]`                   |
| `sampleset.command`                         | Override controller default command                                                                                      | `[]`                   |
| `sampleset.args`                            | Override controller default args                                                                                         | `[]`                   |
| `sampleset.resources.limits`                | The resources limits for the container containers                                                                        | `{}`                   |
| `sampleset.resources.requests`              | The requested resources for the container containers                                                                     | `{}`                   |
| `sampleset.podAffinityPreset`               | Pod affinity preset. Ignored if `sampleset.affinity` is set. Allowed values: `soft` or `hard`                            | `""`                   |
| `sampleset.podAntiAffinityPreset`           | Pod anti-affinity preset. Ignored if `sampleset.affinity` is set. Allowed values: `soft` or `hard`                       | `soft`                 |
| `sampleset.nodeAffinityPreset.type`         | Node affinity preset type. Ignored if `sampleset.affinity` is set. Allowed values: `soft` or `hard`                      | `""`                   |
| `sampleset.nodeAffinityPreset.key`          | Node label key to match. Ignored if `sampleset.affinity` is set                                                          | `""`                   |
| `sampleset.nodeAffinityPreset.values`       | Node label values to match. Ignored if `sampleset.affinity` is set                                                       | `[]`                   |
| `sampleset.affinity`                        | Affinity for pod assignment.                                                                                             | `{}`                   |
| `sampleset.nodeSelector`                    | Node labels for pod assignment.                                                                                          | `{}`                   |
| `sampleset.tolerations`                     | Tolerations for pod assignment.                                                                                          | `[]`                   |


### Serving controller parameters

| Name                                        | Description                                                                                                              | Value                  |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ---------------------- |
| `serving.enabled`                           | Enable serving controller when install                                                                                   | `true`                 |
| `serving.image.registry`                    | Serving controller image registry                                                                                        | `registry.baidubce.com`|
| `serving.image.repository`                  | Serving controller repository                                                                         | `paddleflow-public/paddlecloud/serving`   |
| `serving.image.tag`                         | Serving controller image tag (immutable tags are recommended)                                                            | `v0.4.0`               |
| `serving.image.pullPolicy`                  | Serving controller image pull policy                                                                                     | `IfNotPresent`         |
| `serving.image.pullSecrets`                 | Serving controller image pull secrets                                                                                    | `[]`                   |
| `serving.command`                           | Override controller default command                                                                                      | `[]`                   |
| `serving.args`                              | Override controller default args                                                                                         | `[]`                   |
| `serving.resources.limits`                  | The resources limits for the container containers                                                                        | `{}`                   |
| `serving.resources.requests`                | The requested resources for the container containers                                                                     | `{}`                   |
| `serving.podAffinityPreset`                 | Pod affinity preset. Ignored if `serving.affinity` is set. Allowed values: `soft` or `hard`                              | `""`                   |
| `serving.podAntiAffinityPreset`             | Pod anti-affinity preset. Ignored if `serving.affinity` is set. Allowed values: `soft` or `hard`                         | `soft`                 |
| `serving.nodeAffinityPreset.type`           | Node affinity preset type. Ignored if `serving.affinity` is set. Allowed values: `soft` or `hard`                        | `""`                   |
| `serving.nodeAffinityPreset.key`            | Node label key to match. Ignored if `serving.affinity` is set                                                            | `""`                   |
| `serving.nodeAffinityPreset.values`         | Node label values to match. Ignored if `serving.affinity` is set                                                         | `[]`                   |
| `serving.affinity`                          | Affinity for pod assignment.                                                                                             | `{}`                   |
| `serving.nodeSelector`                      | Node labels for pod assignment.                                                                                          | `{}`                   |
| `serving.tolerations`                       | Tolerations for pod assignment.                                                                                          | `[]`                   |


### Tag for all dependencies used in PaddleCloud

| Name                                          | Description                                                                                                              | Value                             |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | --------------------------------- |
| `tags.all-dep`                                | Enable or disable all all dependencies used in PaddleCloud                                                               | `false`                           |


### HostPath provisioner parameters

for all parameters of HostPath provisioner, please refer to [hostpath-provisioner](https://artifacthub.io/packages/helm/rimusz/hostpath-provisioner)

| Name                                          | Description                                                                                                              | Value                             |
| --------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | --------------------------------- |
| `hostpath.nodeHostPath`                       | Set the local HostPath to be used on the node                                                                            | `/mnt/hostpath`                   |


### JuiceFS csi driver parameters

for all parameters of JuiceFS CSI Driver, please refer to [juicefs-csi-driver](https://github.com/juicedata/juicefs-csi-driver)

| Name                                             | Description                                                                                                                      | Value                         |
| ------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------- | ----------------------------- |
| `juicefs.kubeletDir`                             | kubelet working directory,can be set using `--root-dir` when starting kubelet。usually set as `/var/lib/kubelet`.  | `/var/snap/microk8s/common/var/lib/kubelet` |


### Redis chart parameters

for all parameters of Redis chart, please refer to [redis](https://artifacthub.io/packages/helm/bitnami/redis)

| Name                                             | Description                                              | Value        |
| ------------------------------------------------ | -------------------------------------------------------- | ------------ |
| `redis.architecture`                             | Allowed values: `standalone` or `replication`            | `standalone` |


### Knative serving chart parameters

for all parameters of knative-serving chart, please refer to [knative-serving]()

| Name                                                | Description                                                                                            | Value           |
| --------------------------------------------------- | ------------------------------------------------------------------------------------------------------ | --------------- |
| `knative.kourier.enabled`                           | Hub Dummy authenticator password, please change default value when install formal environment.         | `paddlepaddle`  |
| `knative.netIstio.enabled`                          | The password of postgresql which is used as storage in JupyterHub chart. please change default value.  | `paddlepaddle`  |


### Kubeflow Pipelines chart parameter

Kubeflow pipelines used mysql and minio to storage data, please refer to [kubeflow-pipelines](https://artifacthub.io/packages/helm/paddlecloud/knative-serving) for more details.
For all parameters of mysql chart, please refer to [mysql](https://artifacthub.io/packages/helm/bitnami/mysql).
for all parameters of minio chart, please refer to [minio](https://artifacthub.io/packages/helm/bitnami/minio).

| Name                                            | Description                     | Value                                |
| ----------------------------------------------- | ------------------------------- | ------------------------------------ |
| `pipelines.mysql.auth.rootPassword`             | Password for mysql              | `paddlepaddle`                          |
| `pipelines.minio.auth.rootUser`                 | Username for minio              | `minio`      |
| `pipelines.minio.auth.rootPassword`             | Password for minio              | `minio123`               |


## Configuration and installation details

### Set pod affinity

This chart allows you to set your custom affinity using the `paddlejob.affinity` and `sampleset.affinity` parameters. Refer to the [chart documentation on pod affinity](https://docs.bitnami.com/kubernetes/infrastructure/jupyterhub/configuration/configure-pod-affinity).

## License

Copyright (c) 2021 PaddlePaddle Authors. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
