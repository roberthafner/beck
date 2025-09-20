package main

import (
	"github.com/stretchr/testify/mock"
	"io"

)

// MockEmailSender is a mock implementation of EmailSender
type MockEmailSender struct {
	mock.Mock
}

// SendEmail mocks the SendEmail method
func (m *MockEmailSender) SendEmail(to string, subject string, body string) error {
	args := m.Called(to, subject, body)
	return args.Error(0)
}

// SendEmailWithAttachment mocks the SendEmailWithAttachment method
func (m *MockEmailSender) SendEmailWithAttachment(to string, subject string, body string, attachment io.Reader) error {
	args := m.Called(to, subject, body, attachment)
	return args.Error(0)
}

// ValidateEmail mocks the ValidateEmail method
func (m *MockEmailSender) ValidateEmail(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}

