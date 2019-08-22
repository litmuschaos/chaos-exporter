package util

import (
	"fmt"

	discovery "k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

func GetVersionInfo(cfg *rest.Config) (string, error) {
	clientSet, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		fmt.Println("Unable to create the required ClientSet")
		return "", err
	}
	version, err := clientSet.ServerVersion()
	if err != nil {
		fmt.Println("ClientSet is unable to communicate with the kubernetes cluster")
		return "", err
	}
	return version.GitVersion, nil

	/*fmt.Println("Server Major : ", version.Major)
	fmt.Println("Server Minor : ", version.Minor)
	fmt.Println("Server GitVersion : ", version.GitVersion)
	fmt.Println("Server GitCommit : ", version.GitCommit)
	fmt.Println("Server GitTreeState : ", version.GitTreeState)
	fmt.Println("Server BuildDate : ", version.BuildDate)
	fmt.Println("Server GoVersion : ", version.GoVersion)
	fmt.Println("Server Compiler : ", version.Compiler)
	fmt.Println("Server Platform : ", version.Platform)*/
}
