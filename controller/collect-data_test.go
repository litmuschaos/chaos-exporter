package controller_test

import (
	"context"
	"testing"

	"github.com/litmuschaos/chaos-exporter/controller"
	"github.com/litmuschaos/chaos-exporter/pkg/clients"
	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	litmusFakeClientSet "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/util/workqueue"
)

func TestGetResultList(t *testing.T) {
	FakeChaosNameSpace := "Fake Namespace"
	FakeEngineName := "Fake Engine"

	tests := []struct {
		name        string
		execFunc    func(client clients.ClientSets, chaosResult *v1alpha1.ChaosResult)
		chaosResult *v1alpha1.ChaosResult
		monitoring  *controller.MonitoringEnabled
		isErr       bool
	}{
		{
			name: "success:chaos result found",
			execFunc: func(client clients.ClientSets, chaosResult *v1alpha1.ChaosResult) {
				_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosResults(chaosResult.Namespace).Create(context.Background(), chaosResult, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("chaosresult not created")
				}
			},

			chaosResult: &v1alpha1.ChaosResult{
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
		{
			name:        "success:empty chaosResult",
			chaosResult: &v1alpha1.ChaosResult{},
			execFunc:    func(client clients.ClientSets, chaosResult *v1alpha1.ChaosResult) {},
			isErr:       false,
			monitoring: &controller.MonitoringEnabled{
				IsChaosResultsAvailable: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := CreateFakeClient(t)
			tt.execFunc(client, tt.chaosResult)
			resultDetails := &controller.ResultDetails{}
			_, err := resultDetails.GetResultList(client, FakeChaosNameSpace, tt.monitoring)
			if tt.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
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
		verdict         bool
		execFunc        func(client clients.ClientSets, engine *v1alpha1.ChaosEngine, result *v1alpha1.ChaosResult)
	}{
		"success": {
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

			execFunc: func(client clients.ClientSets, engine *v1alpha1.ChaosEngine, result *v1alpha1.ChaosResult) {
				_, err := client.LitmusClient.LitmuschaosV1alpha1().ChaosEngines(engine.Namespace).Create(context.Background(), engine, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("engine not created for test, err: %v", err)
				}

				_, err = client.LitmusClient.LitmuschaosV1alpha1().ChaosResults(result.Namespace).Create(context.Background(), result, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("chaosresult not created fortest, err: %v", err)
				}
			},
			isErr:   false,
			verdict: true,
		},
		"failure: No Chaos Engine": {
			chaosresult: &v1alpha1.ChaosResult{},
			isErr:       false,
			verdict:     true,
			execFunc: func(client clients.ClientSets, engine *v1alpha1.ChaosEngine, result *v1alpha1.ChaosResult) {
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := CreateFakeClient(t)
			resultDetails := controller.ResultDetails{}
			tt.execFunc(client, tt.chaosengine, tt.chaosresult)
			verdict, err := resultDetails.GetExperimentMetricsFromResult(tt.chaosresult, client)
			assert.Equal(t, tt.verdict, verdict)
			if tt.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func CreateFakeClient(t *testing.T) clients.ClientSets {
	cs := clients.ClientSets{}
	cs.KubeClient = fake.NewSimpleClientset([]runtime.Object{}...)
	cs.LitmusClient = litmusFakeClientSet.NewSimpleClientset([]runtime.Object{}...)
	stopCh := make(chan struct{})
	err := cs.SetupInformers(stopCh, cs.KubeClient, cs.LitmusClient, 0, workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()))
	require.NoError(t, err)
	return cs
}
