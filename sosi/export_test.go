package sosi

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kradalby/gososi/geojson"
)

// TestToGeoJSON tests the basic GeoJSON export functionality
func TestToGeoJSON(t *testing.T) {
	parser := NewParser()

	// Test SOSI data with different geometry types
	sosiData := `.HODE
..TEGNSETT UTF-8
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0 0
...ENHET 1
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
..EIER Testdata
..PRODUSENT TestExport
.PUNKT 100:
..OBJTYPE Fastmerke
..KVALITET 22
..NØ
7000000 300000
.KURVE 200:
..OBJTYPE Bygning
..KVALITET 22
..NØ
7000010 300010 ...KP 1
..NØ
7000020 300010 ...KP 1
..NØ
7000020 300020 ...KP 1
.FLATE 300:
..OBJTYPE Område
..KVALITET 22
..NØ
7000000 300000
7000010 300000
7000010 300010
7000000 300010
7000000 300000
.SLUTT`

	tests := []struct {
		name                 string
		sosiData             string
		expectedFeatures     int
		expectedPointCount   int
		expectedLineCount    int
		expectedPolygonCount int
		checkProperties      bool
	}{
		{
			name:                 "mixed geometries",
			sosiData:             sosiData,
			expectedFeatures:     3,
			expectedPointCount:   1,
			expectedLineCount:    1,
			expectedPolygonCount: 1,
			checkProperties:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.sosiData)

			// Parse SOSI data
			doc, err := parser.Parse(reader)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Convert to GeoJSON
			fc, err := doc.ToGeoJSON()
			if err != nil {
				t.Fatalf("ToGeoJSON() error = %v", err)
			}

			// Test feature count
			if len(fc.Features) != tt.expectedFeatures {
				t.Errorf("Expected %d features, got %d", tt.expectedFeatures, len(fc.Features))
			}

			// Count geometry types
			pointCount := 0
			lineCount := 0
			polygonCount := 0

			for _, feature := range fc.Features {
				switch feature.Geometry.GeoJSONType() {
				case "Point":
					pointCount++
				case "LineString":
					lineCount++
				case "Polygon":
					polygonCount++
				}
			}

			if pointCount != tt.expectedPointCount {
				t.Errorf("Expected %d points, got %d", tt.expectedPointCount, pointCount)
			}

			if lineCount != tt.expectedLineCount {
				t.Errorf("Expected %d linestrings, got %d", tt.expectedLineCount, lineCount)
			}

			if polygonCount != tt.expectedPolygonCount {
				t.Errorf("Expected %d polygons, got %d", tt.expectedPolygonCount, polygonCount)
			}

			// Test properties if requested
			if tt.checkProperties {
				for _, feature := range fc.Features {
					// Check that SOSI ID is preserved
					if _, exists := feature.Properties["sosi_id"]; !exists {
						t.Error("Feature missing sosi_id property")
					}

					// Check that object type is preserved
					if _, exists := feature.Properties["objtype"]; !exists {
						t.Error("Feature missing objtype property")
					}

					// Check coordinate system info
					if coordSystem, exists := feature.Properties["coord_system"]; !exists || coordSystem != 25 {
						t.Errorf("Feature missing or wrong coord_system, got %v", coordSystem)
					}

					// Check SRID mapping
					if srid, exists := feature.Properties["srid"]; !exists || srid != "EPSG:32635" {
						t.Errorf("Feature missing or wrong srid, got %v", srid)
					}
				}
			}

			t.Logf("Successfully converted SOSI to GeoJSON:")
			t.Logf("- Features: %d", len(fc.Features))
			t.Logf("- Points: %d", pointCount)
			t.Logf("- LineStrings: %d", lineCount)
			t.Logf("- Polygons: %d", polygonCount)
		})
	}
}

// TestToGeoJSONWithReferences tests polygon reference resolution
func TestToGeoJSONWithReferences(t *testing.T) {
	parser := NewParser()

	// Test data with polygon references
	sosiData := `.HODE
..TEGNSETT UTF-8
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0 0
...ENHET 1
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
.KURVE 100:
..OBJTYPE Bygning
..KVALITET 22
..NØ
7000000 300000 ...KP 1
..NØ
7000010 300000 ...KP 1
.KURVE 101:
..OBJTYPE Bygning
..KVALITET 22
..NØ
7000010 300000 ...KP 1
..NØ
7000010 300010 ...KP 1
.KURVE 102:
..OBJTYPE Bygning
..KVALITET 22
..NØ
7000010 300010 ...KP 1
..NØ
7000000 300010 ...KP 1
.KURVE 103:
..OBJTYPE Bygning
..KVALITET 22
..NØ
7000000 300010 ...KP 1
..NØ
7000000 300000 ...KP 1
.FLATE 200:
..OBJTYPE Område
..KVALITET 22
..REF :100 :101 :102 :103
..NØ
7000005 300005
.SLUTT`

	reader := strings.NewReader(sosiData)

	// Parse SOSI data
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Convert to GeoJSON with references
	fc, err := doc.ToGeoJSONWithReferences()
	if err != nil {
		t.Fatalf("ToGeoJSONWithReferences() error = %v", err)
	}

	// Should have 5 features (4 KURVE + 1 FLATE)
	if len(fc.Features) != 5 {
		t.Errorf("Expected 5 features, got %d", len(fc.Features))
	}

	// Find the polygon feature
	var polygonFeature *geojson.Feature
	for _, feature := range fc.Features {
		if sosiID, exists := feature.Properties["sosi_id"]; exists && sosiID == 200 {
			if feature.Geometry.GeoJSONType() == "Polygon" {
				polygonFeature = feature
				break
			}
		}
	}

	if polygonFeature == nil {
		t.Fatal("Polygon feature not found or not converted to Polygon")
	}

	t.Logf("Successfully converted polygon with references:")
	t.Logf("- Total features: %d", len(fc.Features))
	t.Logf("- Polygon feature found with sosi_id: %v", polygonFeature.Properties["sosi_id"])
}

// TestToGeoJSONSerialization tests that the GeoJSON can be properly serialized
func TestToGeoJSONSerialization(t *testing.T) {
	parser := NewParser()

	sosiData := `.HODE
..TEGNSETT UTF-8
..TRANSPAR
...KOORDSYS 84
...ORIGO-NØ 0 0
...ENHET 1
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
.PUNKT 100:
..OBJTYPE TestPoint
..KVALITET 22
..NØ
59.0 10.0
.SLUTT`

	reader := strings.NewReader(sosiData)

	// Parse SOSI data
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Convert to GeoJSON
	fc, err := doc.ToGeoJSON()
	if err != nil {
		t.Fatalf("ToGeoJSON() error = %v", err)
	}

	// Serialize to JSON
	jsonBytes, err := json.Marshal(fc)
	if err != nil {
		t.Fatalf("Failed to marshal GeoJSON: %v", err)
	}

	// Verify it's valid JSON by unmarshaling
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Invalid JSON produced: %v", err)
	}

	// Check basic GeoJSON structure
	if result["type"] != "FeatureCollection" {
		t.Errorf("Expected type FeatureCollection, got %v", result["type"])
	}

	features, ok := result["features"].([]interface{})
	if !ok || len(features) != 1 {
		t.Errorf("Expected 1 feature, got %v", len(features))
	}

	t.Logf("Successfully serialized GeoJSON:")
	t.Logf("JSON length: %d bytes", len(jsonBytes))

	// Log first 200 chars for inspection
	jsonStr := string(jsonBytes)
	if len(jsonStr) > 200 {
		t.Logf("JSON preview: %s...", jsonStr[:200])
	} else {
		t.Logf("JSON: %s", jsonStr)
	}
}

// TestGeometryCoordinateMapping tests coordinate transformation
func TestGeometryCoordinateMapping(t *testing.T) {
	parser := NewParser()

	sosiData := `.HODE
..TEGNSETT UTF-8  
..TRANSPAR
...KOORDSYS 84
...ORIGO-NØ 0 0
...ENHET 1
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
.PUNKT 100:
..OBJTYPE TestPoint
..NØ
59.1234 10.5678
.SLUTT`

	reader := strings.NewReader(sosiData)

	// Parse SOSI data
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Convert to GeoJSON
	fc, err := doc.ToGeoJSON()
	if err != nil {
		t.Fatalf("ToGeoJSON() error = %v", err)
	}

	if len(fc.Features) != 1 {
		t.Fatalf("Expected 1 feature, got %d", len(fc.Features))
	}

	feature := fc.Features[0]
	if feature.Geometry.GeoJSONType() != "Point" {
		t.Fatalf("Expected Point geometry, got %s", feature.Geometry.GeoJSONType())
	}

	// Extract coordinates from geojson.Point
	point, ok := feature.Geometry.(geojson.Point)
	if !ok {
		t.Fatalf("Expected geojson.Point, got %T", feature.Geometry)
	}

	// Check coordinate values - Lon and Lat are explicit fields
	expectedLon := 10.5678 // longitude
	expectedLat := 59.1234 // latitude

	tolerance := 0.0001
	if abs(point.Lon-expectedLon) > tolerance {
		t.Errorf("Longitude: expected %f, got %f", expectedLon, point.Lon)
	}

	if abs(point.Lat-expectedLat) > tolerance {
		t.Errorf("Latitude: expected %f, got %f", expectedLat, point.Lat)
	}

	t.Logf("Coordinates correctly mapped: [%f, %f]", point.Lon, point.Lat)
}
