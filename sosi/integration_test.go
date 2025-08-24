package sosi

import (
	"strings"
	"testing"
)

// TestIntegrationExampleData tests that our SOSI output matches the expected format
// from the JavaScript implementation for the example.json test case
func TestIntegrationExampleData(t *testing.T) {
	config := DefaultConfig()
	builder := NewBuilder(config)

	// Features from example.json test data
	features := []SOSIFeature{
		{
			ID:         1,
			Type:       "PUNKT",
			ObjectType: "Fordelingsskap",
			Coordinates: []Coordinate{
				{Y: 63.4856654467292, X: 10.921292458550036, Z: 0.0},
			},
		},
		{
			ID:         2,
			Type:       "PUNKT",
			ObjectType: "Fordelingsskap",
			Coordinates: []Coordinate{
				{Y: 63.486702373293724, X: 10.920455609197402, Z: 0.0},
			},
		},
		{
			ID:         3,
			Type:       "KURVE",
			ObjectType: "TeleFibertrase",
			Coordinates: []Coordinate{
				{Y: 63.48565826234929, X: 10.92128709413111, Z: 0.0},
				{Y: 63.48581871307087, X: 10.921147619239004, Z: 0.0},
				{Y: 63.486024663423365, X: 10.920975957833337, Z: 0.0},
				{Y: 63.486216243814425, X: 10.920825754103376, Z: 0.0},
				{Y: 63.48636232299978, X: 10.920713101305905, Z: 0.0},
				{Y: 63.486539532812685, X: 10.920584355251654, Z: 0.0},
				{Y: 63.48665447913092, X: 10.920493160129892, Z: 0.0},
				{Y: 63.486702373293724, X: 10.920444880359549, Z: 0.0},
			},
		},
		{
			ID:         4,
			Type:       "KURVE",
			ObjectType: "TeleFibertrase",
			Coordinates: []Coordinate{
				{Y: 63.486702373293724, X: 10.92046097361633, Z: 0.0},
				{Y: 63.48669758388104, X: 10.92017665941319, Z: 0.0},
				{Y: 63.486659268550795, X: 10.919726048223312, Z: 0.0},
				{Y: 63.48664250557766, X: 10.919543657979787, Z: 0.0},
				{Y: 63.486644900288695, X: 10.919457827276956, Z: 0.0},
			},
		},
		{
			ID:         5,
			Type:       "PUNKT",
			ObjectType: "Kum",
			Coordinates: []Coordinate{
				{Y: 63.485960004751576, X: 10.919962082656106, Z: 0.0},
			},
		},
	}

	for _, feature := range features {
		builder.AddFeature(feature)
	}

	result, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() returned error: %v", err)
	}

	// Expected key components that must be present and correctly formatted
	expectedComponents := []string{
		".HODE 0:",
		"..TEGNSETT UTF-8",
		"..TRANSPAR",
		"...KOORDSYS 84",
		"...ORIGO-NØ 0 0",
		"...ENHET 0.0000001",
		"...ENHET-H 0.001",
		"...ENHET-D 0.001",
		"..PRODUSENT \"GeoJSONtoSOSI\"",
		"..SOSI-VERSJON 4.0",
		"..SOSI-NIVÅ 2",
		"..OMRÅDE",
		"...MIN-NØ 63 10",
		"...MAX-NØ 64 11",
		".PUNKT 1:",
		"..OBJTYPE Fordelingsskap",
		"..NØH",
		"634856654 109212925 0000",
		".PUNKT 2:",
		"634867024 109204556 0000",
		".KURVE 3:",
		"..OBJTYPE TeleFibertrase",
		"634856583 109212871 0000",
		"634858187 109211476 0000",
		"634860247 109209760 0000",
		"634862162 109208258 0000",
		"634863623 109207131 0000",
		"634865395 109205844 0000",
		"634866545 109204932 0000",
		"634867024 109204449 0000",
		".KURVE 4:",
		"634867024 109204610 0000",
		"634866976 109201767 0000",
		"634866593 109197260 0000",
		"634866425 109195437 0000",
		"634866449 109194578 0000",
		".PUNKT 5:",
		"..OBJTYPE Kum",
		"634859600 109199621 0000",
		".SLUTT",
	}

	// Check that all expected components are present
	for _, component := range expectedComponents {
		if !strings.Contains(result, component) {
			t.Errorf("Missing expected component: %s", component)
			t.Logf("Full result:\n%s", result)
		}
	}

	// Verify the structure by splitting into lines
	lines := strings.Split(strings.TrimSpace(result), "\n")

	// Should start with header
	if !strings.HasPrefix(lines[0], ".HODE") {
		t.Errorf("First line should start with .HODE, got: %s", lines[0])
	}

	// Should end with .SLUTT
	if lines[len(lines)-1] != ".SLUTT" {
		t.Errorf("Last line should be .SLUTT, got: %s", lines[len(lines)-1])
	}
}

// TestIntegrationCoordinateTransformation verifies specific coordinate transformations
func TestIntegrationCoordinateTransformation(t *testing.T) {
	config := DefaultConfig()

	// Test cases from the example data with expected SOSI output
	testCases := []struct {
		name     string
		input    Coordinate
		expected string
	}{
		{
			name:     "Point 1 from example",
			input:    ConvertCoordinate(63.4856654467292, 10.921292458550036, 0.0),
			expected: "634856654 109212925 0000",
		},
		{
			name:     "Point 2 from example",
			input:    ConvertCoordinate(63.486702373293724, 10.920455609197402, 0.0),
			expected: "634867024 109204556 0000",
		},
		{
			name:     "LineString coordinate from example",
			input:    ConvertCoordinate(63.48565826234929, 10.92128709413111, 0.0),
			expected: "634856583 109212871 0000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := TransformCoordinate(tc.input, config)
			if result != tc.expected {
				t.Errorf("TransformCoordinate() = %s, want %s", result, tc.expected)
			}
		})
	}
}
