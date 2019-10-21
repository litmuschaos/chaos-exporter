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
