package controller_test

import (
	"github.com/golang/mock/gomock"
	"github.com/litmuschaos/chaos-exporter/controller"
	"github.com/litmuschaos/chaos-exporter/controller/mocks"
	v1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetLitmusChaosMetrics(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCollectData := mocks.NewMockResultCollector(mockCtl)

	r := controller.MetricesCollecter{
		ResultCollector: mockCollectData,
	}

	r.GaugeMetrics.InitializeGaugeMetrics().RegisterFixedMetrics()

	tests := []struct {
		name               string
		execFunc           func()
		isErr              bool
		monitoring         *controller.MonitoringEnabled
		overallChaosResult *v1alpha1.ChaosResultList
	}{
		{
			name: "success",
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
				mockCollectData.EXPECT().SetResultDetails()
				mockCollectData.EXPECT().GetResultDetails().Return(controller.ChaosResultDetails{
					UID: "FAKE-UID",
				}).Times(1)
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
			isErr:      false,
		},
		{
			name: "failure: no ChaosResultList found",
			execFunc: func() {
				mockCollectData.EXPECT().GetResultList(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(v1alpha1.ChaosResultList{}, errors.New("Fake Error")).Times(1)
			},
			overallChaosResult: &v1alpha1.ChaosResultList{},
			monitoring:         &controller.MonitoringEnabled{},
			isErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.execFunc()

			client := CreateFakeClient(t)
			err := r.GetLitmusChaosMetrics(client, tt.overallChaosResult, tt.monitoring)
			if tt.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}

}
