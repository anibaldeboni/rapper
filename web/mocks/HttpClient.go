// Code generated by mockery v2.40.1. DO NOT EDIT.

package mocks

import (
	io "io"
	web "rapper/web"

	mock "github.com/stretchr/testify/mock"
)

// HttpClient is an autogenerated mock type for the HttpClient type
type HttpClient struct {
	mock.Mock
}

// Get provides a mock function with given fields: url, headers
func (_m *HttpClient) Get(url string, headers map[string]string) (web.Response, error) {
	ret := _m.Called(url, headers)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 web.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, map[string]string) (web.Response, error)); ok {
		return rf(url, headers)
	}
	if rf, ok := ret.Get(0).(func(string, map[string]string) web.Response); ok {
		r0 = rf(url, headers)
	} else {
		r0 = ret.Get(0).(web.Response)
	}

	if rf, ok := ret.Get(1).(func(string, map[string]string) error); ok {
		r1 = rf(url, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Post provides a mock function with given fields: url, body, headers
func (_m *HttpClient) Post(url string, body io.Reader, headers map[string]string) (web.Response, error) {
	ret := _m.Called(url, body, headers)

	if len(ret) == 0 {
		panic("no return value specified for Post")
	}

	var r0 web.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader, map[string]string) (web.Response, error)); ok {
		return rf(url, body, headers)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader, map[string]string) web.Response); ok {
		r0 = rf(url, body, headers)
	} else {
		r0 = ret.Get(0).(web.Response)
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader, map[string]string) error); ok {
		r1 = rf(url, body, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Put provides a mock function with given fields: url, body, headers
func (_m *HttpClient) Put(url string, body io.Reader, headers map[string]string) (web.Response, error) {
	ret := _m.Called(url, body, headers)

	if len(ret) == 0 {
		panic("no return value specified for Put")
	}

	var r0 web.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(string, io.Reader, map[string]string) (web.Response, error)); ok {
		return rf(url, body, headers)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader, map[string]string) web.Response); ok {
		r0 = rf(url, body, headers)
	} else {
		r0 = ret.Get(0).(web.Response)
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader, map[string]string) error); ok {
		r1 = rf(url, body, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewHttpClient creates a new instance of HttpClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHttpClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *HttpClient {
	mock := &HttpClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}