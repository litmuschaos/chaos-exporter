package controller_test

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/litmuschaos/chaos-exporter/controller"
	"github.com/litmuschaos/chaos-exporter/controller/mocks"
	v1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
)

func TestGetLitmusChaosMetrics(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCollectData := mocks.NewMockResultCollector(mockCtl)

	//FakeEngineName := "Fake Engine"
	//FakeNamespace := "Fake Namespace"
	//fakeServiceAcc := "Fake Service Account"
	//fakeAppLabel := "Fake Label"
	//FakeAppName := "Fake App"
	//FakeClusterName := "Fake Cluster"

	tests := []struct {
		name               string
		execFunc           func()
		chaosengine        *v1alpha1.ChaosEngine
		chaosresult        *v1alpha1.ChaosResult
		isErr              bool
		monitoring         *controller.MonitoringEnabled
		overallChaosResult *v1alpha1.ChaosResultList
	}{
		{
			name: "Test Positive-1",
			execFunc: func() {
				mockCollectData.EXPECT().GetResultList(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(v1alpha1.ChaosResultList{
						Items: []v1alpha1.ChaosResult{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name: "chaosresult-1",
								},
							},
						},
					}, nil).Times(1)
				mockCollectData.EXPECT().GetExperimentMetricsFromResult(gomock.Any(), gomock.Any()).Return(false, nil).Times(1)

				os.Setenv("AWS_CLOUDWATCH_METRIC_NAMESPACE", "")
				os.Setenv("CLUSTER_NAME", "")
				os.Setenv("APP_NAME", "")
				os.Setenv("WATCH_NAMESPACE", "")
			},
			overallChaosResult: &v1alpha1.ChaosResultList{
				Items: []v1alpha1.ChaosResult{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "chaosresult-1",
						},
					},
				},
			},
			monitoring: &controller.MonitoringEnabled{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.execFunc()

			client := CreateFakeClient(t)
			r := controller.MetricesCollecter{
				ResultCollector: mockCollectData,
			}

			r.GaugeMetrics.InitializeGaugeMetrics().RegisterFixedMetrics()
			err := r.GetLitmusChaosMetrics(client, tt.overallChaosResult, tt.monitoring)
			fmt.Print(err)
		})
	}

}
