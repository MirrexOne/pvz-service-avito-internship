// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// UnsafePVZServiceServer is an autogenerated mock type for the UnsafePVZServiceServer type
type UnsafePVZServiceServer struct {
	mock.Mock
}

// mustEmbedUnimplementedPVZServiceServer provides a mock function with no fields
func (_m *UnsafePVZServiceServer) mustEmbedUnimplementedPVZServiceServer() {
	_m.Called()
}

// NewUnsafePVZServiceServer creates a new instance of UnsafePVZServiceServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUnsafePVZServiceServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *UnsafePVZServiceServer {
	mock := &UnsafePVZServiceServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
