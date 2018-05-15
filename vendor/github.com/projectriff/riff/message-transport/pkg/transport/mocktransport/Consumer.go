// Code generated by mockery v1.0.0
package mocktransport

import message "github.com/projectriff/riff/message-transport/pkg/message"
import mock "github.com/stretchr/testify/mock"

// Consumer is an autogenerated mock type for the Consumer type
type Consumer struct {
	mock.Mock
}

// Messages provides a mock function with given fields:
func (_m *Consumer) Messages() <-chan message.Message {
	ret := _m.Called()

	var r0 <-chan message.Message
	if rf, ok := ret.Get(0).(func() <-chan message.Message); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan message.Message)
		}
	}

	return r0
}
