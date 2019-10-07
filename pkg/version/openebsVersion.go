package version

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"fmt"
)

var openebsVersion = ""
var openebsLabel = "openebs.io/component-name=maya-apiserver"

// ObtainList fetches the list of pods
func ObtainList(clientSet *kubernetes.Clientset, namespace string) (*v1.PodList, error) {
	list, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: openebsLabel,
		Limit:         1,
	})
	return list, err
}

// CheckIfEmptyList checks if the obtained pod list is empty
func CheckIfEmptyList(list *v1.PodList) bool {
	if len(list.Items) == 0 {
		return true
	}
	return false
}

// GetOpenebsVersion function fetchs the OpenEBS version
func GetOpenebsVersion(clientSet *kubernetes.Clientset, namespace string) (string, error) {
	list, err := ObtainList(clientSet, namespace)
	if err != nil {
		return openebsVersion, fmt.Errorf("unable to find openebs/maya api-server %s", err)
	}
	if CheckIfEmptyList(list) {
		return openebsVersion, fmt.Errorf("no resources with labels 'openebs.io/component-name=maya-apiserver' found")
	}
	for _, v := range list.Items {
		openebsVersion = v.GetLabels()["openebs.io/version"]
	}
	return openebsVersion, err
}
