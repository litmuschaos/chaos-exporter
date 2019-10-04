package version

import (
	log "github.com/Sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var openebsVersion = "N/A"

func Check(err error, msg string) {
	if err != nil {
		log.info(msg)
	}
}

func GetClientSet(cfg *rest.Config) {
  clientSet, err := kubernetes.NewForConfig(cfg)
	Check(err, "Unable to create the required ClientSet")
  return clientSet, err
}

func ObtainList(clientSet Clientset) {
	list, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: "openebs.io/component-name=maya-apiserver",
		Limit:         1,
	})
	Check(err, "Unable to find openebs / maya api-server")
	return list, err
}
// GetOpenebsVersion function fetchs the OpenEBS version
func GetOpenebsVersion(cfg *rest.Config, namespace string) (string, error) {
	clientSet, err := GetClientSet(cfg)
	if err != nil {
		return openebsVersion, err
	}
	list, err := ObtainList(clientSet)
	if err != nil {
		return  openebsVersion, err
	}
	if len(list.Items) == 0 {
		log.Info("No resources with labels 'openebs.io/component-name=maya-apiserver' found")
		return openebsVersion, err
	}
	for _, v := range list.Items {
		openebsVersion = v.GetLabels()["openebs.io/version"]
	}

	return openebsVersion, err

}
