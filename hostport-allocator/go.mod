module hostport-manager

go 1.13

require (
	github.com/PaddlePaddle/cloud v0.1.1-beta.3.0.20180109073715-c66c772e8bab
	github.com/emicklei/go-restful-swagger12 v0.0.0-20201014110547-68ccff494617 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/howeyc/gopass v0.0.0-20190910152052-7cb4b85ec19c // indirect
	github.com/juju/ratelimit v0.0.0-20151125201925-77ed1c8a0121 // indirect
	github.com/paddleflow/paddle-operator v0.1.0 // indirect
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v4.0.0+incompatible
	k8s.io/kubernetes v1.20.4
)

replace (
	k8s.io/api v0.0.0 => k8s.io/api v0.20.4
	k8s.io/apiextensions-apiserver v0.0.0 => k8s.io/apiextensions-apiserver v0.20.4
	k8s.io/apimachinery v0.0.0 => k8s.io/apimachinery v0.20.4
	k8s.io/apiserver v0.0.0 => k8s.io/apiserver v0.20.4
	k8s.io/cli-runtime v0.0.0 => k8s.io/cli-runtime v0.20.4
	k8s.io/client-go v0.0.0 => k8s.io/client-go v0.20.4
	k8s.io/cloud-provider v0.0.0 => k8s.io/cloud-provider v0.20.4
	k8s.io/cluster-bootstrap v0.0.0 => k8s.io/cluster-bootstrap v0.20.4
	k8s.io/code-generator v0.0.0 => k8s.io/code-generator v0.20.4
	k8s.io/component-base v0.0.0 => k8s.io/component-base v0.20.4
	k8s.io/component-helpers v0.0.0 => k8s.io/component-helpers v0.20.4
	k8s.io/controller-manager v0.0.0 => k8s.io/controller-manager v0.20.4
	k8s.io/cri-api v0.0.0 => k8s.io/cri-api v0.20.4
	k8s.io/csi-translation-lib v0.0.0 => k8s.io/csi-translation-lib v0.20.4
	k8s.io/kube-aggregator v0.0.0 => k8s.io/kube-aggregator v0.20.4
	k8s.io/kube-controller-manager v0.0.0 => k8s.io/kube-controller-manager v0.20.4
	k8s.io/kube-proxy v0.0.0 => k8s.io/kube-proxy v0.20.4
	k8s.io/kube-scheduler v0.0.0 => k8s.io/kube-scheduler v0.20.4
	k8s.io/kubectl v0.0.0 => k8s.io/kubectl v0.20.4
	k8s.io/kubelet v0.0.0 => k8s.io/kubelet v0.20.4
	k8s.io/legacy-cloud-providers v0.0.0 => k8s.io/legacy-cloud-providers v0.20.4
	k8s.io/metrics v0.0.0 => k8s.io/metrics v0.20.4
	k8s.io/mount-utils v0.0.0 => k8s.io/mount-utils v0.20.4
	k8s.io/sample-apiserver v0.0.0 => k8s.io/sample-apiserver v0.20.4
)
