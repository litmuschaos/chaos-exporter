package version

import (
	"fmt"

	discovery "k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// GetKubernetesVersion function gets kubernetes Version
func GetKubernetesVersion(cfg *rest.Config) (string, error) {
	// function to get Kubernetes Version
	clientSet, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		fmt.Println("Unable to create the required ClientSet")
		return "N/A", err
	}
	version, err := clientSet.ServerVersion()
	if err != nil {
		fmt.Println("ClientSet is unable to communicate with the kubernetes cluster")
		return "N/A", err
	}
	return version.GitVersion, nil

}
