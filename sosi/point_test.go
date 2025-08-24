package sosi

import (
	"fmt"
	"strings"
	"testing"
)

// TestPointGeometry tests point geometry parsing functionality
// Ports JavaScript punkt-test.js functionality using Go table-driven test pattern
func TestPointGeometry(t *testing.T) {
	parser := NewParser()

	// Test data from punkttest.sos
	punktTestData := `.HODE                                           !SOSI-filas hode.
..TRANSPAR
...KOORDSYS 5
...ORIGO-NØ 100000 10000
...ENHET 0.010
...VERT-DATUM NN54 SJØ0
..OMRÅDE
...MIN-NØ 100000 10000
...MAX-NØ 102400 13200
..SOSI-VERSJON 4.5
..SOSI-NIVÅ    5
!
..VERIFISERINGSDATO 19890623
..KVALITET 11 300
!
! data er bare delvis synfart.                  !Kommentar i hode
!
..EIER "Statens kartverk"
..PRODUSENT "SØRKART A/S"
..OBJEKTKATALOG Eksempel 4.5
!
.PUNKT 1:                                       !Geometritype PUNKT.
..OBJTYPE Fastmerke
! NB! Flere påkrevde egenskaper for Fastmerke er utelatt i dette eksempelet
..NØ
23456 2345
.SLUTT`

	tests := []struct {
		name             string
		sosiData         string
		wantFeatureCount int
		wantFeatureID    int
		wantFeatureType  string
		wantObjectType   string
		wantCoordinates  Coordinate
		wantCoordSystem  int
		wantUnit         float64
	}{
		{
			name:             "punkt geometry parsing",
			sosiData:         punktTestData,
			wantFeatureCount: 1,
			wantFeatureID:    1,
			wantFeatureType:  "PUNKT",
			wantObjectType:   "Fastmerke",
			wantCoordinates: Coordinate{
				X: 10023.45,  // (2345 * 0.01) + 10000 = 23.45 + 10000
				Y: 100234.56, // (23456 * 0.01) + 100000 = 234.56 + 100000
				Z: 0,
			},
			wantCoordSystem: 5,
			wantUnit:        0.01,
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

			// Test feature count
			if len(doc.Features) != tt.wantFeatureCount {
				t.Errorf("Feature count = %d, want %d", len(doc.Features), tt.wantFeatureCount)
			}

			if len(doc.Features) == 0 {
				t.Fatal("No features found")
			}

			feature := doc.Features[0]

			// Test feature ID
			if feature.ID != tt.wantFeatureID {
				t.Errorf("Feature ID = %d, want %d", feature.ID, tt.wantFeatureID)
			}

			// Test feature type
			if feature.Type != tt.wantFeatureType {
				t.Errorf("Feature Type = %s, want %s", feature.Type, tt.wantFeatureType)
			}

			// Test object type
			if feature.ObjectType != tt.wantObjectType {
				t.Errorf("Feature ObjectType = %s, want %s", feature.ObjectType, tt.wantObjectType)
			}

			// Test coordinates
			if len(feature.Coordinates) == 0 {
				t.Fatal("Feature has no coordinates")
			}

			coord := feature.Coordinates[0]
			tolerance := 0.001 // Allow small floating-point differences

			if abs(coord.X-tt.wantCoordinates.X) > tolerance {
				t.Errorf("Coordinate X = %f, want %f", coord.X, tt.wantCoordinates.X)
			}

			if abs(coord.Y-tt.wantCoordinates.Y) > tolerance {
				t.Errorf("Coordinate Y = %f, want %f", coord.Y, tt.wantCoordinates.Y)
			}

			if abs(coord.Z-tt.wantCoordinates.Z) > tolerance {
				t.Errorf("Coordinate Z = %f, want %f", coord.Z, tt.wantCoordinates.Z)
			}

			// Test header coordinate system
			if doc.Header.CoordSystem != tt.wantCoordSystem {
				t.Errorf("Header CoordSystem = %d, want %d", doc.Header.CoordSystem, tt.wantCoordSystem)
			}

			// Test header unit
			if abs(doc.Header.Unit-tt.wantUnit) > tolerance {
				t.Errorf("Header Unit = %f, want %f", doc.Header.Unit, tt.wantUnit)
			}
		})
	}
}

// TestPointAttributes tests point attribute parsing
// Equivalent to JavaScript "should be able to read attributes" test
func TestPointAttributes(t *testing.T) {
	parser := NewParser()

	punktTestData := `.HODE
..TRANSPAR
...KOORDSYS 5
...ORIGO-NØ 100000 10000
...ENHET 0.010
..SOSI-VERSJON 4.5
..SOSI-NIVÅ 5
.PUNKT 1:
..OBJTYPE Fastmerke
..KVALITET 82 50
..DATAFANGSTDATO 20031002
..NØ
23456 2345
.SLUTT`

	reader := strings.NewReader(punktTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Fatal("No features found")
	}

	feature := doc.Features[0]

	// Test basic attributes
	if feature.ObjectType != "Fastmerke" {
		t.Errorf("ObjectType = %s, want Fastmerke", feature.ObjectType)
	}

	// Test properties parsing
	if feature.Properties == nil {
		t.Fatal("Feature Properties is nil")
	}

	// Check that OBJTYPE is in properties
	if objtype, exists := feature.Properties["OBJTYPE"]; !exists {
		t.Error("OBJTYPE not found in properties")
	} else if objtype != "Fastmerke" {
		t.Errorf("Properties OBJTYPE = %v, want Fastmerke", objtype)
	}

	// Test that additional attributes are parsed
	t.Logf("Feature properties: %+v", feature.Properties)
}

// TestPointCoordinateTransformation tests coordinate transformation logic
func TestPointCoordinateTransformation(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name      string
		rawN      string // Raw N coordinate from SOSI
		rawO      string // Raw Ø coordinate from SOSI
		origoN    float64
		origoO    float64
		unit      float64
		expectedX float64 // Expected longitude
		expectedY float64 // Expected latitude
	}{
		{
			name:      "basic coordinate transformation",
			rawN:      "23456", // N (North/Latitude)
			rawO:      "2345",  // Ø (East/Longitude)
			origoN:    100000,
			origoO:    10000,
			unit:      0.01,
			expectedX: 10023.45,  // (2345 * 0.01) + 10000
			expectedY: 100234.56, // (23456 * 0.01) + 100000
		},
		{
			name:      "different unit",
			rawN:      "1000",
			rawO:      "500",
			origoN:    0,
			origoO:    0,
			unit:      1.0,
			expectedX: 500.0,  // (500 * 1.0) + 0
			expectedY: 1000.0, // (1000 * 1.0) + 0
		},
		{
			name:      "with origo offset",
			rawN:      "0",
			rawO:      "0",
			origoN:    59.0,
			origoO:    10.0,
			unit:      1.0,
			expectedX: 10.0, // (0 * 1.0) + 10.0
			expectedY: 59.0, // (0 * 1.0) + 59.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sosiData := fmt.Sprintf(`.HODE
..TRANSPAR
...KOORDSYS 5
...ORIGO-NØ %g %g
...ENHET %g
..SOSI-VERSJON 4.5
..SOSI-NIVÅ 5
.PUNKT 1:
..OBJTYPE Test
..NØ
%s %s
.SLUTT`, tt.origoN, tt.origoO, tt.unit, tt.rawN, tt.rawO)

			reader := strings.NewReader(sosiData)
			doc, err := parser.Parse(reader)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if len(doc.Features) == 0 || len(doc.Features[0].Coordinates) == 0 {
				t.Fatal("No coordinates found")
			}

			coord := doc.Features[0].Coordinates[0]
			tolerance := 0.001

			if abs(coord.X-tt.expectedX) > tolerance {
				t.Errorf("Coordinate X = %f, want %f", coord.X, tt.expectedX)
			}

			if abs(coord.Y-tt.expectedY) > tolerance {
				t.Errorf("Coordinate Y = %f, want %f", coord.Y, tt.expectedY)
			}
		})
	}
}

// TestPointSRIDMapping tests SRID lookup for coordinate systems
func TestPointSRIDMapping(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name         string
		coordSystem  int
		expectedSRID string
	}{
		{
			name:         "NGO coordinate system 5",
			coordSystem:  5,
			expectedSRID: "EPSG:27395",
		},
		{
			name:         "WGS84",
			coordSystem:  84,
			expectedSRID: "EPSG:4326",
		},
		{
			name:         "UTM Zone 32N",
			coordSystem:  22,
			expectedSRID: "EPSG:32632",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if coordSys, exists := parser.coordSystems[tt.coordSystem]; exists {
				if coordSys.SRID != tt.expectedSRID {
					t.Errorf("SRID = %s, want %s", coordSys.SRID, tt.expectedSRID)
				}
			} else {
				t.Errorf("Coordinate system %d not found in lookup table", tt.coordSystem)
			}
		})
	}
}

// Helper function for floating-point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestPointFromFile tests loading point geometry from actual test file
func TestPointFromFile(t *testing.T) {
	parser := NewParser()

	// Read from testdata directory
	testFile := "punkttest.sos"

	// For now, use embedded data - later this could read from testdata/punkttest.sos
	punktTestData := `.HODE                                           !SOSI-filas hode.
..TRANSPAR
...KOORDSYS 5
...ORIGO-NØ 100000 10000
...ENHET 0.010
...VERT-DATUM NN54 SJØ0
..OMRÅDE
...MIN-NØ 100000 10000
...MAX-NØ 102400 13200
..SOSI-VERSJON 4.5
..SOSI-NIVÅ    5
!
..VERIFISERINGSDATO 19890623
..KVALITET 11 300
!
! data er bare delvis synfart.                  !Kommentar i hode
!
..EIER "Statens kartverk"
..PRODUSENT "SØRKART A/S"
..OBJEKTKATALOG Eksempel 4.5
!
.PUNKT 1:                                       !Geometritype PUNKT.
..OBJTYPE Fastmerke
! NB! Flere påkrevde egenskaper for Fastmerke er utelatt i dette eksempelet
..NØ
23456 2345
.SLUTT`

	reader := strings.NewReader(punktTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", testFile, err)
	}

	// Validate structure
	if len(doc.Features) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(doc.Features))
	}

	feature := doc.Features[0]
	if feature.Type != "PUNKT" {
		t.Errorf("Expected PUNKT feature, got %s", feature.Type)
	}

	if len(feature.Coordinates) != 1 {
		t.Errorf("Expected 1 coordinate, got %d", len(feature.Coordinates))
	}

	// Validate exact coordinates from JavaScript test
	expectedX := 10023.45
	expectedY := 100234.56

	coord := feature.Coordinates[0]
	tolerance := 0.001

	if abs(coord.X-expectedX) > tolerance {
		t.Errorf("Coordinate X = %f, want %f (from JavaScript test)", coord.X, expectedX)
	}

	if abs(coord.Y-expectedY) > tolerance {
		t.Errorf("Coordinate Y = %f, want %f (from JavaScript test)", coord.Y, expectedY)
	}
}
