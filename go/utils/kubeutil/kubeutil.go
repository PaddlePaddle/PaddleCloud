package kubeutil

import (
	"fmt"
	"os"
	"strings"

	edlresource "github.com/PaddlePaddle/cloud/go/edl/resource"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

// CreateClient creates ClientSet and rest.RESTClient used by client.
func CreateClient(kubeconfig string) (*rest.RESTClient, *kubernetes.Clientset, error) {
	config, err := buildConfig(kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("init from config '%s' error: %v", kubeconfig, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("create clientset from config '%s' error: %v", kubeconfig, err)
	}

	edlresource.RegisterTrainingJob(config)

	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, nil, fmt.Errorf("create rest client from config '%s' error: %v", kubeconfig, err)
	}

	return client, clientset, nil
}

// FindNamespace finds whether a namespace exists.
func FindNamespace(clientset *kubernetes.Clientset, namespace string) error {
	n := v1.Namespace{}
	n.SetName(namespace)

	if _, err := clientset.CoreV1().Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("get namespace '%s' error:%v", namespace, err)
	}

	return nil
}

// EnsureTPR ensure a TPR should exists and create it if not.
func EnsureTPR(clientset *kubernetes.Clientset, resourceName, apiversion string) error {
	tpr, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Get(resourceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			tpr := &v1beta1.ThirdPartyResource{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ThirdPartyResource",
					APIVersion: "extensions/v1beta1",
				},

				ObjectMeta: metav1.ObjectMeta{
					Name: resourceName,
				},

				Versions: []v1beta1.APIVersion{
					{Name: apiversion},
				},
				Description: "PaddlePaddle TrainingJob operator",
			}

			_, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Create(tpr)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		fmt.Printf("SKIPPING: already exists %#v\n", tpr)
	}

	return nil
}

// CreateTrainingJob try to create a training-job under namespace.
func CreateTrainingJob(restClient *rest.RESTClient, namespace string, job *edlresource.TrainingJob) error {
	var result edlresource.TrainingJob
	err := restClient.Post().
		Resource("trainingjobs").
		Namespace(namespace).
		Body(job).
		Do().Into(&result)

	if err != nil {
		fmt.Fprintf(os.Stderr, "can't create TPR extenion TrainningJob: %s\n", job)
		return err
	}

	return nil
}

// NameEscape replace characters not supported by Kubernetes.
func NameEscape(name string) string {
	name = strings.Replace(name, "@", "-", -1)
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	return name
}
