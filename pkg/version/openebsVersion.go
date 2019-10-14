package version

import (
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var openebsVersion string

const (
	openebsMayaLabelKey    = "openebs.io/component-name"
	openebsMayaLabelValue  = "maya-apiserver"
	openebsVersionLabelKey = "openebs.io/version"
)

// getPodList fetches the list of pods
func getPodList(clientSet kubernetes.Interface, namespace string) (*v1.PodList, error) {
	list, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: openebsMayaLabelKey + "=" + openebsMayaLabelValue,
		Limit:         1,
	})
	return list, err
}

// GetOpenebsVersion function fetches the OpenEBS version
func GetOpenebsVersion(clientSet kubernetes.Interface, namespace string) (string, error) {
	podList, err := getPodList(clientSet, namespace)
	if err != nil {
		return openebsVersion, fmt.Errorf("unable to find openebs/maya api-server %s", err)
	}
	if len(podList.Items) == 0 {
		return openebsVersion, fmt.Errorf("no resources with labels 'openebs.io/component-name=maya-apiserver' found")
	}
	for _, v := range podList.Items {
		openebsVersion = v.GetLabels()[openebsVersionLabelKey]
	}
	return openebsVersion, err
}
