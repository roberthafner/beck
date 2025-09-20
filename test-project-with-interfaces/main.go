package main

import (
	"context"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Interface Test Project")

	// This is a simple main function to make the project buildable
	// In a real application, these would be properly initialized
	// with concrete implementations

	ctx := context.Background()

	// Example usage of the services (with nil implementations for demo)
	// In real code, these would be proper implementations
	var repo Repository
	var logger Logger
	var email EmailSender
	var cache Cache
	var client HTTPClient
	var processor FileProcessor

	// These will be nil, but the functions are still testable
	userService := NewUserService(repo, logger, email, cache, client)
	fileService := NewFileService(processor, logger)

	// Print service info
	if userService != nil {
		fmt.Println("UserService created")
	}

	if fileService != nil {
		fmt.Println("FileService created")
	}

	// Example of functions that could be called (but would panic due to nil interfaces)
	// These are here to show the function signatures
	_ = ctx

	fmt.Println("Available UserService methods:")
	fmt.Println("- CreateUser(ctx, userID, userData)")
	fmt.Println("- GetUser(ctx, userID)")
	fmt.Println("- DeleteUser(ctx, userID)")
	fmt.Println("- SendWelcomeEmail(userID, email)")
	fmt.Println("- SendEmailWithAttachment(userID, email, subject, body, attachment)")
	fmt.Println("- ListUsers(ctx, prefix)")
	fmt.Println("- GetUserFromAPI(userID)")
	fmt.Println("- UpdateUserStatus(userID, status)")

	fmt.Println("\nAvailable FileService methods:")
	fmt.Println("- ProcessFiles(filenames)")
	fmt.Println("- GetFileInfo(filename)")

	// Utility functions that don't depend on interfaces
	fmt.Println("\nUtility functions:")
	result := AddNumbers(5, 3)
	fmt.Printf("AddNumbers(5, 3) = %d\n", result)

	even := IsEvenNumber(4)
	fmt.Printf("IsEvenNumber(4) = %t\n", even)

	reversed := ReverseString("hello")
	fmt.Printf("ReverseString(\"hello\") = %s\n", reversed)

	fmt.Println("\nThis project demonstrates interfaces that will be mocked during test generation.")
}

// AddNumbers adds two integers
func AddNumbers(a, b int) int {
	return a + b
}

// IsEvenNumber checks if a number is even
func IsEvenNumber(n int) bool {
	return n%2 == 0
}

// ReverseString reverses a string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// ValidateInput validates user input
func ValidateInput(input string) error {
	if strings.TrimSpace(input) == "" {
		return fmt.Errorf("input cannot be empty")
	}

	if len(input) > 100 {
		return fmt.Errorf("input too long: maximum 100 characters")
	}

	return nil
}

// ParseConfig parses configuration from environment
func ParseConfig() *Config {
	config := &Config{
		Host:     getEnvOrDefault("HOST", "localhost"),
		Port:     getEnvOrDefault("PORT", "8080"),
		Database: getEnvOrDefault("DATABASE", "sqlite"),
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
	}

	return config
}

// Config represents application configuration
type Config struct {
	Host     string
	Port     string
	Database string
	LogLevel string
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CalculateTotal calculates total with tax
func CalculateTotal(amount float64, taxRate float64) float64 {
	if amount < 0 {
		return 0
	}

	tax := amount * taxRate
	return amount + tax
}

// FormatCurrency formats a number as currency
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}
