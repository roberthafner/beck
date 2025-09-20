package main

import (
	"github.com/stretchr/testify/mock"

)

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mock.Mock
}

// Info mocks the Info method
func (m *MockLogger) Info(msg string) {
	m.Called(msg)
}

// Error mocks the Error method
func (m *MockLogger) Error(msg string, err error) {
	m.Called(msg, err)
}

// Debug mocks the Debug method
func (m *MockLogger) Debug(msg string) {
	m.Called(msg)
}

// With mocks the With method
func (m *MockLogger) With(key string, value interface{}) Logger {
	args := m.Called(key, value)
	return args.Get(0).(Logger)
}

