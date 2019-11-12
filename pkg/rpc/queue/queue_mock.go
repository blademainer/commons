// Code generated by MockGen. DO NOT EDIT.
// Source: queue.go

// Package queue is a generated GoMock package.
package queue

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockQueue is a mock of Queue interface
type MockQueue struct {
	ctrl     *gomock.Controller
	recorder *MockQueueMockRecorder
}

// MockQueueMockRecorder is the mock recorder for MockQueue
type MockQueueMockRecorder struct {
	mock *MockQueue
}

// NewMockQueue creates a new mock instance
func NewMockQueue(ctrl *gomock.Controller) *MockQueue {
	mock := &MockQueue{ctrl: ctrl}
	mock.recorder = &MockQueueMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockQueue) EXPECT() *MockQueueMockRecorder {
	return m.recorder
}

// Produce mocks base method
func (m *MockQueue) Produce(payload []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Produce", payload)
	ret0, _ := ret[0].(error)
	return ret0
}

// Produce indicates an expected call of Produce
func (mr *MockQueueMockRecorder) Produce(payload interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Produce", reflect.TypeOf((*MockQueue)(nil).Produce), payload)
}

// Consume mocks base method
func (m *MockQueue) Consume() ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Consume")
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Consume indicates an expected call of Consume
func (mr *MockQueueMockRecorder) Consume() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Consume", reflect.TypeOf((*MockQueue)(nil).Consume))
}
