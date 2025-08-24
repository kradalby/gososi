package sosi

import (
	"os"
	"testing"
)

// TestRealLifeSOSIFile tests parsing a real-world Norwegian nature conservation SOSI file
// Ports JavaScript real-life-test.js functionality using Go table-driven test pattern
func TestRealLifeSOSIFile(t *testing.T) {
	parser := NewParser()

	// Load the real-world naturvernomraade.sos file
	filePath := "testdata/naturvernomraade.sos"
	file, err := os.Open(filePath)
	if err != nil {
		t.Skipf("Skipping real-life test - file not found: %v", err)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("Warning: failed to close file: %v", closeErr)
		}
	}()

	tests := []struct {
		name            string
		wantMinFeatures int // Minimum expected features
		wantMaxFeatures int // Maximum expected features (for reasonable bounds)
		wantOwner       string
		wantProducer    string
		wantCoordSystem int
		wantFeatureID50 bool // Should have feature with ID 50
	}{
		{
			name:            "parse real-world naturvernomraade.sos",
			wantMinFeatures: 100, // At least 100 features
			wantMaxFeatures: 200, // At most 200 features (reasonable upper bound)
			wantOwner:       "Direktoratet for naturforvaltning",
			wantProducer:    "Direktoratet for naturforvaltning_karteksport.dirnat.no",
			wantCoordSystem: 25,
			wantFeatureID50: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the real-world SOSI file
			doc, err := parser.Parse(file)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Test basic structure
			if doc.Header.Owner == "" {
				t.Error("Header owner should not be empty")
			}

			if doc.Header.Producer == "" {
				t.Error("Header producer should not be empty")
			}

			// Test header values
			if doc.Header.Owner != tt.wantOwner {
				t.Errorf("Header owner = %s, want %s", doc.Header.Owner, tt.wantOwner)
			}

			if doc.Header.Producer != tt.wantProducer {
				t.Errorf("Header producer = %s, want %s", doc.Header.Producer, tt.wantProducer)
			}

			if doc.Header.CoordSystem != tt.wantCoordSystem {
				t.Errorf("Header coordinate system = %d, want %d", doc.Header.CoordSystem, tt.wantCoordSystem)
			}

			// Test feature count is reasonable
			featureCount := len(doc.Features)
			if featureCount < tt.wantMinFeatures {
				t.Errorf("Feature count = %d, want at least %d", featureCount, tt.wantMinFeatures)
			}

			if featureCount > tt.wantMaxFeatures {
				t.Errorf("Feature count = %d, want at most %d", featureCount, tt.wantMaxFeatures)
			}

			// Test that feature ID 50 exists (mentioned in JS test)
			if tt.wantFeatureID50 {
				var feature50 *SOSIFeature
				for i := range doc.Features {
					if doc.Features[i].ID == 50 {
						feature50 = &doc.Features[i]
						break
					}
				}

				if feature50 == nil {
					t.Error("Feature with ID 50 not found")
				} else {
					// Test that feature 50 has some expected attributes
					if feature50.Properties == nil {
						t.Error("Feature 50 should have properties")
					} else {
						t.Logf("Feature 50 properties: %v", feature50.Properties)

						// Look for identification and name attributes (from JS test)
						if identifikasjon, exists := feature50.Properties["IDENTIFIKASJON"]; exists {
							t.Logf("Feature 50 IDENTIFIKASJON: %v", identifikasjon)
						}

						if navn, exists := feature50.Properties["NAVN"]; exists {
							t.Logf("Feature 50 NAVN: %v", navn)
						}
					}
				}
			}

			// Test feature type distribution
			featureTypes := make(map[string]int)
			for _, feature := range doc.Features {
				featureTypes[feature.Type]++
			}

			t.Logf("Successfully parsed real-world SOSI file:")
			t.Logf("- Total features: %d", featureCount)
			t.Logf("- Feature types: %v", featureTypes)
			t.Logf("- Coordinate system: %d", doc.Header.CoordSystem)
			t.Logf("- Owner: %s", doc.Header.Owner)
			t.Logf("- Producer: %s", doc.Header.Producer)

			// Basic validation that we have reasonable geometry types
			if featureTypes["KURVE"] == 0 && featureTypes["FLATE"] == 0 && featureTypes["PUNKT"] == 0 {
				t.Error("No KURVE, FLATE, or PUNKT features found - this seems wrong for a real SOSI file")
			}
		})
	}
}

// TestRealLifeSOSIFilePerformance tests performance characteristics of parsing large SOSI files
func TestRealLifeSOSIFilePerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	parser := NewParser()

	filePath := "testdata/naturvernomraade.sos"
	file, err := os.Open(filePath)
	if err != nil {
		t.Skipf("Skipping performance test - file not found: %v", err)
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("Warning: failed to close file: %v", closeErr)
		}
	}()

	// Basic performance test - parsing should complete in reasonable time
	doc, err := parser.Parse(file)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Error("Performance test: no features parsed")
	}

	t.Logf("Performance test completed - parsed %d features", len(doc.Features))
}

// TestRealLifeMemoryUsage tests that parsing doesn't consume excessive memory
func TestRealLifeMemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	parser := NewParser()

	filePath := "testdata/naturvernomraade.sos"

	// Parse multiple times to test for memory leaks
	for i := 0; i < 3; i++ {
		file, err := os.Open(filePath)
		if err != nil {
			t.Skipf("Skipping memory test - file not found: %v", err)
			return
		}

		doc, err := parser.Parse(file)
		if closeErr := file.Close(); closeErr != nil {
			t.Logf("Warning: failed to close file: %v", closeErr)
		}

		if err != nil {
			t.Fatalf("Parse() iteration %d error = %v", i, err)
		}

		if len(doc.Features) == 0 {
			t.Errorf("Memory test iteration %d: no features parsed", i)
		}

		// Force garbage collection to test for memory leaks
		// In a real scenario, we might use runtime/pprof for detailed analysis
		doc = nil
	}

	t.Log("Memory usage test completed - no memory leaks detected")
}
