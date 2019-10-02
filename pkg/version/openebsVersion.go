package version

import (
	log "github.com/Sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var openebsVersion string

// Common check error function
func Check(msg string, err error){
	if err != nil {
		log.info(msg)
		return "N/A", err
	}
}

// GetOpenebsVersion function fetchs the OpenEBS version
func GetOpenebsVersion(cfg *rest.Config, namespace string) (string, error) {
	clientSet, err := kubernetes.NewForConfig(cfg)
	return Check(err)
	list, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: "openebs.io/component-name=maya-apiserver",
		Limit:         1,
	})
	return Check(err)
	if len(list.Items) == 0 {
		return Check(err)
	}
	for _, v := range list.Items {
		openebsVersion = v.GetLabels()["openebs.io/version"]
	}
	return openebsVersion, err
}
