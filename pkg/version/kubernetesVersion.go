package version

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

const (
	k8sVersionNotFound = "N/A"
)

// GetKubernetesVersion function gets kubernetes Version
func GetKubernetesVersion(clientSet kubernetes.Interface) (string, error) {
	// function to get Kubernetes Version
	version, err := clientSet.Discovery().ServerVersion()
	if err != nil {
		fmt.Println("ClientSet is unable to communicate with the kubernetes cluster")
		return k8sVersionNotFound, err
	}
	return version.GitVersion, nil

}
