package version

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var openebsVersion string
var openebsLabel = "openebs.io/component-name=maya-apiserver"

// GetOpenebsVersion function fetchs the OpenEBS version
func GetOpenebsVersion(clientSet *kubernetes.Clientset, namespace string) (string, error) {
	podList, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{LabelSelector: openebsLabel, Limit: 1})
	if err != nil {
		return "", fmt.Errorf("unable to find openebs/maya api-server %s", err)
	}
	if len(podList.Items) == 0 {
		return "", fmt.Errorf("no resources with labels 'openebs.io/component-name=maya-apiserver' found")
	}
	for _, v := range podList.Items {
		openebsVersion = v.GetLabels()["openebs.io/version"]
	}
	return openebsVersion, err
}
