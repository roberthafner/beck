package main

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// UserService provides user management functionality
type UserService struct {
	repo   Repository
	logger Logger
	email  EmailSender
	cache  Cache
	client HTTPClient
}

// NewUserService creates a new user service
func NewUserService(repo Repository, logger Logger, email EmailSender, cache Cache, client HTTPClient) *UserService {
	return &UserService{
		repo:   repo,
		logger: logger,
		email:  email,
		cache:  cache,
		client: client,
	}
}

// CreateUser creates a new user in the system
func (s *UserService) CreateUser(ctx context.Context, userID string, userData []byte) error {
	s.logger.Info(fmt.Sprintf("Creating user: %s", userID))

	// Check cache first
	if _, exists := s.cache.Get(userID); exists {
		s.logger.Error("User already exists in cache", fmt.Errorf("duplicate user: %s", userID))
		return fmt.Errorf("user already exists: %s", userID)
	}

	// Save to repository
	if err := s.repo.Save(ctx, userID, userData); err != nil {
		s.logger.Error("Failed to save user to repository", err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Cache the user
	if err := s.cache.Set(userID, userData, 3600); err != nil {
		s.logger.Error("Failed to cache user", err)
		// Don't fail the operation if caching fails
	}

	s.logger.Info(fmt.Sprintf("User created successfully: %s", userID))
	return nil
}

// GetUser retrieves a user from the system
func (s *UserService) GetUser(ctx context.Context, userID string) ([]byte, error) {
	s.logger.Debug(fmt.Sprintf("Retrieving user: %s", userID))

	// Check cache first
	if data, exists := s.cache.Get(userID); exists {
		s.logger.Debug("User found in cache")
		if userData, ok := data.([]byte); ok {
			return userData, nil
		}
	}

	// Load from repository
	data, err := s.repo.Load(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to load user from repository", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Cache the result
	if err := s.cache.Set(userID, data, 3600); err != nil {
		s.logger.Error("Failed to cache user data", err)
		// Continue even if caching fails
	}

	return data, nil
}

// DeleteUser removes a user from the system
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	s.logger.Info(fmt.Sprintf("Deleting user: %s", userID))

	// Delete from repository
	if err := s.repo.Delete(ctx, userID); err != nil {
		s.logger.Error("Failed to delete user from repository", err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Remove from cache
	if err := s.cache.Delete(userID); err != nil {
		s.logger.Error("Failed to remove user from cache", err)
		// Don't fail the operation if cache removal fails
	}

	s.logger.Info(fmt.Sprintf("User deleted successfully: %s", userID))
	return nil
}

// SendWelcomeEmail sends a welcome email to a user
func (s *UserService) SendWelcomeEmail(userID, email string) error {
	s.logger.Info(fmt.Sprintf("Sending welcome email to user: %s", userID))

	if !s.email.ValidateEmail(email) {
		return fmt.Errorf("invalid email address: %s", email)
	}

	subject := "Welcome to our service!"
	body := fmt.Sprintf("Hello! Welcome to our service. Your user ID is: %s", userID)

	if err := s.email.SendEmail(email, subject, body); err != nil {
		s.logger.Error("Failed to send welcome email", err)
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	s.logger.Info("Welcome email sent successfully")
	return nil
}

// SendEmailWithAttachment sends an email with an attachment
func (s *UserService) SendEmailWithAttachment(userID, email, subject, body string, attachment io.Reader) error {
	s.logger.Info(fmt.Sprintf("Sending email with attachment to user: %s", userID))

	if !s.email.ValidateEmail(email) {
		return fmt.Errorf("invalid email address: %s", email)
	}

	if err := s.email.SendEmailWithAttachment(email, subject, body, attachment); err != nil {
		s.logger.Error("Failed to send email with attachment", err)
		return fmt.Errorf("failed to send email with attachment: %w", err)
	}

	s.logger.Info("Email with attachment sent successfully")
	return nil
}

// ListUsers returns a list of user IDs
func (s *UserService) ListUsers(ctx context.Context, prefix string) ([]string, error) {
	s.logger.Debug(fmt.Sprintf("Listing users with prefix: %s", prefix))

	users, err := s.repo.List(ctx, prefix)
	if err != nil {
		s.logger.Error("Failed to list users", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("Found %d users", len(users)))
	return users, nil
}

// GetUserFromAPI retrieves user data from an external API
func (s *UserService) GetUserFromAPI(userID string) (*APIUser, error) {
	s.logger.Info(fmt.Sprintf("Fetching user data from API: %s", userID))

	url := fmt.Sprintf("https://api.example.com/users/%s", userID)

	resp, err := s.client.Get(url)
	if err != nil {
		s.logger.Error("Failed to call API", err)
		return nil, fmt.Errorf("failed to get user from API: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	// Parse response (simplified)
	user := &APIUser{
		ID:   userID,
		Name: string(resp.Body),
	}

	s.logger.Info("User data retrieved from API successfully")
	return user, nil
}

// UpdateUserStatus updates a user's status via API
func (s *UserService) UpdateUserStatus(userID, status string) error {
	s.logger.Info(fmt.Sprintf("Updating user status: %s to %s", userID, status))

	url := fmt.Sprintf("https://api.example.com/users/%s/status", userID)
	data := []byte(fmt.Sprintf(`{"status": "%s"}`, status))

	resp, err := s.client.Put(url, data)
	if err != nil {
		s.logger.Error("Failed to update user status", err)
		return fmt.Errorf("failed to update user status: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	s.logger.Info("User status updated successfully")
	return nil
}

// APIUser represents a user from the external API
type APIUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// FileService provides file management functionality
type FileService struct {
	processor FileProcessor
	logger    Logger
}

// NewFileService creates a new file service
func NewFileService(processor FileProcessor, logger Logger) *FileService {
	return &FileService{
		processor: processor,
		logger:    logger,
	}
}

// ProcessFiles processes multiple files
func (fs *FileService) ProcessFiles(filenames []string) error {
	fs.logger.Info(fmt.Sprintf("Processing %d files", len(filenames)))

	var errors []string

	for _, filename := range filenames {
		if !fs.processor.ValidateFile(filename) {
			errors = append(errors, fmt.Sprintf("invalid file: %s", filename))
			continue
		}

		if err := fs.processor.ProcessFile(filename); err != nil {
			errors = append(errors, fmt.Sprintf("failed to process %s: %v", filename, err))
			continue
		}

		fs.logger.Debug(fmt.Sprintf("Processed file: %s", filename))
	}

	if len(errors) > 0 {
		return fmt.Errorf("processing errors: %s", strings.Join(errors, "; "))
	}

	fs.logger.Info("All files processed successfully")
	return nil
}

// GetFileInfo retrieves information about a file
func (fs *FileService) GetFileInfo(filename string) (*FileInfo, error) {
	fs.logger.Debug(fmt.Sprintf("Getting file info: %s", filename))

	if !fs.processor.ValidateFile(filename) {
		return nil, fmt.Errorf("invalid file: %s", filename)
	}

	info, err := fs.processor.GetFileInfo(filename)
	if err != nil {
		fs.logger.Error("Failed to get file info", err)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return info, nil
}
