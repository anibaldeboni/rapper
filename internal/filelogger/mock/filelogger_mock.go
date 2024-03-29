// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/anibaldeboni/rapper/internal/filelogger (interfaces: FileLogger)
//
// Generated by this command:
//
//	mockgen -destination mock/filelogger_mock.go github.com/anibaldeboni/rapper/internal/filelogger FileLogger
//

// Package mock_filelogger is a generated GoMock package.
package mock_filelogger

import (
	reflect "reflect"

	filelogger "github.com/anibaldeboni/rapper/internal/filelogger"
	gomock "go.uber.org/mock/gomock"
)

// MockFileLogger is a mock of FileLogger interface.
type MockFileLogger struct {
	ctrl     *gomock.Controller
	recorder *MockFileLoggerMockRecorder
}

// MockFileLoggerMockRecorder is the mock recorder for MockFileLogger.
type MockFileLoggerMockRecorder struct {
	mock *MockFileLogger
}

// NewMockFileLogger creates a new mock instance.
func NewMockFileLogger(ctrl *gomock.Controller) *MockFileLogger {
	mock := &MockFileLogger{ctrl: ctrl}
	mock.recorder = &MockFileLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFileLogger) EXPECT() *MockFileLoggerMockRecorder {
	return m.recorder
}

// Write mocks base method.
func (m *MockFileLogger) Write(arg0 filelogger.Line) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Write", arg0)
}

// Write indicates an expected call of Write.
func (mr *MockFileLoggerMockRecorder) Write(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockFileLogger)(nil).Write), arg0)
}
