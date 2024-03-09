// Code generated by mockery v2.40.1. DO NOT EDIT.

package outputMocks

import (
	output "github.com/anibaldeboni/rapper/internal/output"
	mock "github.com/stretchr/testify/mock"
)

// Stream is an autogenerated mock type for the Stream type
type Stream struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Stream) Close() {
	_m.Called()
}

// Enabled provides a mock function with given fields:
func (_m *Stream) Enabled() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Enabled")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Send provides a mock function with given fields: _a0
func (_m *Stream) Send(_a0 output.Message) {
	_m.Called(_a0)
}

// NewStream creates a new instance of Stream. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStream(t interface {
	mock.TestingT
	Cleanup(func())
}) *Stream {
	mock := &Stream{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
