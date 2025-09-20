package main

import (
	"errors"
	"math"
)

// Calculator provides basic mathematical operations
type Calculator struct {
	name string
}

// NewCalculator creates a new calculator instance
func NewCalculator(name string) *Calculator {
	return &Calculator{name: name}
}

// Add performs addition of two numbers
func (c *Calculator) Add(a, b float64) float64 {
	return a + b
}

// Subtract performs subtraction of two numbers
func (c *Calculator) Subtract(a, b float64) float64 {
	return a - b
}

// Multiply performs multiplication of two numbers
func (c *Calculator) Multiply(a, b float64) float64 {
	return a * b
}

// Divide performs division of two numbers
func (c *Calculator) Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

// Power calculates a to the power of b
func (c *Calculator) Power(a, b float64) float64 {
	return math.Pow(a, b)
}

// Sqrt calculates the square root of a number
func (c *Calculator) Sqrt(a float64) (float64, error) {
	if a < 0 {
		return 0, errors.New("cannot calculate square root of negative number")
	}
	return math.Sqrt(a), nil
}

// GetName returns the calculator's name
func (c *Calculator) GetName() string {
	return c.name
}

// IsEven checks if a number is even
func IsEven(n int) bool {
	return n%2 == 0
}

// Factorial calculates the factorial of a number
func Factorial(n int) (int, error) {
	if n < 0 {
		return 0, errors.New("factorial is not defined for negative numbers")
	}
	if n == 0 || n == 1 {
		return 1, nil
	}

	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result, nil
}

// Fibonacci returns the nth Fibonacci number
func Fibonacci(n int) int {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}

	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// IsPrime checks if a number is prime
func IsPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}

	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// Max returns the maximum of two numbers
func Max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// Min returns the minimum of two numbers
func Min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Abs returns the absolute value of a number
func Abs(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}

func main() {
	calc := NewCalculator("My Calculator")
	println("Calculator:", calc.GetName())
	println("2 + 3 =", calc.Add(2, 3))
	println("10 - 4 =", calc.Subtract(10, 4))
	println("5 * 6 =", calc.Multiply(5, 6))

	result, err := calc.Divide(15, 3)
	if err != nil {
		println("Error:", err.Error())
	} else {
		println("15 / 3 =", result)
	}

	println("Is 4 even?", IsEven(4))
	println("Is 7 prime?", IsPrime(7))

	fact, _ := Factorial(5)
	println("5! =", fact)

	println("Fibonacci(10) =", Fibonacci(10))
}
