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
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetKubernetesVersion(t *testing.T) {
	tests := map[string]struct {
		cs      kubernetes.Interface
		version string
		isErr   bool
	}{
		"pass mock clientset": {
			cs:      fake.NewSimpleClientset(),
			version: k8sVersionNotFound,
			isErr:   false,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			_, err := GetKubernetesVersion(mock.cs)
			if mock.isErr != (err != nil) {
				t.Fatalf("Should return error %t but got %v", mock.isErr, err)
			}
		})
	}
}
