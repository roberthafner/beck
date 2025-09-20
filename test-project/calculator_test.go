package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCalculator(t *testing.T) {
	calc := NewCalculator("Test Calculator")
	assert.NotNil(t, calc)
	assert.Equal(t, "Test Calculator", calc.GetName())
}

func TestCalculator_Add(t *testing.T) {
	calc := NewCalculator("Test")
	result := calc.Add(2, 3)
	assert.Equal(t, 5.0, result)
}

func TestCalculator_Subtract(t *testing.T) {
	calc := NewCalculator("Test")
	result := calc.Subtract(10, 4)
	assert.Equal(t, 6.0, result)
}

func TestCalculator_Multiply(t *testing.T) {
	calc := NewCalculator("Test")
	result := calc.Multiply(5, 6)
	assert.Equal(t, 30.0, result)
}

func TestCalculator_Divide(t *testing.T) {
	calc := NewCalculator("Test")

	// Test normal division
	result, err := calc.Divide(15, 3)
	assert.NoError(t, err)
	assert.Equal(t, 5.0, result)

	// Test division by zero
	_, err = calc.Divide(10, 0)
	assert.Error(t, err)
	assert.Equal(t, "division by zero", err.Error())
}

func TestIsEven(t *testing.T) {
	assert.True(t, IsEven(4))
	assert.False(t, IsEven(5))
	assert.True(t, IsEven(0))
}

func TestFactorial(t *testing.T) {
	// Test normal cases
	result, err := Factorial(5)
	assert.NoError(t, err)
	assert.Equal(t, 120, result)

	result, err = Factorial(0)
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	// Test error case
	_, err = Factorial(-1)
	assert.Error(t, err)
	assert.Equal(t, "factorial is not defined for negative numbers", err.Error())
}

// Note: Power, Sqrt, Fibonacci, IsPrime, Max, Min, and Abs functions are intentionally not tested
// to demonstrate coverage gaps
