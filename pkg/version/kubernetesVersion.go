/*
Copyright 2019 LitmusChaos Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
