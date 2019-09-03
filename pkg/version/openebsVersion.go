package version

import (
	log "github.com/Sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var openebsVersion string

// function to get the OpenEBS Version for metrics
func GetopenebsVersion(cfg *rest.Config) (string, error) {
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Info("Unable to create the required ClientSet")
		return "N/A", err
	}
	list, err := clientSet.CoreV1().Pods("openebs").List(metav1.ListOptions{
		LabelSelector: "openebs.io/component-name=maya-apiserver",
		Limit:         1,
	})
	if err != nil {
		log.Info("Unable to fing openebs / maya api-server")
		return "N/A", err
	}
	if len(list.Items) == 0 {
		log.Info("No resources with labels 'openebs.io/component-name=maya-apiserver' found")
		return "N/A", err
	}
	for _, v := range list.Items {
		openebsVersion = v.GetLabels()["openebs.io/version"]
	}

	return openebsVersion, err

}
