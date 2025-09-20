package main

import (
	"context"
	"io"
)

// Repository represents a data storage interface
type Repository interface {
	Save(ctx context.Context, id string, data []byte) error
	Load(ctx context.Context, id string) ([]byte, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, prefix string) ([]string, error)
}

// Logger represents a logging interface
type Logger interface {
	Info(msg string)
	Error(msg string, err error)
	Debug(msg string)
	With(key string, value interface{}) Logger
}

// EmailSender represents an email sending interface
type EmailSender interface {
	SendEmail(to, subject, body string) error
	SendEmailWithAttachment(to, subject, body string, attachment io.Reader) error
	ValidateEmail(email string) bool
}

// Cache represents a caching interface
type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl int) error
	Delete(key string) error
	Clear() error
	Keys() []string
}

// HTTPClient represents an HTTP client interface
type HTTPClient interface {
	Get(url string) (*HTTPResponse, error)
	Post(url string, data []byte) (*HTTPResponse, error)
	Put(url string, data []byte) (*HTTPResponse, error)
	Delete(url string) (*HTTPResponse, error)
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

// FileProcessor represents a file processing interface
type FileProcessor interface {
	ProcessFile(filename string) error
	ValidateFile(filename string) bool
	GetFileInfo(filename string) (*FileInfo, error)
}

// FileInfo represents file information
type FileInfo struct {
	Name string
	Size int64
	Type string
}
