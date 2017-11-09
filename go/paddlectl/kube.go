package paddlectl

import (
	"fmt"
	"os"
	"strings"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func createClient(kubeconfig string) (*rest.RESTClient, *kubernetes.Clientset, error) {
	config, err := buildConfig(kubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("init from config '%s' error: %v", kubeconfig, err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("create clientset from config '%s' error: %v", kubeconfig, err)
	}

	paddlejob.ConfigureClient(config)

	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, nil, fmt.Errorf("create rest client from config '%s' error: %v", kubeconfig, err)
	}

	return client, clientset, nil
}
func ensureNamespace(clientset *kubernetes.Clientset, namespace string) error {
	n := v1.Namespace{}
	n.SetName(namespace)

	if _, err := clientset.Namespaces().Get(namespace, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("get namespace '%s' error:%v", namespace, err)
	}

	return nil
}

func ensureTPR(clientset *kubernetes.Clientset, resource, namespace, apiversion string) {
	tpr, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Get(resource, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			tpr := &v1beta1.ThirdPartyResource{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ThirdPartyResource",
					APIVersion: "extensions/v1beta1",
				},

				ObjectMeta: metav1.ObjectMeta{
					Name: resource,
				},

				Versions: []v1beta1.APIVersion{
					{Name: apiversion},
				},
				Description: "PaddlePaddle TrainingJob operator",
			}

			_, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Create(tpr)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	} else {
		fmt.Printf("SKIPPING: already exists %#v\n", tpr)
	}
}

func createTrainingJob(restClient *rest.RESTClient, namespace string, job *paddlejob.TrainingJob) error {
	var result paddlejob.TrainingJob
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

func nameEscape(name string) string {
	name = strings.Replace(name, "@", "-", -1)
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	return name
}
