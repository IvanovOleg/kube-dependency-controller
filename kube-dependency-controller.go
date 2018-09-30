/*

Kubernetes service endpoint discovery

*/

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/kubernetes/pkg/apis/core"
)

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

var err error

func main() {
	var config *rest.Config
	//namespaceName := os.Getenv("ENDPOINT_NAMESPACE_NAME")
	kubernetesServiceHost := os.Getenv("KUBERNETES_SERVICE_HOST")
	kubernetesServicePort := os.Getenv("KUBERNETES_SERVICE_PORT")
	kubeconfigPath := parseConfig()
	//deployments := os.Getenv("DEPLOYMENTS")

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

	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}

	if "aaa" in list {
		
	}

	deployment, err := deploymentsClient.Get("aaa", metav1.GetOptions{})
	fmt.Printf(strconv.Itoa(int(deployment.Status.ReadyReplicas)))
}
