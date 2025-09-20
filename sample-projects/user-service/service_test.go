package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewUserService(t *testing.T) {
	mockRepo := &MockRepository{}
	mockLogger := &MockLogger{}
	mockEmail := &MockEmailSender{}
	mockCache := &MockCache{}
	mockClient := &MockHTTPClient{}

	service := NewUserService(mockRepo, mockLogger, mockEmail, mockCache, mockClient)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
	assert.Equal(t, mockLogger, service.logger)
	assert.Equal(t, mockEmail, service.email)
	assert.Equal(t, mockCache, service.cache)
	assert.Equal(t, mockClient, service.client)
}

func TestNewFileService(t *testing.T) {
	mockProcessor := &MockFileProcessor{}
	mockLogger := &MockLogger{}

	service := NewFileService(mockProcessor, mockLogger)

	assert.NotNil(t, service)
	assert.Equal(t, mockProcessor, service.processor)
	assert.Equal(t, mockLogger, service.logger)
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		data       []byte
		setupMocks func(*MockRepository, *MockLogger, *MockCache)
		wantErr    bool
	}{
		{
			name:   "successful creation",
			userID: "user123",
			data:   []byte("user data"),
			setupMocks: func(repo *MockRepository, logger *MockLogger, cache *MockCache) {
				cache.On("Get", "user123").Return(nil, false)
				repo.On("Save", mock.Anything, "user123", []byte("user data")).Return(nil)
				cache.On("Set", "user123", []byte("user data"), 3600).Return(nil)
				logger.On("Info", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Creating user: user123") ||
						strings.Contains(msg, "User created successfully: user123")
				}))
			},
			wantErr: false,
		},
		{
			name:   "user already exists in cache",
			userID: "user123",
			data:   []byte("user data"),
			setupMocks: func(repo *MockRepository, logger *MockLogger, cache *MockCache) {
				cache.On("Get", "user123").Return([]byte("existing data"), true)
				logger.On("Info", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Creating user: user123")
				}))
				logger.On("Error", mock.AnythingOfType("string"), mock.AnythingOfType("*errors.errorString"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			mockLogger := &MockLogger{}
			mockEmail := &MockEmailSender{}
			mockCache := &MockCache{}
			mockClient := &MockHTTPClient{}

			tt.setupMocks(mockRepo, mockLogger, mockCache)

			service := NewUserService(mockRepo, mockLogger, mockEmail, mockCache, mockClient)

			err := service.CreateUser(context.Background(), tt.userID, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		setupMocks func(*MockRepository, *MockLogger, *MockCache)
		wantData   []byte
		wantErr    bool
	}{
		{
			name:   "user found in cache",
			userID: "user123",
			setupMocks: func(repo *MockRepository, logger *MockLogger, cache *MockCache) {
				cache.On("Get", "user123").Return([]byte("cached data"), true)
				logger.On("Debug", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Retrieving user: user123") ||
						strings.Contains(msg, "User found in cache")
				}))
			},
			wantData: []byte("cached data"),
			wantErr:  false,
		},
		{
			name:   "user not in cache, load from repo",
			userID: "user123",
			setupMocks: func(repo *MockRepository, logger *MockLogger, cache *MockCache) {
				cache.On("Get", "user123").Return(nil, false)
				repo.On("Load", mock.Anything, "user123").Return([]byte("repo data"), nil)
				cache.On("Set", "user123", []byte("repo data"), 3600).Return(nil)
				logger.On("Debug", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Retrieving user: user123")
				}))
			},
			wantData: []byte("repo data"),
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			mockLogger := &MockLogger{}
			mockEmail := &MockEmailSender{}
			mockCache := &MockCache{}
			mockClient := &MockHTTPClient{}

			tt.setupMocks(mockRepo, mockLogger, mockCache)

			service := NewUserService(mockRepo, mockLogger, mockEmail, mockCache, mockClient)

			data, err := service.GetUser(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantData, data)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}

func TestUserService_SendWelcomeEmail(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		email      string
		setupMocks func(*MockEmailSender, *MockLogger)
		wantErr    bool
	}{
		{
			name:   "successful email send",
			userID: "user123",
			email:  "test@example.com",
			setupMocks: func(emailSender *MockEmailSender, logger *MockLogger) {
				emailSender.On("ValidateEmail", "test@example.com").Return(true)
				emailSender.On("SendEmail", "test@example.com", "Welcome to our service!", mock.MatchedBy(func(body string) bool {
					return strings.Contains(body, "user123")
				})).Return(nil)
				logger.On("Info", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Sending welcome email") ||
						strings.Contains(msg, "Welcome email sent successfully")
				}))
			},
			wantErr: false,
		},
		{
			name:   "invalid email",
			userID: "user123",
			email:  "invalid-email",
			setupMocks: func(emailSender *MockEmailSender, logger *MockLogger) {
				emailSender.On("ValidateEmail", "invalid-email").Return(false)
				logger.On("Info", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Sending welcome email")
				}))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockRepository{}
			mockLogger := &MockLogger{}
			mockEmail := &MockEmailSender{}
			mockCache := &MockCache{}
			mockClient := &MockHTTPClient{}

			tt.setupMocks(mockEmail, mockLogger)

			service := NewUserService(mockRepo, mockLogger, mockEmail, mockCache, mockClient)

			err := service.SendWelcomeEmail(tt.userID, tt.email)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockEmail.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestFileService_ProcessFiles(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		setupMocks func(*MockFileProcessor, *MockLogger)
		wantErr    bool
	}{
		{
			name:  "successful processing",
			files: []string{"file1.txt", "file2.txt"},
			setupMocks: func(processor *MockFileProcessor, logger *MockLogger) {
				processor.On("ValidateFile", "file1.txt").Return(true)
				processor.On("ProcessFile", "file1.txt").Return(nil)
				processor.On("ValidateFile", "file2.txt").Return(true)
				processor.On("ProcessFile", "file2.txt").Return(nil)
				logger.On("Info", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Processing 2 files") ||
						strings.Contains(msg, "All files processed successfully")
				}))
				logger.On("Debug", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Processed file:")
				}))
			},
			wantErr: false,
		},
		{
			name:  "invalid file",
			files: []string{"invalid.txt"},
			setupMocks: func(processor *MockFileProcessor, logger *MockLogger) {
				processor.On("ValidateFile", "invalid.txt").Return(false)
				logger.On("Info", mock.MatchedBy(func(msg string) bool {
					return strings.Contains(msg, "Processing 1 files")
				}))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProcessor := &MockFileProcessor{}
			mockLogger := &MockLogger{}

			tt.setupMocks(mockProcessor, mockLogger)

			service := NewFileService(mockProcessor, mockLogger)

			err := service.ProcessFiles(tt.files)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockProcessor.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
