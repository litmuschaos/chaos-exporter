package controller_test

import (
	"context"
	"github.com/litmuschaos/chaos-exporter/controller"
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
		chaosresult *v1alpha1.ChaosResult
		monitoring  *controller.MonitoringEnabled
		isErr       bool
	}{
		"Test Positive-1": {
			chaosresult: &v1alpha1.ChaosResult{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FakeEngineName,
					Namespace: FakeChaosNameSpace,
				},
				Spec: v1alpha1.ChaosResultSpec{
					ExperimentName: "exp-1",
					EngineName:     FakeEngineName,
				},
			},
			isErr: false,
			monitoring: &controller.MonitoringEnabled{
				IsChaosResultsAvailable: true,
			},
		},
		"Test Negative-1": {
			chaosresult: &v1alpha1.ChaosResult{},
			isErr:       true,
			monitoring: &controller.MonitoringEnabled{
				IsChaosResultsAvailable: true,
			},
		},
		"Test Negative-2": {
			isErr:      true,
			monitoring: &controller.MonitoringEnabled{},
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)
			if !mock.isErr {
				_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosResults(mock.chaosresult.Namespace).Create(context.Background(), mock.chaosresult, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("chaosresult not created for %v test, err: %v", name, err)
				}
			}
			resultDetails := &controller.ResultDetails{}
			_, err := resultDetails.GetResultList(client, FakeChaosNameSpace, mock.monitoring)
			if !mock.isErr && err != nil {
				t.Fatalf("test Failed as not able to get the Chaos result list")
			}

		})
	}
}

func TestGetExperimentMetricsFromResult(t *testing.T) {
	FakeEngineName := "Fake Engine"
	FakeNamespace := "Fake Namespace"
	fakeServiceAcc := "Fake Service Account"
	fakeAppLabel := "Fake Label"
	fakeAppKind := "Fake Kind"

	tests := map[string]struct {
		chaosengine     *v1alpha1.ChaosEngine
		chaosresult     *v1alpha1.ChaosResult
		expectedVerdict bool
		isErr           bool
	}{
		"Test Positive-1": {
			chaosengine: &v1alpha1.ChaosEngine{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FakeEngineName,
					Namespace: FakeNamespace,
				},
				Spec: v1alpha1.ChaosEngineSpec{
					ChaosServiceAccount: fakeServiceAcc,
					Appinfo: v1alpha1.ApplicationParams{
						Appns:    FakeNamespace,
						Applabel: fakeAppLabel,
						AppKind:  fakeAppKind,
					},
					Experiments: []v1alpha1.ExperimentList{
						{
							Name: "Fake-Exp-Name",
						},
					},
				},
				Status: v1alpha1.ChaosEngineStatus{
					EngineStatus: v1alpha1.EngineStatusCompleted,
					Experiments: []v1alpha1.ExperimentStatuses{
						{
							Name:   "Fake-Exp-Name",
							Status: v1alpha1.ExperimentStatusRunning,
						},
					},
				},
			},
			chaosresult: &v1alpha1.ChaosResult{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FakeEngineName + "-" + "Fake-Exp-Name",
					Namespace: FakeNamespace,
					UID:       "Fake-UID",
				},
				Spec: v1alpha1.ChaosResultSpec{
					EngineName:     FakeEngineName,
					ExperimentName: "Fake-Exp-Name",
				},
				Status: v1alpha1.ChaosResultStatus{
					ExperimentStatus: v1alpha1.TestStatus{
						Phase:   "Completed",
						Verdict: "Pass",
					},
					History: &v1alpha1.HistoryDetails{},
				},
			},
			isErr: false,
		},
		"Test Negative-1": {
			chaosresult: &v1alpha1.ChaosResult{},
			isErr:       true,
		},
	}

	for name, mock := range tests {
		t.Run(name, func(t *testing.T) {

			client := CreateFakeClient(t)
			resultDetails := &controller.ResultDetails{}
			if !mock.isErr {
				_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(mock.chaosengine.Namespace).Create(context.Background(), mock.chaosengine, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("engine not created for %v test, err: %v", name, err)
				}

				_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosResults(mock.chaosresult.Namespace).Create(context.Background(), mock.chaosresult, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("chaosresult not created for %v test, err: %v", name, err)
				}
			}
			var err error
			_, err = resultDetails.GetExperimentMetricsFromResult(mock.chaosresult, client)
			if !mock.isErr && err != nil {
				t.Fatalf("Test %q failed: expected error to be nil", name)
			}

		})
	}
}

func CreateFakeClient(t *testing.T) clients.ClientSets {
	cs := clients.ClientSets{}
	cs.KubeClient = fake.NewSimpleClientset([]runtime.Object{}...)
	cs.LitmusClient = litmusFakeClientSet.NewSimpleClientset([]runtime.Object{}...)
	return cs
}
