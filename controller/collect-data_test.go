package controller

import (
	"context"
	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	v1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	litmusFakeClientSet "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/fake"
	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"testing"
)

func TestGetResultList(t *testing.T) {
	chaosNamespace := "Fake Namespace"
	chaosEngine := &v1alpha1.ChaosEngine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Fake Engine",
			Namespace: chaosNamespace,
		},
		Spec: v1alpha1.ChaosEngineSpec{
			Appinfo: v1alpha1.ApplicationParams{
				Appns:    "litmus",
				Applabel: "app=nginx",
				AppKind:  "deployment",
			},
			ChaosServiceAccount: "pod-delete-sa",
			Components: v1alpha1.ComponentParams{
				Runner: v1alpha1.RunnerInfo{
					Image: "litmuschaos/chaos-runner:ci",
					Type:  "go",
				},
			},
			JobCleanUpPolicy: "retain",
			EngineState:      "active",
			Experiments: []v1alpha1.ExperimentList{
				{
					Name: "pod-delete",
					Spec: v1alpha1.ExperimentAttributes{
						Components: v1alpha1.ExperimentComponents{
							ExperimentImage: "litmuschaos/go-runner:ci",
						},
					},
				},
			},
		},
	}
	client := CreateFakeClient(t)
	_, err := client.LitmusClient.ChaosEngines(chaosNamespace).Create(context.Background(), chaosEngine, metav1.CreateOptions{})
	assert.NoError(t, err, "Failed to create ChaosEngine")

	// Create the monitoringEnabled object
	monitoringEnabled := &MonitoringEnabled{
		IsChaosResultsAvailable: true,
	}
	t.Run("Success", func(t *testing.T) {
		resultList, err := GetResultList(client, chaosNamespace, monitoringEnabled)

		// Assertions
		assert.NoError(t, err, "Failed to get ChaosResultList")
		assert.NotNil(t, resultList, "ChaosResultList is nil")
	})

}

func CreateFakeClient(t *testing.T) clients.ClientSets {
	cs := clients.ClientSets{}
	cs.KubeClient = fake.NewSimpleClientset([]runtime.Object{}...)

	clientSet := litmusFakeClientSet.NewSimpleClientset()
	LitmusRestClient := clientSet.LitmuschaosV1alpha1().RESTClient()
	cs.LitmusClient = chaosClient.New(LitmusRestClient)
	return cs
}
