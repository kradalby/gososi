package sosi

import (
	"testing"
)

func TestFormatCoordinateToSOSI(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		precision int
		expected  string
	}{
		{
			name:      "latitude with 7 decimals",
			value:     63.4856654467292,
			precision: 7,
			expected:  "634856654",
		},
		{
			name:      "longitude with 7 decimals",
			value:     10.921292458550036,
			precision: 7,
			expected:  "109212925",
		},
		{
			name:      "altitude with 3 decimals",
			value:     0.0,
			precision: 3,
			expected:  "0000",
		},
		{
			name:      "altitude with value",
			value:     123.456,
			precision: 3,
			expected:  "123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCoordinateToSOSI(tt.value, tt.precision)
			if result != tt.expected {
				t.Errorf("FormatCoordinateToSOSI() = %v, want %v", result, tt.expected)
			}
		})
	}
}
