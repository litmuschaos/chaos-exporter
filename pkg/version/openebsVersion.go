package version

import (
	log "github.com/Sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var openebsVersion string

// Common check error function
func Check(msg string, err error) {
	if err != nil {
		log.Info(msg)
		return "N/A", err
	}
}

// GetOpenebsVersion function fetchs the OpenEBS version
func GetOpenebsVersion(cfg *rest.Config, namespace string) (string, error) {
	clientSet, err := kubernetes.NewForConfig(cfg)
	openebsVersion, err := Check("Unable to create the required ClientSet", err)
	if err != nil {
		list, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{
			LabelSelector: "openebs.io/component-name=maya-apiserver",
			Limit:         1,
		})
		openebsVersion, err := Check("Unable to find openebs / maya api-server", err)
		if err != nil {
			if len(list.Items) == 0 {
				openebsVersion, err := Check("No resources with labels 'openebs.io/component-name=maya-apiserver' found", err)
			}
			for _, v := range list.Items {
				openebsVersion = v.GetLabels()["openebs.io/version"]
			}
		}
	}
	return openebsVersion, err
}
