// Code generated by mockery v2.22.1. DO NOT EDIT.

package async

import (
	context "context"
	time "time"

	mock "github.com/stretchr/testify/mock"
)

// MockTask is an autogenerated mock type for the Task type
type MockTask[T interface{}] struct {
	mock.Mock
}

// Cancel provides a mock function with given fields:
func (_m *MockTask[T]) Cancel() {
	_m.Called()
}

// CancelWithReason provides a mock function with given fields: _a0
func (_m *MockTask[T]) CancelWithReason(_a0 error) {
	_m.Called(_a0)
}

// Duration provides a mock function with given fields:
func (_m *MockTask[T]) Duration() time.Duration {
	ret := _m.Called()

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func() time.Duration); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	return r0
}

// Error provides a mock function with given fields:
func (_m *MockTask[T]) Error() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Execute provides a mock function with given fields: ctx
func (_m *MockTask[T]) Execute(ctx context.Context) SilentTask {
	ret := _m.Called(ctx)

	var r0 SilentTask
	if rf, ok := ret.Get(0).(func(context.Context) SilentTask); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(SilentTask)
		}
	}

	return r0
}

// ExecuteSync provides a mock function with given fields: ctx
func (_m *MockTask[T]) ExecuteSync(ctx context.Context) SilentTask {
	ret := _m.Called(ctx)

	var r0 SilentTask
	if rf, ok := ret.Get(0).(func(context.Context) SilentTask); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(SilentTask)
		}
	}

	return r0
}

// Outcome provides a mock function with given fields:
func (_m *MockTask[T]) Outcome() (T, error) {
	ret := _m.Called()

	var r0 T
	var r1 error
	if rf, ok := ret.Get(0).(func() (T, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() T); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(T)
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResultOrDefault provides a mock function with given fields: _a0
func (_m *MockTask[T]) ResultOrDefault(_a0 T) T {
	ret := _m.Called(_a0)

	var r0 T
	if rf, ok := ret.Get(0).(func(T) T); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(T)
	}

	return r0
}

// Run provides a mock function with given fields: ctx
func (_m *MockTask[T]) Run(ctx context.Context) Task[T] {
	ret := _m.Called(ctx)

	var r0 Task[T]
	if rf, ok := ret.Get(0).(func(context.Context) Task[T]); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Task[T])
		}
	}

	return r0
}

// RunSync provides a mock function with given fields: ctx
func (_m *MockTask[T]) RunSync(ctx context.Context) Task[T] {
	ret := _m.Called(ctx)

	var r0 Task[T]
	if rf, ok := ret.Get(0).(func(context.Context) Task[T]); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Task[T])
		}
	}

	return r0
}

// State provides a mock function with given fields:
func (_m *MockTask[T]) State() State {
	ret := _m.Called()

	var r0 State
	if rf, ok := ret.Get(0).(func() State); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(State)
	}

	return r0
}

// Wait provides a mock function with given fields:
func (_m *MockTask[T]) Wait() {
	_m.Called()
}

// WithRecoverAction provides a mock function with given fields: recoverAction
func (_m *MockTask[T]) WithRecoverAction(recoverAction PanicRecoverWork) {
	_m.Called(recoverAction)
}

type mockConstructorTestingTNewMockTask interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockTask creates a new instance of MockTask. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockTask[T interface{}](t mockConstructorTestingTNewMockTask) *MockTask[T] {
	mock := &MockTask[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
