package main

import (
	"github.com/stretchr/testify/mock"

)

// MockFileProcessor is a mock implementation of FileProcessor
type MockFileProcessor struct {
	mock.Mock
}

// ProcessFile mocks the ProcessFile method
func (m *MockFileProcessor) ProcessFile(filename string) error {
	args := m.Called(filename)
	return args.Error(0)
}

// ValidateFile mocks the ValidateFile method
func (m *MockFileProcessor) ValidateFile(filename string) bool {
	args := m.Called(filename)
	return args.Bool(0)
}

// GetFileInfo mocks the GetFileInfo method
func (m *MockFileProcessor) GetFileInfo(filename string) (*FileInfo, error) {
	args := m.Called(filename)
	return args.Get(0).(*FileInfo), args.Error(1)
}

