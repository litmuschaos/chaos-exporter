package controller

import (
	"errors"
	"testing"

	"github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_unsetDeletedChaosResults(t *testing.T) {

	tests := []struct {
		name           string
		execFunc       func(details *ChaosResultDetails)
		isErr          bool
		resultDetails  *ChaosResultDetails
		oldChaosResult []*v1alpha1.ChaosResult
		newChaosResult []*v1alpha1.ChaosResult
	}{
		{
			name: "success: deleted chaosResult",
			execFunc: func(details *ChaosResultDetails) {
				details.setResultData()
			},
			resultDetails: &ChaosResultDetails{
				UID: "FAKE-UID-OLD",
			},
			oldChaosResult: []*v1alpha1.ChaosResult{
				{
					ObjectMeta: metav1.ObjectMeta{
						UID: "FAKE-UID-OLD",
					},
				},
			},
			newChaosResult: []*v1alpha1.ChaosResult{
				{
					ObjectMeta: metav1.ObjectMeta{
						UID: "FAKE-UID-NEW",
					},
				},
			},
			isErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.execFunc(tt.resultDetails)

			r := MetricesCollecter{}
			r.GaugeMetrics.InitializeGaugeMetrics()
			r.GaugeMetrics.unsetDeletedChaosResults(tt.oldChaosResult, tt.newChaosResult)
			if len(resultStore) != 0 && tt.isErr {
				require.Error(t, errors.New("not able to remove result from resultStore"))
			}
		})
	}

}

func Test_unsetOutdatedChaosResults(t *testing.T) {

	tests := []struct {
		name     string
		execFunc func(details ChaosResultDetails)
		isErr    bool

		oldResultDetails ChaosResultDetails
		newResultDetails ChaosResultDetails
	}{
		{
			name: "success: verdict changed",
			execFunc: func(details ChaosResultDetails) {
				r := &ResultData{}
				matchVerdict[string(details.UID)] = r.setVerdict(details.Verdict)
			},
			oldResultDetails: ChaosResultDetails{
				UID:     "UID",
				Verdict: "Awaited",
			},
			newResultDetails: ChaosResultDetails{
				UID:     "UID",
				Verdict: "Pass",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.execFunc(tt.oldResultDetails)

			r := MetricesCollecter{}
			r.GaugeMetrics.InitializeGaugeMetrics()
			r.GaugeMetrics.unsetOutdatedMetrics(tt.newResultDetails)
		})
	}

}
