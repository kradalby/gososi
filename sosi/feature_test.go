package sosi

import (
	"strings"
	"testing"
)

func TestConvertGeometryTypeToSOSI(t *testing.T) {
	tests := []struct {
		name        string
		geojsonType string
		expected    string
		shouldError bool
	}{
		{
			name:        "Point to PUNKT",
			geojsonType: "Point",
			expected:    "PUNKT",
			shouldError: false,
		},
		{
			name:        "LineString to KURVE",
			geojsonType: "LineString",
			expected:    "KURVE",
			shouldError: false,
		},
		{
			name:        "MultiPoint to SVERM",
			geojsonType: "MultiPoint",
			expected:    "SVERM",
			shouldError: false,
		},
		{
			name:        "Polygon type",
			geojsonType: "Polygon",
			expected:    "FLATE",
			shouldError: false,
		},
		{
			name:        "Unsupported type",
			geojsonType: "MultiLineString",
			expected:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertGeometryTypeToSOSI(tt.geojsonType)

			if tt.shouldError && err == nil {
				t.Errorf("ConvertGeometryTypeToSOSI() should have returned an error for %s", tt.geojsonType)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("ConvertGeometryTypeToSOSI() returned unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("ConvertGeometryTypeToSOSI() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateFeature(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name     string
		feature  SOSIFeature
		expected string
	}{
		{
			name: "Point feature",
			feature: SOSIFeature{
				ID:         1,
				Type:       "PUNKT",
				ObjectType: "Fordelingsskap",
				Coordinates: []Coordinate{
					{Y: 63.4856654467292, X: 10.921292458550036, Z: 0.0},
				},
			},
			expected: `.PUNKT 1:
..OBJTYPE Fordelingsskap
..NØH
634856654 109212925 0000
`,
		},
		{
			name: "LineString feature",
			feature: SOSIFeature{
				ID:         3,
				Type:       "KURVE",
				ObjectType: "TeleFibertrase",
				Coordinates: []Coordinate{
					{Y: 63.48565826234929, X: 10.92128709413111, Z: 0.0},
					{Y: 63.48581871307087, X: 10.921147619239004, Z: 0.0},
				},
			},
			expected: `.KURVE 3:
..OBJTYPE TeleFibertrase
..NØH
634856583 109212871 0000
634858187 109211476 0000
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateFeature(tt.feature, config)
			if result != tt.expected {
				t.Errorf("GenerateFeature() = %v, want %v", result, tt.expected)

				// Show differences line by line for debugging
				resultLines := strings.Split(result, "\n")
				expectedLines := strings.Split(tt.expected, "\n")

				for i, line := range expectedLines {
					if i < len(resultLines) {
						if resultLines[i] != line {
							t.Errorf("Line %d: got %q, want %q", i+1, resultLines[i], line)
						}
					} else {
						t.Errorf("Missing line %d: want %q", i+1, line)
					}
				}
			}
		})
	}
}
