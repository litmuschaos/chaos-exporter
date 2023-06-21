package controller

import (
	"context"
	"fmt"
	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	litmusFakeClientSet "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"testing"
)

func TestGetResultList(t *testing.T) {
	FakeChaosNameSpace := "Fake Namespace"
	FakeEngineName := "Fake Engine"

	tests := map[string]struct {
		instance        *v1alpha1.ChaosEngine
		isErr           bool
		chaosengine     *v1alpha1.ChaosEngine
		chaosresultlist *v1alpha1.ChaosResultList
		monitoring      *MonitoringEnabled
	}{
		"Test Positive-1": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FakeEngineName,
					Namespace: FakeChaosNameSpace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					Appinfo: v1alpha1.ApplicationParams{
						Applabel: "app=nginx",
						AppKind:  "deployment",
					},
					EngineState: v1alpha1.EngineStateActive,
					Components: v1alpha1.ComponentParams{
						Runner: v1alpha1.RunnerInfo{
							Image: "fake-runner-image",
						},
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "exp-1",
						},
					},
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
				},
			},
			isErr: false,
			monitoring: &MonitoringEnabled{
				IsChaosResultsAvailable: true,
				IsChaosEnginesAvailable: true,
			},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {

			client := CreateFakeClient(t)
			_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(FakeChaosNameSpace).Create(context.Background(), mock.chaosengine, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("engine not created for %v test, err: %v", name, err)
			}

			resultList, err := GetResultList(client, FakeChaosNameSpace, mock.monitoring)
			//if !mock.isErr && err != nil && mock.chaosresultlist != resultList {
			//	t.Fatalf("test Failed as not able to get the Chaos result list")
			//}
			fmt.Print(resultList)

		})
	}
}

func CreateFakeClient(t *testing.T) clients.ClientSets {
	cs := clients.ClientSets{}
	cs.KubeClient = fake.NewSimpleClientset([]runtime.Object{}...)
	cs.LitmusClient = litmusFakeClientSet.NewSimpleClientset([]runtime.Object{}...)
	return cs
}
