package version

import (
	log "github.com/Sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var openebsVersion = "N/A"

func Check(err error, msg string) {
	if err != nil {
		log.Info(msg)
	}
}

func GetClientSet(cfg *rest.Config) (*kubernetes.Clientset, error) {
	clientSet, err := kubernetes.NewForConfig(cfg)
	Check(err, "Unable to create the required ClientSet")
	return clientSet, err
}

func ObtainList(clientSet *kubernetes.Clientset, namespace string) (*v1.PodList, error) {
	list, err := clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{
		LabelSelector: "openebs.io/component-name=maya-apiserver",
		Limit:         1,
	})
	Check(err, "Unable to find openebs / maya api-server")
	return list, err
}

func CheckIfEmptyList(list *v1.PodList) bool {
	if len(list.Items) == 0 {
		log.Info("No resources with labels 'openebs.io/component-name=maya-apiserver' found")
		return true
	}
	return false
}

// GetOpenebsVersion function fetchs the OpenEBS version
func GetOpenebsVersion(cfg *rest.Config, namespace string) (string, error) {
	if clientSet, err := GetClientSet(cfg); err != nil {
		return openebsVersion, err
	}
	if list, err := ObtainList(clientSet, namespace); err != nil {
		return openebsVersion, err
	}
	if ok := CheckIfEmptyList(list); ok {
		return openebsVersion, err
	}
	for _, v := range list.Items {
		openebsVersion = v.GetLabels()["openebs.io/version"]
	}
	return openebsVersion, err
}
