package sosi

import (
	"strings"
	"testing"
)

// TestLineStringGeometry tests linestring geometry parsing functionality
// Ports JavaScript kurve-test.js functionality using Go table-driven test pattern
func TestLineStringGeometry(t *testing.T) {
	parser := NewParser()

	// Test data from kurvetest.sos
	kurveTestData := `.HODE                                           !SOSI-filas hode.
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
.KURVE 250:                                     !Geometritype KURVE.
..OBJTYPE EiendomsGrense
..KVALITET 40 58
..NØ
23456 2345
23460 2345
23470 2346
23480 2347
23490 2350
23500 2366
23512 2345
23565 2370
23460 2356 ...KP 1                              !Knutepunkt
..NØ
23500 2350
.SLUTT`

	tests := []struct {
		name              string
		sosiData          string
		wantFeatureCount  int
		wantFeatureID     int
		wantFeatureType   string
		wantObjectType    string
		wantCoordCount    int
		wantFirstCoord    Coordinate
		wantLastCoord     Coordinate
		wantTiePointCount int
		wantTiePoint      Coordinate
	}{
		{
			name:             "linestring geometry parsing",
			sosiData:         kurveTestData,
			wantFeatureCount: 1,
			wantFeatureID:    250,
			wantFeatureType:  "KURVE",
			wantObjectType:   "EiendomsGrense",
			wantCoordCount:   10,
			wantFirstCoord: Coordinate{
				X: 10023.45,  // (2345 * 0.01) + 10000
				Y: 100234.56, // (23456 * 0.01) + 100000
				Z: 0,
			},
			wantLastCoord: Coordinate{
				X: 10023.50,  // (2350 * 0.01) + 10000
				Y: 100235.00, // (23500 * 0.01) + 100000
				Z: 0,
			},
			wantTiePointCount: 1,
			wantTiePoint: Coordinate{
				X: 10023.56,  // (2356 * 0.01) + 10000
				Y: 100234.60, // (23460 * 0.01) + 100000
				Z: 0,
			},
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

			// Test coordinate count
			if len(feature.Coordinates) != tt.wantCoordCount {
				t.Errorf("Coordinate count = %d, want %d", len(feature.Coordinates), tt.wantCoordCount)
			}

			if len(feature.Coordinates) < 2 {
				t.Fatal("Feature has insufficient coordinates for linestring")
			}

			// Test first coordinate
			firstCoord := feature.Coordinates[0]
			tolerance := 0.001

			if abs(firstCoord.X-tt.wantFirstCoord.X) > tolerance {
				t.Errorf("First coordinate X = %f, want %f", firstCoord.X, tt.wantFirstCoord.X)
			}

			if abs(firstCoord.Y-tt.wantFirstCoord.Y) > tolerance {
				t.Errorf("First coordinate Y = %f, want %f", firstCoord.Y, tt.wantFirstCoord.Y)
			}

			// Test last coordinate
			lastCoord := feature.Coordinates[len(feature.Coordinates)-1]

			if abs(lastCoord.X-tt.wantLastCoord.X) > tolerance {
				t.Errorf("Last coordinate X = %f, want %f", lastCoord.X, tt.wantLastCoord.X)
			}

			if abs(lastCoord.Y-tt.wantLastCoord.Y) > tolerance {
				t.Errorf("Last coordinate Y = %f, want %f", lastCoord.Y, tt.wantLastCoord.Y)
			}

			// Test tie points (knutepunkter)
			tiePointCount := 0
			var foundTiePoint Coordinate

			for _, coord := range feature.Coordinates {
				if coord.TiePointCode != 0 {
					tiePointCount++
					foundTiePoint = coord
				}
			}

			if tiePointCount != tt.wantTiePointCount {
				t.Errorf("Tie point count = %d, want %d", tiePointCount, tt.wantTiePointCount)
			}

			if tt.wantTiePointCount > 0 {
				if abs(foundTiePoint.X-tt.wantTiePoint.X) > tolerance {
					t.Errorf("Tie point X = %f, want %f", foundTiePoint.X, tt.wantTiePoint.X)
				}

				if abs(foundTiePoint.Y-tt.wantTiePoint.Y) > tolerance {
					t.Errorf("Tie point Y = %f, want %f", foundTiePoint.Y, tt.wantTiePoint.Y)
				}

				if foundTiePoint.TiePointCode != 1 {
					t.Errorf("Tie point code = %d, want 1", foundTiePoint.TiePointCode)
				}
			}
		})
	}
}

// TestLineStringAttributes tests linestring attribute parsing
// Equivalent to JavaScript "should be able to read attributes" test
func TestLineStringAttributes(t *testing.T) {
	parser := NewParser()

	kurveTestData := `.HODE
..TRANSPAR
...KOORDSYS 5
...ORIGO-NØ 100000 10000
...ENHET 0.010
..SOSI-VERSJON 4.5
..SOSI-NIVÅ 5
.KURVE 250:
..OBJTYPE EiendomsGrense
..KVALITET 40 58
..NØ
23456 2345
23460 2345
.SLUTT`

	reader := strings.NewReader(kurveTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Fatal("No features found")
	}

	feature := doc.Features[0]

	// Test basic attributes
	if feature.ObjectType != "EiendomsGrense" {
		t.Errorf("ObjectType = %s, want EiendomsGrense", feature.ObjectType)
	}

	// Test properties parsing
	if feature.Properties == nil {
		t.Fatal("Feature Properties is nil")
	}

	// Check OBJTYPE in properties
	if objtype, exists := feature.Properties["OBJTYPE"]; !exists {
		t.Error("OBJTYPE not found in properties")
	} else if objtype != "EiendomsGrense" {
		t.Errorf("Properties OBJTYPE = %v, want EiendomsGrense", objtype)
	}

	// Test nested kvalitet attributes (from JavaScript test)
	if kvalitetRaw, exists := feature.Properties["KVALITET"]; exists {
		if kvalitet, ok := kvalitetRaw.(map[string]interface{}); ok {
			if målemetode, exists := kvalitet["målemetode"]; !exists {
				t.Error("målemetode not found in kvalitet")
			} else if målemetode != 40 {
				t.Errorf("målemetode = %v, want 40", målemetode)
			}

			if nøyaktighet, exists := kvalitet["nøyaktighet"]; !exists {
				t.Error("nøyaktighet not found in kvalitet")
			} else if nøyaktighet != 58 {
				t.Errorf("nøyaktighet = %v, want 58", nøyaktighet)
			}
		} else {
			t.Errorf("KVALITET is not a map: %T", kvalitetRaw)
		}
	} else {
		t.Error("KVALITET not found in properties")
	}
}

// TestLineStringCoordinateSequence tests specific coordinate sequence from kurve-test.js
func TestLineStringCoordinateSequence(t *testing.T) {
	parser := NewParser()

	kurveTestData := `.HODE
..TRANSPAR
...KOORDSYS 5
...ORIGO-NØ 100000 10000
...ENHET 0.010
..SOSI-VERSJON 4.5
..SOSI-NIVÅ 5
.KURVE 250:
..OBJTYPE EiendomsGrense
..KVALITET 40 58
..NØ
23456 2345
23460 2345
23470 2346
23480 2347
23490 2350
23500 2366
23512 2345
23565 2370
23460 2356 ...KP 1
..NØ
23500 2350
.SLUTT`

	// Expected coordinates from JavaScript test (kurve-test.js lines 46-65)
	expectedCoords := []Coordinate{
		{X: 10023.45, Y: 100234.56}, // kurve[0]
		{X: 10023.45, Y: 100234.60}, // kurve[1]
		{X: 10023.46, Y: 100234.70}, // kurve[2]
		{X: 10023.47, Y: 100234.80}, // kurve[3]
		{X: 10023.50, Y: 100234.90}, // kurve[4]
		{X: 10023.66, Y: 100235.00}, // kurve[5]
		{X: 10023.45, Y: 100235.12}, // kurve[6]
		{X: 10023.70, Y: 100235.65}, // kurve[7]
		{X: 10023.56, Y: 100234.60}, // kurve[8] - tie point
		{X: 10023.50, Y: 100235.00}, // kurve[9]
	}

	reader := strings.NewReader(kurveTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Fatal("No features found")
	}

	feature := doc.Features[0]

	if len(feature.Coordinates) != len(expectedCoords) {
		t.Errorf("Coordinate count = %d, want %d", len(feature.Coordinates), len(expectedCoords))
	}

	tolerance := 0.001

	for i, expectedCoord := range expectedCoords {
		if i >= len(feature.Coordinates) {
			t.Errorf("Missing coordinate at index %d", i)
			continue
		}

		actualCoord := feature.Coordinates[i]

		if abs(actualCoord.X-expectedCoord.X) > tolerance {
			t.Errorf("Coordinate[%d] X = %f, want %f", i, actualCoord.X, expectedCoord.X)
		}

		if abs(actualCoord.Y-expectedCoord.Y) > tolerance {
			t.Errorf("Coordinate[%d] Y = %f, want %f", i, actualCoord.Y, expectedCoord.Y)
		}
	}

	// Verify tie point at index 8 (kurve[8])
	tiePointCoord := feature.Coordinates[8]
	if tiePointCoord.TiePointCode != 1 {
		t.Errorf("Coordinate[8] TiePointCode = %d, want 1", tiePointCoord.TiePointCode)
	}
}

// TestLineStringFromFile tests loading linestring geometry from actual test file
func TestLineStringFromFile(t *testing.T) {
	parser := NewParser()

	// Use the actual kurvetest.sos data
	kurveTestData := `.HODE                                           !SOSI-filas hode.
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
.KURVE 250:                                     !Geometritype KURVE.
..OBJTYPE EiendomsGrense
..KVALITET 40 58
..NØ
23456 2345
23460 2345
23470 2346
23480 2347
23490 2350
23500 2366
23512 2345
23565 2370
23460 2356 ...KP 1                              !Knutepunkt
..NØ
23500 2350
.SLUTT`

	reader := strings.NewReader(kurveTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Failed to parse kurvetest.sos: %v", err)
	}

	// Validate structure
	if len(doc.Features) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(doc.Features))
	}

	feature := doc.Features[0]
	if feature.Type != "KURVE" {
		t.Errorf("Expected KURVE feature, got %s", feature.Type)
	}

	if feature.ID != 250 {
		t.Errorf("Expected feature ID 250, got %d", feature.ID)
	}

	if len(feature.Coordinates) != 10 {
		t.Errorf("Expected 10 coordinates, got %d", len(feature.Coordinates))
	}

	// Verify first and last coordinates match JavaScript test expectations
	if len(feature.Coordinates) >= 2 {
		firstCoord := feature.Coordinates[0]
		lastCoord := feature.Coordinates[len(feature.Coordinates)-1]

		tolerance := 0.001

		// From JavaScript test
		expectedFirstX, expectedFirstY := 10023.45, 100234.56
		expectedLastX, expectedLastY := 10023.50, 100235.00

		if abs(firstCoord.X-expectedFirstX) > tolerance || abs(firstCoord.Y-expectedFirstY) > tolerance {
			t.Errorf("First coordinate = (%f, %f), want (%f, %f)", firstCoord.X, firstCoord.Y, expectedFirstX, expectedFirstY)
		}

		if abs(lastCoord.X-expectedLastX) > tolerance || abs(lastCoord.Y-expectedLastY) > tolerance {
			t.Errorf("Last coordinate = (%f, %f), want (%f, %f)", lastCoord.X, lastCoord.Y, expectedLastX, expectedLastY)
		}
	}

	// Verify tie point
	tiePointFound := false
	for i, coord := range feature.Coordinates {
		if coord.TiePointCode == 1 {
			tiePointFound = true
			expectedX, expectedY := 10023.56, 100234.60
			tolerance := 0.001

			if abs(coord.X-expectedX) > tolerance || abs(coord.Y-expectedY) > tolerance {
				t.Errorf("Tie point coordinate[%d] = (%f, %f), want (%f, %f)", i, coord.X, coord.Y, expectedX, expectedY)
			}
			break
		}
	}

	if !tiePointFound {
		t.Error("No tie point found with code 1")
	}
}
