// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"
	domain "pvz-service-avito-internship/internal/domain"

	mock "github.com/stretchr/testify/mock"

	uuid "github.com/google/uuid"
)

// ProductService is an autogenerated mock type for the ProductService type
type ProductService struct {
	mock.Mock
}

// AddProduct provides a mock function with given fields: ctx, pvzID, productType
func (_m *ProductService) AddProduct(ctx context.Context, pvzID uuid.UUID, productType domain.ProductType) (*domain.Product, error) {
	ret := _m.Called(ctx, pvzID, productType)

	if len(ret) == 0 {
		panic("no return value specified for AddProduct")
	}

	var r0 *domain.Product
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, domain.ProductType) (*domain.Product, error)); ok {
		return rf(ctx, pvzID, productType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, domain.ProductType) *domain.Product); ok {
		r0 = rf(ctx, pvzID, productType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Product)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, domain.ProductType) error); ok {
		r1 = rf(ctx, pvzID, productType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteLastProduct provides a mock function with given fields: ctx, pvzID
func (_m *ProductService) DeleteLastProduct(ctx context.Context, pvzID uuid.UUID) error {
	ret := _m.Called(ctx, pvzID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteLastProduct")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, pvzID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewProductService creates a new instance of ProductService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProductService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProductService {
	mock := &ProductService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
