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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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
	dependenciesArray := strings.Split(strings.Replace(dependencies, " ", "", -1), ",")
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

	for _, dependency := range dependencies {

		switch dependency.dependencyType {
		case "deployment":
			deploymentsClient := clientset.AppsV1().Deployments(dependency.dependencyNamespace)
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
			fmt.Print("Deployment is ready\n")

		case "daemonset":
			daemonsetClient := clientset.AppsV1().DaemonSets(dependency.dependencyNamespace)
			for exists := false; exists; exists = true {
				list, err := daemonsetClient.List(metav1.ListOptions{})

				if err != nil {
					panic(err)
				}

				if inArray(dependency.dependencyName, list) {
					for ready := false; ready; ready = true {
						daemonset, err := daemonsetClient.Get(dependency.dependencyName, metav1.GetOptions{})

						if err != nil {
							panic(err)
						}

						if daemonset.Status.NumberAvailable == daemonset.Status.NumberReady {
							ready = true
						} else {
							time.Sleep(5 * time.Second)
						}
					}
				}

				time.Sleep(5 * time.Second)
			}
			fmt.Print("Daemonset is ready\n")

		case "statefulset":
			statefulsetClient := clientset.AppsV1().StatefulSets(dependency.dependencyNamespace)
			for exists := false; exists; exists = true {
				list, err := statefulsetClient.List(metav1.ListOptions{})

				if err != nil {
					panic(err)
				}

				if inArray(dependency.dependencyName, list) {
					for ready := false; ready; ready = true {
						statefulset, err := statefulsetClient.Get(dependency.dependencyName, metav1.GetOptions{})

						if err != nil {
							panic(err)
						}

						if statefulset.Status.Replicas == statefulset.Status.ReadyReplicas {
							ready = true
						} else {
							time.Sleep(5 * time.Second)
						}
					}
				}

				time.Sleep(5 * time.Second)
			}
			fmt.Print("Statefulset is ready\n")

		case "service":
			serviceClient := clientset.Core().Services(dependency.dependencyNamespace)
			for exists := false; exists; exists = true {
				list, err := serviceClient.List(metav1.ListOptions{})

				if err != nil {
					panic(err)
				}

				if inArray(dependency.dependencyName, list) {
					exists = true
				}

				time.Sleep(5 * time.Second)
			}
			fmt.Print("Service exists\n")

		case "configmap":
			configmapClient := clientset.Core().ConfigMaps(dependency.dependencyNamespace)
			for exists := false; exists; exists = true {
				list, err := configmapClient.List(metav1.ListOptions{})

				if err != nil {
					panic(err)
				}

				if inArray(dependency.dependencyName, list) {
					exists = true
				}

				time.Sleep(5 * time.Second)
			}
			fmt.Print("Configmap exists\n")

		case "secret":
			secretClient := clientset.Core().Secrets(dependency.dependencyNamespace)
			for exists := false; exists; exists = true {
				list, err := secretClient.List(metav1.ListOptions{})

				if err != nil {
					panic(err)
				}

				if inArray(dependency.dependencyName, list) {
					exists = true
				}

				time.Sleep(5 * time.Second)
			}
			fmt.Print("Secret exists\n")
		case "job":
			jobClient := clientset.BatchV1().Jobs(dependency.dependencyNamespace)
			for exists := false; exists; exists = true {
				list, err := jobClient.List(metav1.ListOptions{})

				if err != nil {
					panic(err)
				}

				if inArray(dependency.dependencyName, list) {
					exists = true
				}

				time.Sleep(5 * time.Second)
			}
			fmt.Print("Job exists\n")
		case "serviceaccount":
			serviceAccountClient := clientset.Core().ServiceAccounts(dependency.dependencyNamespace)
			for exists := false; exists; exists = true {
				list, err := serviceAccountClient.List(metav1.ListOptions{})

				if err != nil {
					panic(err)
				}

				if inArray(dependency.dependencyName, list) {
					exists = true
				}

				time.Sleep(5 * time.Second)
			}
			fmt.Print("Service account exists\n")
		}
	}
}
