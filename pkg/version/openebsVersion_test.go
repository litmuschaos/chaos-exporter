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
