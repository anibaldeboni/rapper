// Code generated by mockery v2.40.1. DO NOT EDIT.

package logMocks

import mock "github.com/stretchr/testify/mock"

// LogMessage is an autogenerated mock type for the LogMessage type
type LogMessage struct {
	mock.Mock
}

// String provides a mock function with given fields:
func (_m *LogMessage) String() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for String")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NewLogMessage creates a new instance of LogMessage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLogMessage(t interface {
	mock.TestingT
	Cleanup(func())
}) *LogMessage {
	mock := &LogMessage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}