// Code generated by mockery v2.15.0. DO NOT EDIT.

package promising

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockIOutcome is an autogenerated mock type for the IOutcome type
type MockIOutcome[V interface{}] struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx
func (_m *MockIOutcome[V]) Get(ctx context.Context) (V, error) {
	ret := _m.Called(ctx)

	var r0 V
	if rf, ok := ret.Get(0).(func(context.Context) V); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(V)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// cancel provides a mock function with given fields:
func (_m *MockIOutcome[V]) cancel() {
	_m.Called()
}

// complete provides a mock function with given fields: val, err
func (_m *MockIOutcome[V]) complete(val V, err error) {
	_m.Called(val, err)
}

type mockConstructorTestingTNewMockIOutcome interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockIOutcome creates a new instance of MockIOutcome. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockIOutcome[V interface{}](t mockConstructorTestingTNewMockIOutcome) *MockIOutcome[V] {
	mock := &MockIOutcome[V]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
