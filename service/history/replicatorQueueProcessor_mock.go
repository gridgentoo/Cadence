// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/uber/cadence/service/history (interfaces: ReplicatorQueueProcessor)

// Package history is a generated GoMock package.
package history

import (
	"context"
	gomock "github.com/golang/mock/gomock"
	replicator "github.com/uber/cadence/.gen/go/replicator"
	reflect "reflect"
)

// MockReplicatorQueueProcessor is a mock of ReplicatorQueueProcessor interface
type MockReplicatorQueueProcessor struct {
	ctrl     *gomock.Controller
	recorder *MockReplicatorQueueProcessorMockRecorder
}

// MockReplicatorQueueProcessorMockRecorder is the mock recorder for MockReplicatorQueueProcessor
type MockReplicatorQueueProcessorMockRecorder struct {
	mock *MockReplicatorQueueProcessor
}

// NewMockReplicatorQueueProcessor creates a new mock instance
func NewMockReplicatorQueueProcessor(ctrl *gomock.Controller) *MockReplicatorQueueProcessor {
	mock := &MockReplicatorQueueProcessor{ctrl: ctrl}
	mock.recorder = &MockReplicatorQueueProcessorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockReplicatorQueueProcessor) EXPECT() *MockReplicatorQueueProcessorMockRecorder {
	return m.recorder
}

// Start mocks base method
func (m *MockReplicatorQueueProcessor) Start() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Start")
}

// Start indicates an expected call of Start
func (mr *MockReplicatorQueueProcessorMockRecorder) Start() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockReplicatorQueueProcessor)(nil).Start))
}

// Stop mocks base method
func (m *MockReplicatorQueueProcessor) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop
func (mr *MockReplicatorQueueProcessorMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockReplicatorQueueProcessor)(nil).Stop))
}

// getTasks mocks base method
func (m *MockReplicatorQueueProcessor) getTasks(arg0 context.Context, arg1 int64) (*replicator.ReplicationMessages, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "getTasks", arg0, arg1)
	ret0, _ := ret[0].(*replicator.ReplicationMessages)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// getTasks indicates an expected call of getTasks
func (mr *MockReplicatorQueueProcessorMockRecorder) getTasks(arg0 interface{}, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "getTasks", reflect.TypeOf((*MockReplicatorQueueProcessor)(nil).getTasks), arg0, arg1)
}

// notifyNewTask mocks base method
func (m *MockReplicatorQueueProcessor) notifyNewTask() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "notifyNewTask")
}

// notifyNewTask indicates an expected call of notifyNewTask
func (mr *MockReplicatorQueueProcessorMockRecorder) notifyNewTask() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "notifyNewTask", reflect.TypeOf((*MockReplicatorQueueProcessor)(nil).notifyNewTask))
}
