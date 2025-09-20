package main

import (
	"github.com/stretchr/testify/mock"
	"context"

)

// MockRepository is a mock implementation of Repository
type MockRepository struct {
	mock.Mock
}

// Save mocks the Save method
func (m *MockRepository) Save(ctx context.Context, id string, data []byte) error {
	args := m.Called(ctx, id, data)
	return args.Error(0)
}

// Load mocks the Load method
func (m *MockRepository) Load(ctx context.Context, id string) ([]byte, error) {
	args := m.Called(ctx, id)
	return args.Get(0).([]byte), args.Error(1)
}

// Delete mocks the Delete method
func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// List mocks the List method
func (m *MockRepository) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

