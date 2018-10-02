/*

Kubernetes service endpoint discovery

*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/kubernetes/pkg/apis/core"
)

type dependency struct {
	dependencyNamespace string
	dependencyType      string
	dependencyName      string
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func parseConfig() *string {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	return kubeconfig
}

func buildExternalConfig(kubeconfig *string) *rest.Config {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return config
}

func parseDependenciesString(dependencies string) []dependency {
	var dependenciesList []dependency
	dependenciesArray := strings.Split(dependencies, ",")
	for _, element := range dependenciesArray {
		dependencyElements := strings.Split(element, "/")
		dep := dependency{
			dependencyNamespace: dependencyElements[0],
			dependencyType:      dependencyElements[1],
			dependencyName:      dependencyElements[2],
		}
		dependenciesList = append(dependenciesList, dep)
	}
	return dependenciesList
}

func inArray(val interface{}, array interface{}) (exists bool) {
	exists = false

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				exists = true
				return
			}
		}
	}

	return
}

var err error

func main() {
	var config *rest.Config
	kubernetesServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	kubernetesServicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	kubeconfigPath := parseConfig()
	dependencies := parseDependenciesString(os.Getenv("DEPENDENCIES"))

	//check if the app is running inside the kubernetes cluster
	if (kubernetesServiceHost != "") && (kubernetesServicePort != "") {
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	} else {
		if _, err := os.Stat(*kubeconfigPath); err == nil {
			config = buildExternalConfig(kubeconfigPath)
		}
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	deploymentsClient := clientset.AppsV1().Deployments(core.NamespaceDefault)

	for _, dependency := range dependencies {

		switch dependency.dependencyType {
		case "deployment":
			for exists := false; exists; exists = true {
				list, err := deploymentsClient.List(metav1.ListOptions{})

				if err != nil {
					panic(err)
				}

				if inArray(dependency.dependencyName, list) {
					for ready := false; ready; ready = true {
						deployment, err := deploymentsClient.Get(dependency.dependencyName, metav1.GetOptions{})

						if err != nil {
							panic(err)
						}

						if deployment.Status.Replicas == deployment.Status.ReadyReplicas {
							ready = true
						} else {
							time.Sleep(5 * time.Second)
						}
					}
				}

				time.Sleep(5 * time.Second)
			}
			fmt.Print("ready")
		}
	}
}
