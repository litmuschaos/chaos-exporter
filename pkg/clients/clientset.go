package clients

import (
	"flag"
	"fmt"
	"os"
	"time"

	clientv1alpha1 "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned"
	litmusInformer "github.com/litmuschaos/chaos-operator/pkg/client/informers/externalversions"
	"github.com/litmuschaos/chaos-operator/pkg/client/listers/litmuschaos/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
)

// ClientSets is a collection of clientSets and kubeConfig needed
type ClientSets struct {
	KubeClient     kubernetes.Interface
	EventsInformer v1.EventLister
	EngineInformer v1alpha1.ChaosEngineLister
	ResultInformer v1alpha1.ChaosResultLister
	LitmusClient   clientv1alpha1.Interface
	KubeConfig     *rest.Config
}

const (
	ProcessKey = "process"
)

// NewClientSet will generation both ClientSets (k8s, and Litmus) as well as the KubeConfig
func NewClientSet(stopCh <-chan struct{}, resyncDuration time.Duration, wq workqueue.RateLimitingInterface) (ClientSets, error) {

	config, err := getKubeConfig()
	if err != nil {
		return ClientSets{}, err
	}

	k8sClientSet, err := GenerateK8sClientSet(config)
	if err != nil {
		return ClientSets{}, err
	}

	litmusClientSet, err := GenerateLitmusClientSet(config)
	if err != nil {
		return ClientSets{}, err
	}

	clientSets := ClientSets{}
	clientSets.KubeClient = k8sClientSet
	clientSets.LitmusClient = litmusClientSet
	clientSets.KubeConfig = config

	if err := clientSets.SetupInformers(stopCh, k8sClientSet, litmusClientSet, resyncDuration, wq); err != nil {
		return ClientSets{}, err
	}
	return clientSets, nil
}

func (clientSets *ClientSets) SetupInformers(stopCh <-chan struct{}, k8sClientSet kubernetes.Interface, litmusClientSet clientv1alpha1.Interface, resyncDuration time.Duration, wq workqueue.RateLimitingInterface) error {
	watchNamespace := os.Getenv("WATCH_NAMESPACE")
	var (
		factory       informers.SharedInformerFactory
		litmusFactory litmusInformer.SharedInformerFactory
	)
	if watchNamespace == "" {
		factory = informers.NewSharedInformerFactory(k8sClientSet, resyncDuration)
		litmusFactory = litmusInformer.NewSharedInformerFactory(litmusClientSet, resyncDuration)
	} else {
		factory = informers.NewSharedInformerFactoryWithOptions(k8sClientSet, resyncDuration, informers.WithNamespace(watchNamespace))
		litmusFactory = litmusInformer.NewSharedInformerFactoryWithOptions(litmusClientSet, resyncDuration, litmusInformer.WithNamespace(watchNamespace))
	}

	eventsInformer := factory.Core().V1().Events().Informer()
	clientSets.EventsInformer = factory.Core().V1().Events().Lister()

	chaosEngineInformer := litmusFactory.Litmuschaos().V1alpha1().ChaosEngines().Informer()
	chaosResultInformer := litmusFactory.Litmuschaos().V1alpha1().ChaosResults().Informer()

	// queue up for processing if there is any change in the resources
	chaosEngineInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			wq.Add(ProcessKey)
		},
		UpdateFunc: func(old, new interface{}) {
			wq.Add(ProcessKey)
		},
		DeleteFunc: func(obj interface{}) {
			wq.Add(ProcessKey)
		},
	})
	chaosResultInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			wq.Add(ProcessKey)
		},
		UpdateFunc: func(old, new interface{}) {
			wq.Add(ProcessKey)
		},
		DeleteFunc: func(obj interface{}) {
			wq.Add(ProcessKey)
		},
	})

	clientSets.EngineInformer = litmusFactory.Litmuschaos().V1alpha1().ChaosEngines().Lister()
	clientSets.ResultInformer = litmusFactory.Litmuschaos().V1alpha1().ChaosResults().Lister()

	go eventsInformer.Run(stopCh)
	go chaosEngineInformer.Run(stopCh)
	go chaosResultInformer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, eventsInformer.HasSynced, chaosEngineInformer.HasSynced, chaosResultInformer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}
	return nil
}

// getKubeConfig setup the config for access cluster resource
func getKubeConfig() (*rest.Config, error) {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()
	// It uses in-cluster config, if kubeconfig path is not specified
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	return config, err
}

// GenerateK8sClientSet will generation k8s client
func GenerateK8sClientSet(config *rest.Config) (*kubernetes.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to generate kubernetes clientSet, err: %v: ", err)
	}
	return k8sClientSet, nil
}

// GenerateLitmusClientSet will generate a LitmusClient
func GenerateLitmusClientSet(config *rest.Config) (*clientv1alpha1.Clientset, error) {
	litmusClientSet, err := clientv1alpha1.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create LitmusClientSet, err: %v", err)
	}
	return litmusClientSet, nil
}
