// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aereal/frontier (interfaces: CloudFrontClient)

// Package frontier_test is a generated GoMock package.
package frontier_test

import (
	context "context"
	cloudfront "github.com/aws/aws-sdk-go-v2/service/cloudfront"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockCloudFrontClient is a mock of CloudFrontClient interface
type MockCloudFrontClient struct {
	ctrl     *gomock.Controller
	recorder *MockCloudFrontClientMockRecorder
}

// MockCloudFrontClientMockRecorder is the mock recorder for MockCloudFrontClient
type MockCloudFrontClientMockRecorder struct {
	mock *MockCloudFrontClient
}

// NewMockCloudFrontClient creates a new mock instance
func NewMockCloudFrontClient(ctrl *gomock.Controller) *MockCloudFrontClient {
	mock := &MockCloudFrontClient{ctrl: ctrl}
	mock.recorder = &MockCloudFrontClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCloudFrontClient) EXPECT() *MockCloudFrontClientMockRecorder {
	return m.recorder
}

// CreateFunction mocks base method
func (m *MockCloudFrontClient) CreateFunction(arg0 context.Context, arg1 *cloudfront.CreateFunctionInput, arg2 ...func(*cloudfront.Options)) (*cloudfront.CreateFunctionOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateFunction", varargs...)
	ret0, _ := ret[0].(*cloudfront.CreateFunctionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateFunction indicates an expected call of CreateFunction
func (mr *MockCloudFrontClientMockRecorder) CreateFunction(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateFunction", reflect.TypeOf((*MockCloudFrontClient)(nil).CreateFunction), varargs...)
}

// GetFunction mocks base method
func (m *MockCloudFrontClient) GetFunction(arg0 context.Context, arg1 *cloudfront.GetFunctionInput, arg2 ...func(*cloudfront.Options)) (*cloudfront.GetFunctionOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "GetFunction", varargs...)
	ret0, _ := ret[0].(*cloudfront.GetFunctionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetFunction indicates an expected call of GetFunction
func (mr *MockCloudFrontClientMockRecorder) GetFunction(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFunction", reflect.TypeOf((*MockCloudFrontClient)(nil).GetFunction), varargs...)
}

// PublishFunction mocks base method
func (m *MockCloudFrontClient) PublishFunction(arg0 context.Context, arg1 *cloudfront.PublishFunctionInput, arg2 ...func(*cloudfront.Options)) (*cloudfront.PublishFunctionOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "PublishFunction", varargs...)
	ret0, _ := ret[0].(*cloudfront.PublishFunctionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PublishFunction indicates an expected call of PublishFunction
func (mr *MockCloudFrontClientMockRecorder) PublishFunction(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishFunction", reflect.TypeOf((*MockCloudFrontClient)(nil).PublishFunction), varargs...)
}

// UpdateFunction mocks base method
func (m *MockCloudFrontClient) UpdateFunction(arg0 context.Context, arg1 *cloudfront.UpdateFunctionInput, arg2 ...func(*cloudfront.Options)) (*cloudfront.UpdateFunctionOutput, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateFunction", varargs...)
	ret0, _ := ret[0].(*cloudfront.UpdateFunctionOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateFunction indicates an expected call of UpdateFunction
func (mr *MockCloudFrontClientMockRecorder) UpdateFunction(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateFunction", reflect.TypeOf((*MockCloudFrontClient)(nil).UpdateFunction), varargs...)
}
