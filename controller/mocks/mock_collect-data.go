// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/litmuschaos/chaos-exporter/controller (interfaces: ResultCollector)

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	controller "github.com/litmuschaos/chaos-exporter/controller"
	clients "github.com/litmuschaos/chaos-exporter/pkg/clients"
	v1alpha1 "github.com/litmuschaos/chaos-operator/api/litmuschaos/v1alpha1"
)

// MockResultCollector is a mock of ResultCollector interface.
type MockResultCollector struct {
	ctrl     *gomock.Controller
	recorder *MockResultCollectorMockRecorder
}

// MockResultCollectorMockRecorder is the mock recorder for MockResultCollector.
type MockResultCollectorMockRecorder struct {
	mock *MockResultCollector
}

// NewMockResultCollector creates a new mock instance.
func NewMockResultCollector(ctrl *gomock.Controller) *MockResultCollector {
	mock := &MockResultCollector{ctrl: ctrl}
	mock.recorder = &MockResultCollectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockResultCollector) EXPECT() *MockResultCollectorMockRecorder {
	return m.recorder
}

// GetExperimentMetricsFromResult mocks base method.
func (m *MockResultCollector) GetExperimentMetricsFromResult(arg0 *v1alpha1.ChaosResult, arg1 clients.ClientSets) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetExperimentMetricsFromResult", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetExperimentMetricsFromResult indicates an expected call of GetExperimentMetricsFromResult.
func (mr *MockResultCollectorMockRecorder) GetExperimentMetricsFromResult(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetExperimentMetricsFromResult", reflect.TypeOf((*MockResultCollector)(nil).GetExperimentMetricsFromResult), arg0, arg1)
}

// GetResultDetails mocks base method.
func (m *MockResultCollector) GetResultDetails() controller.ChaosResultDetails {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResultDetails")
	ret0, _ := ret[0].(controller.ChaosResultDetails)
	return ret0
}

// GetResultDetails indicates an expected call of GetResultDetails.
func (mr *MockResultCollectorMockRecorder) GetResultDetails() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResultDetails", reflect.TypeOf((*MockResultCollector)(nil).GetResultDetails))
}

// GetResultList mocks base method.
func (m *MockResultCollector) GetResultList(arg0 clients.ClientSets, arg1 string, arg2 *controller.MonitoringEnabled) ([]*v1alpha1.ChaosResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetResultList", arg0, arg1, arg2)
	ret0, _ := ret[0].([]*v1alpha1.ChaosResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetResultList indicates an expected call of GetResultList.
func (mr *MockResultCollectorMockRecorder) GetResultList(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetResultList", reflect.TypeOf((*MockResultCollector)(nil).GetResultList), arg0, arg1, arg2)
}

// SetResultDetails mocks base method.
func (m *MockResultCollector) SetResultDetails() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetResultDetails")
}

// SetResultDetails indicates an expected call of SetResultDetails.
func (mr *MockResultCollectorMockRecorder) SetResultDetails() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetResultDetails", reflect.TypeOf((*MockResultCollector)(nil).SetResultDetails))
}
