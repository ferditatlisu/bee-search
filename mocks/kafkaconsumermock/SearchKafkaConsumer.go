// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	context "context"

	kafka "github.com/segmentio/kafka-go"
	mock "github.com/stretchr/testify/mock"
)

// SearchKafkaConsumer is an autogenerated mock type for the SearchKafkaConsumer type
type SearchKafkaConsumer struct {
	mock.Mock
}

// ReadMessage provides a mock function with given fields: ctx
func (_m *SearchKafkaConsumer) ReadMessage(ctx context.Context) (kafka.Message, error) {
	ret := _m.Called(ctx)

	var r0 kafka.Message
	if rf, ok := ret.Get(0).(func(context.Context) kafka.Message); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(kafka.Message)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Stop provides a mock function with given fields:
func (_m *SearchKafkaConsumer) Stop() {
	_m.Called()
}

type mockConstructorTestingTNewSearchKafkaConsumer interface {
	mock.TestingT
	Cleanup(func())
}

// NewSearchKafkaConsumer creates a new instance of SearchKafkaConsumer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSearchKafkaConsumer(t mockConstructorTestingTNewSearchKafkaConsumer) *SearchKafkaConsumer {
	mock := &SearchKafkaConsumer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}