package paddlectl

import (
	"fmt"
	"strings"

	paddlejob "github.com/PaddlePaddle/cloud/go/api"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func buildConfig(kubeconfig string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func createClient(kubeconfig string) (*rest.RESTClient, *kubernetes.Clientset) {
	config, err := buildConfig(kubeconfig)
	if err != nil {
		panic(err)
	}

	// setup some optional configuration
	paddlejob.ConfigureClient(config)

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	client, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}

	return client, clientset
}
func ensureNamespace(clientset *kubernetes.Clientset, namespace string) error {
	n := v1.Namespace{}
	n.SetName(namespace)

	if _, err := clientset.Namespaces().Create(&n); err != nil &&
		!errors.IsAlreadyExists(err) {
		fmt.Printf("create namespace %s error: %v\n", namespace, err)
		return err
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
			// TODO: need namespace?
			//tpr.SetNamespace(namespace)
			result, err := clientset.ExtensionsV1beta1().ThirdPartyResources().Create(tpr)
			if err != nil {
				panic(err)
			}
			fmt.Printf("CREATED: %#v\nFROM: %#v\n", result, tpr)
		} else {
			panic(err)
		}
	} else {
		fmt.Printf("SKIPPING: already exists %#v\n", tpr)
	}
}

func createTrainingJob(restClient *rest.RESTClient, job *paddlejob.TrainingJob) error {
	var result paddlejob.TrainingJob
	err := restClient.Post().
		Resource("TrainingJob").
		Namespace("gongweibao-baidu-com").
		Body(job).
		Do().Into(&result)

	/*
		err := restClient.
			Post().
			Body(job).
			Do().
			Into(&result)
	*/

	if err != nil {
		fmt.Printf("can't create TrainningJob: %s\n", job)
		fmt.Printf("restClient: %v\n", restClient)
		fmt.Printf("error: %v\n", err)
		return err
	}

	return nil
}

/*
def name_escape(name):
    """
        Escape name to a safe string of kubernetes namespace
    """
    safe_name = name.replace("@", "-")
    safe_name = safe_name.replace(".", "-")
    safe_name = safe_name.replace("_", "-")
    return safe_name
*/
func nameEscape(name string) string {
	name = strings.Replace(name, "@", "-", -1)
	name = strings.Replace(name, ".", "-", -1)
	name = strings.Replace(name, "_", "-", -1)
	fmt.Println(name)
	return name
}

func createDemo(restClient *rest.RESTClient) error {
	var example paddlejob.TrainingJob
	var result paddlejob.TrainingJob

	err := restClient.Post().
		Resource("TrainingJob").
		Namespace(api.NamespaceDefault).
		Body(example).
		Do().Into(&result)
	fmt.Printf("create 1\n")
	if err != nil {
		fmt.Printf("create demo error: %v\n", err)
		return err
	}

	return nil
}
