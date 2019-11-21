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
	"k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"testing"

	"k8s.io/client-go/kubernetes/fake"
)

const (
	nsName             = "my-test-ns"
	podName            = "my-test-pod"
	openebsTestVersion = "my-test-version"
)

func getFakeClientSetWithNamespaceAndPod() kubernetes.Interface {
	cs := fake.NewSimpleClientset()
	_,_ = cs.CoreV1().Namespaces().Create(
		&v1.Namespace{
			ObjectMeta: v12.ObjectMeta{
				Name: nsName,
			},
		})
	_, _ = cs.CoreV1().Pods(nsName).Create(
		&v1.Pod{
			ObjectMeta: v12.ObjectMeta{
				Name: podName,
				Labels: map[string]string{
					openebsVersionLabelKey: openebsTestVersion,
					openebsMayaLabelKey:    openebsMayaLabelValue,
				},
			},
		})
	return cs
}

func getFakeClientSetWithoutNamespaceAndPod() kubernetes.Interface {
	cs := fake.NewSimpleClientset()
	return cs
}

func TestGetOpenebsVersion(t *testing.T) {
	tests := map[string]struct {
		cs      kubernetes.Interface
		ns      string
		version string
		isErr   bool
	}{
		"Test Positive-1": {
			cs:      getFakeClientSetWithNamespaceAndPod(),
			ns:      nsName,
			version: openebsTestVersion,
			isErr:   false,
		},
		"Test Negative-1": {
			cs:      getFakeClientSetWithoutNamespaceAndPod(),
			ns:      nsName,
			version: openebsTestVersion,
			isErr:   true,
		},
	}

	for name, mock := range tests {
		name, mock := name, mock
		t.Run(name, func(t *testing.T) {
			version, err := GetOpenebsVersion(mock.cs, mock.ns)
			if mock.isErr && err == nil {
				t.Fatalf("Test %q failed: expected error not to be nil", name)
			}
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}
			if version != mock.version {
				t.Fatalf("Test %q failed: expected version %q but got %q", name, mock.version, version)
			}
		})
	}
}
