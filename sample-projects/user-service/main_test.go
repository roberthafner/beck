package main

import (
	"testing"
)

func TestAddNumbers(t *testing.T) {
	tests := []struct {
		name string
		a    int
		b    int
		want int
	}{
		{"positive numbers", 5, 3, 8},
		{"negative numbers", -2, -3, -5},
		{"zero", 0, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AddNumbers(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("AddNumbers(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsEvenNumber(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want bool
	}{
		{"even number", 4, true},
		{"odd number", 5, false},
		{"zero", 0, true},
		{"negative even", -2, true},
		{"negative odd", -3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEvenNumber(tt.n)
			if got != tt.want {
				t.Errorf("IsEvenNumber(%d) = %v, want %v", tt.n, got, tt.want)
			}
		})
	}
}

func TestReverseString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{"simple string", "hello", "olleh"},
		{"empty string", "", ""},
		{"single char", "a", "a"},
		{"palindrome", "racecar", "racecar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReverseString(tt.s)
			if got != tt.want {
				t.Errorf("ReverseString(%q) = %q, want %q", tt.s, got, tt.want)
			}
		})
	}
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid input", "hello", false},
		{"empty input", "", true},
		{"whitespace only", "   ", true},
		{"too long", string(make([]byte, 101)), true},
		{"normal length", "valid input", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateInput(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	t.Run("returns config", func(t *testing.T) {
		config := ParseConfig()
		if config == nil {
			t.Error("ParseConfig() returned nil")
		}
		if config.Host == "" {
			t.Error("ParseConfig() returned config with empty Host")
		}
		if config.Port == "" {
			t.Error("ParseConfig() returned config with empty Port")
		}
	})
}

func TestCalculateTotal(t *testing.T) {
	tests := []struct {
		name    string
		amount  float64
		taxRate float64
		want    float64
	}{
		{"positive values", 100.0, 0.1, 110.0},
		{"zero amount", 0.0, 0.1, 0.0},
		{"negative amount", -50.0, 0.1, 0.0},
		{"zero tax", 100.0, 0.0, 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTotal(tt.amount, tt.taxRate)
			if got != tt.want {
				t.Errorf("CalculateTotal(%f, %f) = %f, want %f", tt.amount, tt.taxRate, got, tt.want)
			}
		})
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		name   string
		amount float64
		want   string
	}{
		{"positive amount", 123.45, "$123.45"},
		{"zero", 0.0, "$0.00"},
		{"negative", -50.25, "$-50.25"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatCurrency(tt.amount)
			if got != tt.want {
				t.Errorf("FormatCurrency(%f) = %q, want %q", tt.amount, got, tt.want)
			}
		})
	}
}
