package main

import (
	"github.com/stretchr/testify/mock"

)

// MockHTTPClient is a mock implementation of HTTPClient
type MockHTTPClient struct {
	mock.Mock
}

// Get mocks the Get method
func (m *MockHTTPClient) Get(url string) (*HTTPResponse, error) {
	args := m.Called(url)
	return args.Get(0).(*HTTPResponse), args.Error(1)
}

// Post mocks the Post method
func (m *MockHTTPClient) Post(url string, data []byte) (*HTTPResponse, error) {
	args := m.Called(url, data)
	return args.Get(0).(*HTTPResponse), args.Error(1)
}

// Put mocks the Put method
func (m *MockHTTPClient) Put(url string, data []byte) (*HTTPResponse, error) {
	args := m.Called(url, data)
	return args.Get(0).(*HTTPResponse), args.Error(1)
}

// Delete mocks the Delete method
func (m *MockHTTPClient) Delete(url string) (*HTTPResponse, error) {
	args := m.Called(url)
	return args.Get(0).(*HTTPResponse), args.Error(1)
}

