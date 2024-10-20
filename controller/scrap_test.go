package controller_test

import (
	"testing"

	"github.com/litmuschaos/chaos-exporter/controller"
	"github.com/litmuschaos/chaos-exporter/controller/mocks"
	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		overallChaosResult []*v1alpha1.ChaosResult
	}{
		{
			name: "success",
			execFunc: func() {
				mockCollectData.EXPECT().GetResultList(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*v1alpha1.ChaosResult{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "chaosresult-1",
							},
						},
					}, nil).Times(1)
				mockCollectData.EXPECT().GetExperimentMetricsFromResult(gomock.Any(), gomock.Any()).Return(false, nil).Times(1)
				mockCollectData.EXPECT().SetResultDetails()
				mockCollectData.EXPECT().GetResultDetails().Return(controller.ChaosResultDetails{
					UID: "FAKE-UID",
				}).Times(1)
			},
			overallChaosResult: []*v1alpha1.ChaosResult{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "chaosresult-1",
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
					Return([]*v1alpha1.ChaosResult{}, errors.New("Fake Error")).Times(1)
			},
			overallChaosResult: []*v1alpha1.ChaosResult{},
			monitoring:         &controller.MonitoringEnabled{},
			isErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.execFunc()

			client := CreateFakeClient(t)
			_, err := r.GetLitmusChaosMetrics(client, tt.overallChaosResult, tt.monitoring)
			if tt.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}

}
