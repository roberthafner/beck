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

	tests := []struct {
		name string
		a    float64
		b    float64
		want float64
	}{
		{"positive numbers", 2.5, 3.7, 6.2},
		{"negative numbers", -2.5, -1.5, -4.0},
		{"mixed signs", -5.0, 3.0, -2.0},
		{"zero", 0, 5.5, 5.5},
		{"large numbers", 1000000.0, 2000000.0, 3000000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Add(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculator_Subtract(t *testing.T) {
	calc := NewCalculator("Test")

	tests := []struct {
		name string
		a    float64
		b    float64
		want float64
	}{
		{"positive numbers", 10.0, 4.0, 6.0},
		{"negative numbers", -5.0, -3.0, -2.0},
		{"mixed signs", 5.0, -3.0, 8.0},
		{"zero result", 7.0, 7.0, 0.0},
		{"subtract from zero", 0.0, 5.0, -5.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Subtract(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculator_Multiply(t *testing.T) {
	calc := NewCalculator("Test")

	tests := []struct {
		name string
		a    float64
		b    float64
		want float64
	}{
		{"positive numbers", 3.0, 4.0, 12.0},
		{"negative numbers", -2.0, -3.0, 6.0},
		{"mixed signs", -2.0, 3.0, -6.0},
		{"multiply by zero", 5.0, 0.0, 0.0},
		{"multiply by one", 7.0, 1.0, 7.0},
		{"decimal numbers", 2.5, 4.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Multiply(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculator_Divide(t *testing.T) {
	calc := NewCalculator("Test")

	tests := []struct {
		name    string
		a       float64
		b       float64
		want    float64
		wantErr bool
	}{
		{"normal division", 15.0, 3.0, 5.0, false},
		{"decimal result", 7.0, 2.0, 3.5, false},
		{"negative numbers", -10.0, -2.0, 5.0, false},
		{"mixed signs", -8.0, 2.0, -4.0, false},
		{"division by zero", 10.0, 0.0, 0.0, true},
		{"zero dividend", 0.0, 5.0, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Divide(tt.a, tt.b)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, "division by zero", err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCalculator_Power(t *testing.T) {
	calc := NewCalculator("Test")

	tests := []struct {
		name string
		a    float64
		b    float64
		want float64
	}{
		{"positive base and exponent", 2.0, 3.0, 8.0},
		{"power of zero", 5.0, 0.0, 1.0},
		{"power of one", 7.0, 1.0, 7.0},
		{"square", 4.0, 2.0, 16.0},
		{"fractional exponent", 9.0, 0.5, 3.0},
		{"negative exponent", 2.0, -2.0, 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.Power(tt.a, tt.b)
			assert.InDelta(t, tt.want, got, 0.0001)
		})
	}
}

func TestCalculator_Sqrt(t *testing.T) {
	calc := NewCalculator("Test")

	tests := []struct {
		name    string
		a       float64
		want    float64
		wantErr bool
	}{
		{"perfect square", 16.0, 4.0, false},
		{"non-perfect square", 2.0, 1.4142135623730951, false},
		{"zero", 0.0, 0.0, false},
		{"decimal", 6.25, 2.5, false},
		{"negative number", -4.0, 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Sqrt(tt.a)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot calculate square root of negative number")
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.want, got, 0.0001)
			}
		})
	}
}

func TestCalculator_GetName(t *testing.T) {
	tests := []struct {
		name           string
		calculatorName string
	}{
		{"simple name", "MyCalc"},
		{"empty name", ""},
		{"name with spaces", "Scientific Calculator"},
		{"name with numbers", "Calc 3000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calc := NewCalculator(tt.calculatorName)
			got := calc.GetName()
			assert.Equal(t, tt.calculatorName, got)
		})
	}
}

func TestIsEven(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want bool
	}{
		{"even positive", 4, true},
		{"odd positive", 5, false},
		{"zero", 0, true},
		{"even negative", -2, true},
		{"odd negative", -3, false},
		{"large even", 1000, true},
		{"large odd", 999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEven(tt.n)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFactorial(t *testing.T) {
	tests := []struct {
		name    string
		n       int
		want    int
		wantErr bool
	}{
		{"factorial of 0", 0, 1, false},
		{"factorial of 1", 1, 1, false},
		{"factorial of 5", 5, 120, false},
		{"factorial of 6", 6, 720, false},
		{"factorial of 10", 10, 3628800, false},
		{"negative number", -1, 0, true},
		{"large negative", -5, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Factorial(tt.n)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "factorial is not defined for negative numbers")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFibonacci(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"fibonacci 0", 0, 0},
		{"fibonacci 1", 1, 1},
		{"fibonacci 2", 2, 1},
		{"fibonacci 3", 3, 2},
		{"fibonacci 5", 5, 5},
		{"fibonacci 7", 7, 13},
		{"fibonacci 10", 10, 55},
		{"negative input", -5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Fibonacci(tt.n)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsPrime(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want bool
	}{
		{"not prime - negative", -5, false},
		{"not prime - zero", 0, false},
		{"not prime - one", 1, false},
		{"prime - two", 2, true},
		{"prime - three", 3, true},
		{"not prime - four", 4, false},
		{"prime - five", 5, true},
		{"not prime - six", 6, false},
		{"prime - seven", 7, true},
		{"not prime - eight", 8, false},
		{"not prime - nine", 9, false},
		{"prime - eleven", 11, true},
		{"not prime - fifteen", 15, false},
		{"prime - seventeen", 17, true},
		{"prime - large", 97, true},
		{"not prime - large", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPrime(tt.n)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		want float64
	}{
		{"first larger", 5.0, 3.0, 5.0},
		{"second larger", 2.0, 8.0, 8.0},
		{"equal values", 4.5, 4.5, 4.5},
		{"negative numbers", -3.0, -7.0, -3.0},
		{"mixed signs", -2.0, 3.0, 3.0},
		{"zero and positive", 0.0, 1.0, 1.0},
		{"zero and negative", 0.0, -1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Max(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		b    float64
		want float64
	}{
		{"first smaller", 3.0, 5.0, 3.0},
		{"second smaller", 8.0, 2.0, 2.0},
		{"equal values", 4.5, 4.5, 4.5},
		{"negative numbers", -3.0, -7.0, -7.0},
		{"mixed signs", -2.0, 3.0, -2.0},
		{"zero and positive", 0.0, 1.0, 0.0},
		{"zero and negative", 0.0, -1.0, -1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Min(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		name string
		a    float64
		want float64
	}{
		{"positive number", 5.0, 5.0},
		{"negative number", -5.0, 5.0},
		{"zero", 0.0, 0.0},
		{"decimal positive", 3.14, 3.14},
		{"decimal negative", -3.14, 3.14},
		{"large positive", 1000000.0, 1000000.0},
		{"large negative", -1000000.0, 1000000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Abs(tt.a)
			assert.Equal(t, tt.want, got)
		})
	}
}
