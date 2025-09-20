package main

import (
	"errors"
	"fmt"
	"math"
)

// Calculator provides basic arithmetic operations
type Calculator struct {
	memory float64
	history []string
}

// NewCalculator creates a new calculator instance
func NewCalculator() *Calculator {
	return &Calculator{
		memory:  0,
		history: make([]string, 0),
	}
}

// Add performs addition and returns the result
func (c *Calculator) Add(a, b float64) float64 {
	result := a + b
	c.addToHistory(fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result))
	return result
}

// Subtract performs subtraction and returns the result
func (c *Calculator) Subtract(a, b float64) float64 {
	result := a - b
	c.addToHistory(fmt.Sprintf("%.2f - %.2f = %.2f", a, b, result))
	return result
}

// Multiply performs multiplication and returns the result
func (c *Calculator) Multiply(a, b float64) float64 {
	result := a * b
	c.addToHistory(fmt.Sprintf("%.2f * %.2f = %.2f", a, b, result))
	return result
}

// Divide performs division and returns the result with error handling
func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	
	result := a / b
	c.addToHistory(fmt.Sprintf("%.2f / %.2f = %.2f", a, b, result))
	return result, nil
}

// Power calculates a to the power of b
func (c *Calculator) Power(a, b float64) float64 {
	result := math.Pow(a, b)
	c.addToHistory(fmt.Sprintf("%.2f ^ %.2f = %.2f", a, b, result))
	return result
}

// Sqrt calculates square root
func (c *Calculator) Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, errors.New("square root of negative number")
	}
	
	result := math.Sqrt(a)
	c.addToHistory(fmt.Sprintf("√%.2f = %.2f", a, result))
	return result, nil
}

// ComplexOperation performs a complex calculation with multiple branches
func (c *Calculator) ComplexOperation(x, y, z float64, operation string) (float64, error) {
	var result float64
	var err error
	
	switch operation {
	case "sum":
		if x > 0 && y > 0 && z > 0 {
			result = x + y + z
		} else if x < 0 || y < 0 || z < 0 {
			// Handle negative numbers differently
			result = math.Abs(x) + math.Abs(y) + math.Abs(z)
		} else {
			result = 0
		}
	case "product":
		result = x * y * z
		if result > 1000000 {
			return 0, errors.New("result too large")
		}
	case "average":
		if x == 0 && y == 0 && z == 0 {
			return 0, errors.New("cannot average all zeros")
		}
		result = (x + y + z) / 3
	case "max":
		result = math.Max(x, math.Max(y, z))
	case "min":
		result = math.Min(x, math.Min(y, z))
	default:
		return 0, errors.New("unknown operation")
	}
	
	c.addToHistory(fmt.Sprintf("complex(%s): %.2f, %.2f, %.2f = %.2f", operation, x, y, z, result))
	return result, err
}

// GetMemory returns the current memory value
func (c *Calculator) GetMemory() float64 {
	return c.memory
}

// SetMemory sets the memory value
func (c *Calculator) SetMemory(value float64) {
	c.memory = value
	c.addToHistory(fmt.Sprintf("memory set to %.2f", value))
}

// AddToMemory adds a value to memory
func (c *Calculator) AddToMemory(value float64) {
	c.memory += value
	c.addToHistory(fmt.Sprintf("added %.2f to memory (now %.2f)", value, c.memory))
}

// ClearMemory resets memory to zero
func (c *Calculator) ClearMemory() {
	c.memory = 0
	c.addToHistory("memory cleared")
}

// GetHistory returns the calculation history
func (c *Calculator) GetHistory() []string {
	return c.history
}

// ClearHistory clears the calculation history
func (c *Calculator) ClearHistory() {
	c.history = make([]string, 0)
}

// GetLastResult returns the last calculation result from history
func (c *Calculator) GetLastResult() (string, error) {
	if len(c.history) == 0 {
		return "", errors.New("no history available")
	}
	return c.history[len(c.history)-1], nil
}

// addToHistory adds a calculation to the history (private method)
func (c *Calculator) addToHistory(entry string) {
	c.history = append(c.history, entry)
	
	// Keep only last 100 entries to prevent unlimited growth
	if len(c.history) > 100 {
		c.history = c.history[1:]
	}
}

// Utility functions

// IsValidNumber checks if a number is valid (not NaN or infinite)
func IsValidNumber(n float64) bool {
	return !math.IsNaN(n) && !math.IsInf(n, 0)
}

// RoundToDecimals rounds a number to specified decimal places
func RoundToDecimals(n float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(n*multiplier) / multiplier
}

// PercentageOf calculates what percentage 'part' is of 'total'
func PercentageOf(part, total float64) (float64, error) {
	if total == 0 {
		return 0, errors.New("cannot calculate percentage of zero")
	}
	return (part / total) * 100, nil
}

// main function for demonstration
func main() {
	calc := NewCalculator()
	
	fmt.Println("Calculator Demo")
	fmt.Println("===============")
	
	// Basic operations
	sum := calc.Add(10, 5)
	fmt.Printf("10 + 5 = %.2f\n", sum)
	
	diff := calc.Subtract(10, 3)
	fmt.Printf("10 - 3 = %.2f\n", diff)
	
	product := calc.Multiply(4, 7)
	fmt.Printf("4 * 7 = %.2f\n", product)
	
	quotient, err := calc.Divide(15, 3)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("15 / 3 = %.2f\n", quotient)
	}
	
	// Complex operation
	result, err := calc.ComplexOperation(2, 3, 4, "sum")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Complex sum: %.2f\n", result)
	}
	
	// Memory operations
	calc.SetMemory(42)
	calc.AddToMemory(8)
	fmt.Printf("Memory: %.2f\n", calc.GetMemory())
	
	// Show history
	fmt.Println("\nCalculation History:")
	for i, entry := range calc.GetHistory() {
		fmt.Printf("%d. %s\n", i+1, entry)
	}
}