package sosi

import "testing"

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
