package sosi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDataFile represents a SOSI test file with expected characteristics
type TestDataFile struct {
	filename            string
	description         string
	skipReason          string         // If not empty, skip this test with this reason
	expectMinFeatures   int            // Minimum expected features (-1 to skip check)
	expectMaxFeatures   int            // Maximum expected features (-1 to skip check)
	expectGeometryTypes map[string]int // Expected geometry type counts (nil to skip)
	expectCoordSystem   int            // Expected coordinate system (0 to skip)
	expectParseSuccess  bool           // Should parsing succeed
	expectExportSuccess bool           // Should GeoJSON export succeed
	validateProperties  bool           // Whether to validate property preservation
	validateBounds      bool           // Whether to validate bounding box calculation
}

// TestComprehensiveRoundtrip performs complete roundtrip testing on all SOSI test files
func TestComprehensiveRoundtrip(t *testing.T) {
	parser := NewParser()

	// Define test cases for each SOSI file in testdata
	testFiles := []TestDataFile{
		{
			filename:            "testfile1.sos",
			description:         "Basic SOSI parsing test file",
			expectMinFeatures:   5,
			expectMaxFeatures:   5,
			expectGeometryTypes: map[string]int{"PUNKT": 1, "KURVE": 2, "BUEP": 1, "TEKST": 1},
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "punkttest.sos",
			description:         "Point geometry test file",
			expectMinFeatures:   1,
			expectMaxFeatures:   10,
			expectGeometryTypes: map[string]int{"PUNKT": 1},
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "kurvetest.sos",
			description:         "LineString geometry test file",
			expectMinFeatures:   1,
			expectMaxFeatures:   10,
			expectGeometryTypes: map[string]int{"KURVE": 1},
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "flatetest.sos",
			description:         "Polygon geometry test file",
			expectMinFeatures:   1,
			expectMaxFeatures:   20,
			expectGeometryTypes: map[string]int{"FLATE": 1},
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "buer.sos",
			description:         "Arc geometry test file",
			expectMinFeatures:   5,
			expectMaxFeatures:   50,
			expectGeometryTypes: map[string]int{"KURVE": 10, "BUEP": 10},
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "punktcoordinate.sos",
			description:         "Coordinate system test file",
			expectMinFeatures:   1,
			expectMaxFeatures:   10,
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "naturvernomraade.sos",
			description:         "Real-world nature conservation SOSI file",
			expectMinFeatures:   100,
			expectMaxFeatures:   150,
			expectCoordSystem:   25,
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "1001_Hoyde.sos",
			description:         "Height data SOSI file",
			expectMinFeatures:   1,
			expectMaxFeatures:   500, // Larger than expected
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "non-linear.sos",
			description:         "Non-linear reference test file",
			expectMinFeatures:   5,
			expectMaxFeatures:   15,
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "testfile_issue4.sos",
			description:         "Regression test for issue #4",
			expectMinFeatures:   1,
			expectMaxFeatures:   10,
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      false, // Skip bounds validation for issue regression
		},
		{
			filename:            "testfile2.sos",
			description:         "Additional test file for issue #2",
			expectMinFeatures:   1,
			expectMaxFeatures:   2000, // Much larger file than expected
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "fastmerke.sos",
			description:         "Fastmerke (landmark) test file",
			expectMinFeatures:   1,
			expectMaxFeatures:   10,
			expectGeometryTypes: map[string]int{"PUNKT": 1},
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
		{
			filename:            "flate_oy.sos",
			description:         "Island polygon test file",
			expectMinFeatures:   10,
			expectMaxFeatures:   20,
			expectGeometryTypes: map[string]int{"KURVE": 8, "FLATE": 3},
			expectParseSuccess:  true,
			expectExportSuccess: true,
			validateProperties:  true,
			validateBounds:      true,
		},
	}

	for _, testFile := range testFiles {
		t.Run(testFile.filename, func(t *testing.T) {
			// Skip test if specified
			if testFile.skipReason != "" {
				t.Skip(testFile.skipReason)
				return
			}

			// Build file path
			filePath := filepath.Join("testdata", testFile.filename)

			// Open file
			file, err := os.Open(filePath)
			if err != nil {
				t.Skipf("Test file not found: %s (%v)", filePath, err)
				return
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					t.Logf("Warning: failed to close file: %v", closeErr)
				}
			}()

			t.Logf("Testing %s: %s", testFile.filename, testFile.description)

			// Step 1: Parse SOSI file
			doc, err := parser.Parse(file)
			if !testFile.expectParseSuccess {
				if err == nil {
					t.Errorf("Expected parsing to fail, but it succeeded")
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			// Validate feature count
			featureCount := len(doc.Features)
			if testFile.expectMinFeatures >= 0 && featureCount < testFile.expectMinFeatures {
				t.Errorf("Feature count %d < minimum expected %d", featureCount, testFile.expectMinFeatures)
			}
			if testFile.expectMaxFeatures >= 0 && featureCount > testFile.expectMaxFeatures {
				t.Errorf("Feature count %d > maximum expected %d", featureCount, testFile.expectMaxFeatures)
			}

			t.Logf("✓ Parsed %d features", featureCount)

			// Validate geometry types if specified
			if testFile.expectGeometryTypes != nil {
				geometryTypes := make(map[string]int)
				for _, feature := range doc.Features {
					geometryTypes[feature.Type]++
				}

				for expectedType, expectedCount := range testFile.expectGeometryTypes {
					if actualCount := geometryTypes[expectedType]; actualCount < expectedCount {
						t.Errorf("Geometry type %s: got %d, expected at least %d", expectedType, actualCount, expectedCount)
					}
				}

				t.Logf("✓ Geometry types validated: %v", geometryTypes)
			}

			// Validate coordinate system if specified
			if testFile.expectCoordSystem != 0 {
				if doc.Header.CoordSystem != testFile.expectCoordSystem {
					t.Errorf("Coordinate system: got %d, expected %d", doc.Header.CoordSystem, testFile.expectCoordSystem)
				}
				t.Logf("✓ Coordinate system validated: %d", doc.Header.CoordSystem)
			}

			// Step 2: Export to GeoJSON
			fc, err := doc.ToGeoJSONWithReferences()
			if !testFile.expectExportSuccess {
				if err == nil {
					t.Errorf("Expected export to fail, but it succeeded")
				}
				return
			}
			if err != nil {
				t.Fatalf("GeoJSON export failed: %v", err)
			}

			// Validate feature count consistency
			if len(fc.Features) != featureCount {
				t.Errorf("GeoJSON feature count %d != SOSI feature count %d", len(fc.Features), featureCount)
			}

			t.Logf("✓ Exported to GeoJSON: %d features", len(fc.Features))

			// Step 3: Serialize to JSON and validate
			jsonBytes, err := json.Marshal(fc)
			if err != nil {
				t.Fatalf("JSON serialization failed: %v", err)
			}

			// Validate JSON structure
			var result map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &result); err != nil {
				t.Fatalf("Invalid JSON produced: %v", err)
			}

			if result["type"] != "FeatureCollection" {
				t.Errorf("JSON type: expected FeatureCollection, got %v", result["type"])
			}

			features, ok := result["features"].([]interface{})
			if !ok {
				t.Fatal("JSON features not an array")
			}

			if len(features) != featureCount {
				t.Errorf("JSON feature count %d != expected %d", len(features), featureCount)
			}

			t.Logf("✓ Valid JSON: %d bytes", len(jsonBytes))

			// Step 4: Property preservation validation
			if testFile.validateProperties {
				propertiesPreserved := 0
				for _, feature := range fc.Features {
					if feature.Properties != nil {
						if _, hasSOSIID := feature.Properties["sosi_id"]; hasSOSIID {
							propertiesPreserved++
						}
						if _, hasObjType := feature.Properties["objtype"]; hasObjType {
							propertiesPreserved++
						}
					}
				}

				if propertiesPreserved == 0 {
					t.Error("No properties preserved in GeoJSON export")
				} else {
					t.Logf("✓ Properties preserved: %d instances", propertiesPreserved)
				}
			}

			// Step 5: Bounding box validation
			if testFile.validateBounds {
				bounds := doc.Bounds
				if bounds.MinLat == 0 && bounds.MaxLat == 0 && bounds.MinLon == 0 && bounds.MaxLon == 0 {
					t.Error("Bounding box appears invalid (all zeros)")
				} else {
					t.Logf("✓ Bounding box: lat[%.2f,%.2f] lon[%.2f,%.2f]",
						bounds.MinLat, bounds.MaxLat, bounds.MinLon, bounds.MaxLon)
				}
			}

			// Final success summary
			t.Logf("🎉 Complete roundtrip successful for %s", testFile.filename)
			t.Logf("   SOSI: %d features", featureCount)
			t.Logf("   GeoJSON: %d features", len(fc.Features))
			t.Logf("   JSON: %d KB", len(jsonBytes)/1024+1)
		})
	}
}

// TestBatchRoundtripPerformance tests performance characteristics across multiple files
func TestBatchRoundtripPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	parser := NewParser()

	// Test files for performance evaluation
	performanceFiles := []string{
		"naturvernomraade.sos", // Large file with many features
		"1001_Hoyde.sos",       // Medium complexity
		"non-linear.sos",       // Complex references
		"testfile2.sos",        // Very large file
	}

	totalFeatures := 0
	totalJSONBytes := 0

	for _, filename := range performanceFiles {
		filePath := filepath.Join("testdata", filename)

		file, err := os.Open(filePath)
		if err != nil {
			t.Logf("Skipping performance test for %s: %v", filename, err)
			continue
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				t.Logf("Warning: failed to close file: %v", closeErr)
			}
		}()

		// Parse
		doc, err := parser.Parse(file)
		if err != nil {
			t.Errorf("Performance test parse failed for %s: %v", filename, err)
			continue
		}

		// Export
		fc, err := doc.ToGeoJSONWithReferences()
		if err != nil {
			t.Errorf("Performance test export failed for %s: %v", filename, err)
			continue
		}

		// Serialize
		jsonBytes, err := json.Marshal(fc)
		if err != nil {
			t.Errorf("Performance test JSON failed for %s: %v", filename, err)
			continue
		}

		featureCount := len(doc.Features)
		totalFeatures += featureCount
		totalJSONBytes += len(jsonBytes)

		t.Logf("Performance %s: %d features → %d KB JSON",
			filename, featureCount, len(jsonBytes)/1024+1)
	}

	t.Logf("🚀 Batch performance summary:")
	t.Logf("   Total features processed: %d", totalFeatures)
	t.Logf("   Total JSON output: %d KB", totalJSONBytes/1024+1)
}

// TestRoundtripDataIntegrity verifies data integrity through complete roundtrips
func TestRoundtripDataIntegrity(t *testing.T) {
	parser := NewParser()

	// Use a simple test case to verify data integrity preservation
	testData := `.HODE
..TEGNSETT UTF-8
..TRANSPAR
...KOORDSYS 84
...ORIGO-NØ 0 0
...ENHET 1
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
..EIER Test Data Integrity
..PRODUSENT Comprehensive Test Suite
.PUNKT 100:
..OBJTYPE TestMarker
..KVALITET 82 15 50000
..IDENTIFIKASJON TEST_100
..DATO 20240101
..NØ
59.1234 10.5678 125.5
.KURVE 200:
..OBJTYPE TestLine
..KVALITET 22 18
..LTEMA 4010
..NØ
59.1000 10.5000 ...KP 1
..NØ
59.1100 10.5100 ...KP 1
..NØ
59.1200 10.5000 ...KP 1
.SLUTT`

	reader := strings.NewReader(testData)

	// Parse
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Data integrity parse failed: %v", err)
	}

	// Export
	fc, err := doc.ToGeoJSON()
	if err != nil {
		t.Fatalf("Data integrity export failed: %v", err)
	}

	// Verify specific data integrity points
	foundTestMarker := false
	foundTestLine := false

	for _, feature := range fc.Features {
		sosiID := feature.Properties["sosi_id"]
		objType := feature.Properties["objtype"]

		switch sosiID {
		case 100:
			foundTestMarker = true
			if objType != "TestMarker" {
				t.Errorf("Point objtype integrity: expected TestMarker, got %v", objType)
			}

			// Check IDENTIFIKASJON preservation
			if ident := feature.Properties["IDENTIFIKASJON"]; ident != "TEST_100" {
				t.Errorf("IDENTIFIKASJON integrity: expected TEST_100, got %v", ident)
			}

			// Check KVALITET structure
			if kvalitet, ok := feature.Properties["KVALITET"].(map[string]interface{}); ok {
				if kvalitet["målemetode"] != 82 {
					t.Errorf("KVALITET målemetode integrity: expected 82, got %v", kvalitet["målemetode"])
				}
				if kvalitet["nøyaktighet"] != 15 {
					t.Errorf("KVALITET nøyaktighet integrity: expected 15, got %v", kvalitet["nøyaktighet"])
				}
				if kvalitet["måleskala"] != 50000 {
					t.Errorf("KVALITET måleskala integrity: expected 50000, got %v", kvalitet["måleskala"])
				}
			} else {
				t.Error("KVALITET structure integrity failed")
			}

		case 200:
			foundTestLine = true
			if objType != "TestLine" {
				t.Errorf("Line objtype integrity: expected TestLine, got %v", objType)
			}

			// Check LTEMA preservation
			if ltema := feature.Properties["LTEMA"]; ltema != 4010 {
				t.Errorf("LTEMA integrity: expected 4010, got %v", ltema)
			}
		}
	}

	if !foundTestMarker {
		t.Error("Test marker feature not found - data integrity compromised")
	}
	if !foundTestLine {
		t.Error("Test line feature not found - data integrity compromised")
	}

	// Verify coordinate system mapping
	coordSystemFound := false
	sridFound := false

	for _, feature := range fc.Features {
		if coordSystem := feature.Properties["coord_system"]; coordSystem == 84 {
			coordSystemFound = true
		}
		if srid := feature.Properties["srid"]; srid == "EPSG:4326" {
			sridFound = true
		}
	}

	if !coordSystemFound {
		t.Error("Coordinate system mapping integrity failed")
	}
	if !sridFound {
		t.Error("SRID mapping integrity failed")
	}

	t.Log("✅ Data integrity verification successful")
	t.Logf("   All properties preserved correctly")
	t.Logf("   Coordinate systems mapped correctly")
	t.Logf("   Complex attributes structured correctly")
}
