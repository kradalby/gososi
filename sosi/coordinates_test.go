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

func TestTransformCoordinate(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name     string
		coord    Coordinate
		expected string
	}{
		{
			name: "basic coordinate transformation",
			coord: Coordinate{
				Y: 63.4856654467292,   // Y = Latitude
				X: 10.921292458550036, // X = Longitude
				Z: 0.0,                // Z = Altitude
			},
			expected: "634856654 109212925 0000",
		},
		{
			name: "coordinate with altitude",
			coord: Coordinate{
				Y: 63.486702373293724, // Y = Latitude
				X: 10.920455609197402, // X = Longitude
				Z: 123.456,            // Z = Altitude
			},
			expected: "634867024 109204556 123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformCoordinate(tt.coord, config)
			if result != tt.expected {
				t.Errorf("TransformCoordinate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTransformCoordinates(t *testing.T) {
	config := DefaultConfig()

	coords := []Coordinate{
		{Y: 63.4856654467292, X: 10.921292458550036, Z: 0.0},
		{Y: 63.486702373293724, X: 10.920455609197402, Z: 0.0},
	}

	expected := []string{
		"634856654 109212925 0000",
		"634867024 109204556 0000",
	}

	result := TransformCoordinates(coords, config)

	if len(result) != len(expected) {
		t.Fatalf("TransformCoordinates() returned %d items, want %d", len(result), len(expected))
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("TransformCoordinates()[%d] = %v, want %v", i, result[i], exp)
		}
	}
}
