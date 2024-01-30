package controller_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/litmuschaos/chaos-exporter/controller"
	"github.com/litmuschaos/chaos-exporter/controller/mocks"
	"github.com/litmuschaos/chaos-exporter/pkg/log"
)

func TestExporter(t *testing.T) {

	mockClient := CreateFakeClient(t)
	log.Info("Started creating Metrics")
	log.Info("Registering Fixed Metrics")

	//Creating Mock MetricesCollecter
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCollectData := mocks.NewMockResultCollector(mockCtl)

	r := controller.MetricesCollecter{
		ResultCollector: mockCollectData,
	}

	//Chaos Result List
	mockOverallChaosResults := v1alpha1.ChaosResultList{
		Items: []v1alpha1.ChaosResult{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "chaosresult-1",
				},
			},
		},
	}

	// Register Register Fixed Metrics
	r.GaugeMetrics.InitializeGaugeMetrics().RegisterFixedMetrics()

	// Enable Monitoring
	monitoringEnabled := controller.MonitoringEnabled{
		IsChaosResultsAvailable: true,
		IsChaosEnginesAvailable: true,
	}
	fmt.Println("Before Loop")
	// Running the unit test for GetLitmusChaosMetrics
	for {
		assert.NoError(t, r.GetLitmusChaosMetrics(mockClient, &mockOverallChaosResults, &monitoringEnabled))
		/*
		 Unit test GetLitmusChaosMetrics
		*/
		time.Sleep(1000 * time.Millisecond)
	}

}
