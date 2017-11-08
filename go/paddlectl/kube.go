package paddlectl

import (
	"fmt"

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

func createMaster(client *rest.RESTClient,
	resource, namespace, jobName string, job *paddlejob.TrainingJob) error {
	/*
		err = client.Post().
			Resource(Resource).
			Namespace(namespace).
			Body(example).
			Do().Into(&job)

		if err != nil {
			panic(err)
		}
		fmt.Printf("CREATED: %#v\n", result)

			err := tprclient.Get().
				Resource(resource).
				Namespace(namespace).
				Name(jobName).
				Do().Into(&example)

				if err != nil {
					if errors.IsNotFound(err) {
						// Create an instance of our TPR
						example := &Example{
							Metadata: metav1.ObjectMeta{
								Name: "example1",
							},
							Spec: ExampleSpec{
								Foo: "hello",
								Bar: true,
							},
						}

						var result Example
						err = tprclient.Post().
							Resource("examples").
							Namespace(api.NamespaceDefault).
							Body(example).
							Do().Into(&result)

						if err != nil {
							panic(err)
						}
						fmt.Printf("CREATED: %#v\n", result)
					} else {
						panic(err)
					}
				} else {
					fmt.Printf("GET: %#v\n", example)
				}
	*/
	return nil
}
