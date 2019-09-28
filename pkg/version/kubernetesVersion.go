package version

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
)

// GetKubernetesVersion function gets kubernetes Version
func GetKubernetesVersion(clientSet *kubernetes.Clientset) (string, error) {
	// function to get Kubernetes Version
	version, err := clientSet.ServerVersion()
	if err != nil {
		fmt.Println("ClientSet is unable to communicate with the kubernetes cluster")
		return "N/A", err
	}
	return version.GitVersion, nil

}
