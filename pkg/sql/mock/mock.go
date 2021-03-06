// Code generated by MockGen. DO NOT EDIT.
// Source: /Users/qi/workspace/go/src/github.com/blademainer/commons/scripts/.././pkg/sql/flatten.go

// Package mock_sql is a generated GoMock package.
package mock_sql

import (
	sql "github.com/blademainer/commons/pkg/sql"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockParser is a mock of Parser interface
type MockParser struct {
	ctrl     *gomock.Controller
	recorder *MockParserMockRecorder
}

// MockParserMockRecorder is the mock recorder for MockParser
type MockParserMockRecorder struct {
	mock *MockParser
}

// NewMockParser creates a new mock instance
func NewMockParser(ctrl *gomock.Controller) *MockParser {
	mock := &MockParser{ctrl: ctrl}
	mock.recorder = &MockParserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockParser) EXPECT() *MockParserMockRecorder {
	return m.recorder
}

// GetFields mocks base method
func (m *MockParser) GetFields() map[string]sql.FieldEntry {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFields")
	ret0, _ := ret[0].(map[string]sql.FieldEntry)
	return ret0
}

// GetFields indicates an expected call of GetFields
func (mr *MockParserMockRecorder) GetFields() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFields", reflect.TypeOf((*MockParser)(nil).GetFields))
}

// GetValueMap mocks base method
func (m *MockParser) GetValueMap(instance interface{}) (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValueMap", instance)
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetValueMap indicates an expected call of GetValueMap
func (mr *MockParserMockRecorder) GetValueMap(instance interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValueMap", reflect.TypeOf((*MockParser)(nil).GetValueMap), instance)
}

// ResolveFieldsFromMap mocks base method
func (m *MockParser) ResolveFieldsFromMap(value map[string]interface{}, out interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResolveFieldsFromMap", value, out)
	ret0, _ := ret[0].(error)
	return ret0
}

// ResolveFieldsFromMap indicates an expected call of ResolveFieldsFromMap
func (mr *MockParserMockRecorder) ResolveFieldsFromMap(value, out interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResolveFieldsFromMap", reflect.TypeOf((*MockParser)(nil).ResolveFieldsFromMap), value, out)
}
