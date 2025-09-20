package main

import (
	"github.com/stretchr/testify/mock"
)

// MockCache is a mock implementation of Cache
type MockCache struct {
	mock.Mock
}

// Get mocks the Get method
func (m *MockCache) Get(key string) (interface{}, bool) {
	args := m.Called(key)
	return args.Get(0), args.Bool(1)
}

// Set mocks the Set method
func (m *MockCache) Set(key string, value interface{}, ttl int) error {
	args := m.Called(key, value, ttl)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// Clear mocks the Clear method
func (m *MockCache) Clear() error {
	args := m.Called()
	return args.Error(0)
}

// Keys mocks the Keys method
func (m *MockCache) Keys() []string {
	args := m.Called()
	return args.Get(0).([]string)
}
