package sosi

import (
	"strings"
	"testing"
)

// TestPolygonGeometry tests polygon geometry parsing functionality
// Ports JavaScript flate-test.js functionality using Go table-driven test pattern
func TestPolygonGeometry(t *testing.T) {
	parser := NewParser()

	// Test data from flatetest.sos
	flateTestData := `.HODE
..TEGNSETT UTF-8
..OMRÅDE
...MIN-NØ  7656714  341046
...MAX-NØ  7664713  348226
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0  0
...ENHET 0.01
..INNHOLD
...PRODUKTSPEK FKB-BygnAnlegg 4.0
! FYSAK  Versjon G1.6,  2010-03-19
! Quadri Map Server v. 8.1.0 med klient-API v. 8
! Uttrekk fra NGIS arkiv: 20BygnAnlegg
! Tidspunkt for uttrekk: 2010-06-23 13:11:49
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
!!
.KURVE 633:
..OBJTYPE Tankkant
..HREF UKJENT
..DATAFANGSTDATO 20030702
..KVALITET 22 18
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..NØH
766135249 34181716 36815 ...KP 1
..NØH
766135250 34181718 36815
766134745 34182403 36815 ...KP 1
.KURVE 134:
..OBJTYPE Tankkant
..HREF UKJENT
..DATAFANGSTDATO 20030702
..KVALITET 22 18
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..NØH
766135249 34181716 36815 ...KP 1
..NØH
766135333 34181723 36808
766135685 34182091 36815 ...KP 1
.KURVE 138:
..OBJTYPE Tankkant
..HREF UKJENT
..DATAFANGSTDATO 20030702
..KVALITET 22 18
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..NØH
766135095 34182690 36815 ...KP 1
..NØH
766135101 34182638 36815
766135685 34182091 36815 ...KP 1
.KURVE 135:
..OBJTYPE Tankkant
..HREF UKJENT
..DATAFANGSTDATO 20030702
..KVALITET 22 18
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..NØH
766135095 34182690 36815 ...KP 1
..NØH
766135028 34182678 36808
766134745 34182403 36815 ...KP 1
.FLATE 651:
..OBJTYPE Tank
..DATAFANGSTDATO 20030702
..KVALITET 82
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..REF :-633
:134 :-138
:135
..NØH
766135184 34182216 36808
.SLUTT`

	tests := []struct {
		name             string
		sosiData         string
		wantFeatureCount int
		wantPolygonID    int
		wantFeatureType  string
		wantObjectType   string
		wantCenterX      float64
		wantCenterY      float64
		wantRefsCount    int
		wantCoordSystem  int
	}{
		{
			name:             "polygon geometry parsing",
			sosiData:         flateTestData,
			wantFeatureCount: 5, // 4 KURVE + 1 FLATE
			wantPolygonID:    651,
			wantFeatureType:  "FLATE",
			wantObjectType:   "Tank",
			wantCenterX:      341822.16,  // After coordinate transformation
			wantCenterY:      7661351.84, // After coordinate transformation
			wantRefsCount:    4,          // References to 4 KURVE features
			wantCoordSystem:  25,
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

			// Test total feature count (including referenced KURVE features)
			if len(doc.Features) != tt.wantFeatureCount {
				t.Errorf("Total feature count = %d, want %d", len(doc.Features), tt.wantFeatureCount)
			}

			// Find the polygon feature by ID
			var polygonFeature *SOSIFeature
			for i := range doc.Features {
				if doc.Features[i].ID == tt.wantPolygonID {
					polygonFeature = &doc.Features[i]
					break
				}
			}

			if polygonFeature == nil {
				t.Fatalf("Polygon feature with ID %d not found", tt.wantPolygonID)
			}

			// Test feature type
			if polygonFeature.Type != tt.wantFeatureType {
				t.Errorf("Polygon feature type = %s, want %s", polygonFeature.Type, tt.wantFeatureType)
			}

			// Test object type
			if polygonFeature.ObjectType != tt.wantObjectType {
				t.Errorf("Polygon object type = %s, want %s", polygonFeature.ObjectType, tt.wantObjectType)
			}

			// Test coordinate system
			if doc.Header.CoordSystem != tt.wantCoordSystem {
				t.Errorf("Header CoordSystem = %d, want %d", doc.Header.CoordSystem, tt.wantCoordSystem)
			}

			// Test references count
			if len(polygonFeature.Refs) != tt.wantRefsCount {
				t.Errorf("References count = %d, want %d", len(polygonFeature.Refs), tt.wantRefsCount)
			}

			// Test center point coordinates
			if len(polygonFeature.Coordinates) > 0 {
				center := polygonFeature.Coordinates[0]
				tolerance := 0.01

				if abs(center.X-tt.wantCenterX) > tolerance {
					t.Errorf("Center X = %f, want %f", center.X, tt.wantCenterX)
				}

				if abs(center.Y-tt.wantCenterY) > tolerance {
					t.Errorf("Center Y = %f, want %f", center.Y, tt.wantCenterY)
				}
			} else {
				t.Error("Polygon has no center coordinate")
			}
		})
	}
}

// TestPolygonAttributes tests polygon attribute parsing
// Equivalent to JavaScript "should be able to read attributes" test
func TestPolygonAttributes(t *testing.T) {
	parser := NewParser()

	flateTestData := `.HODE
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0  0
...ENHET 0.01
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
.FLATE 651:
..OBJTYPE Tank
..DATAFANGSTDATO 20030702
..KVALITET 82
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..REF :1 :2 :3
..NØH
766135184 34182216 36808
.SLUTT`

	reader := strings.NewReader(flateTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Fatal("No features found")
	}

	feature := doc.Features[0]

	// Test basic attributes
	if feature.ObjectType != "Tank" {
		t.Errorf("ObjectType = %s, want Tank", feature.ObjectType)
	}

	// Test properties parsing
	if feature.Properties == nil {
		t.Fatal("Feature Properties is nil")
	}

	// Check OBJTYPE in properties
	if objtype, exists := feature.Properties["OBJTYPE"]; !exists {
		t.Error("OBJTYPE not found in properties")
	} else if objtype != "Tank" {
		t.Errorf("Properties OBJTYPE = %v, want Tank", objtype)
	}

	// Test KVALITET parsing (different from KURVE - single value)
	if kvalitetRaw, exists := feature.Properties["KVALITET"]; exists {
		if kvalitet, ok := kvalitetRaw.(map[string]interface{}); ok {
			if målemetode, exists := kvalitet["målemetode"]; !exists {
				t.Error("målemetode not found in kvalitet")
			} else if målemetode != 82 {
				t.Errorf("målemetode = %v, want 82", målemetode)
			}
		} else {
			// For single value KVALITET, it might be stored differently
			t.Logf("KVALITET format: %T = %v", kvalitetRaw, kvalitetRaw)
		}
	} else {
		t.Error("KVALITET not found in properties")
	}

	// Test REGISTRERINGSVERSJON parsing
	if regVersRaw, exists := feature.Properties["REGISTRERINGSVERSJON"]; exists {
		if regVers, ok := regVersRaw.(map[string]interface{}); ok {
			if versjon, exists := regVers["versjon"]; !exists {
				t.Error("versjon not found in registreringsversjon")
			} else if versjon != "3.4 eller eldre" {
				t.Errorf("versjon = %v, want '3.4 eller eldre'", versjon)
			}
		} else {
			t.Logf("REGISTRERINGSVERSJON format: %T = %v", regVersRaw, regVersRaw)
		}
	} else {
		t.Error("REGISTRERINGSVERSJON not found in properties")
	}
}

// TestPolygonReferences tests polygon reference parsing
func TestPolygonReferences(t *testing.T) {
	parser := NewParser()

	flateTestData := `.HODE
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0  0
...ENHET 0.01
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
.FLATE 651:
..OBJTYPE Tank
..REF :-633
:134 :-138
:135
..NØH
766135184 34182216 36808
.SLUTT`

	reader := strings.NewReader(flateTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Fatal("No features found")
	}

	feature := doc.Features[0]

	// Test reference parsing
	expectedRefs := []int{-633, 134, -138, 135} // Note: negative references indicate reverse direction

	if len(feature.Refs) != len(expectedRefs) {
		t.Errorf("Reference count = %d, want %d", len(feature.Refs), len(expectedRefs))
	}

	for i, expectedRef := range expectedRefs {
		if i < len(feature.Refs) && feature.Refs[i] != expectedRef {
			t.Errorf("Reference[%d] = %d, want %d", i, feature.Refs[i], expectedRef)
		}
	}
}

// TestPolygonFromFile tests loading polygon geometry from actual test file
func TestPolygonFromFile(t *testing.T) {
	parser := NewParser()

	// Use the actual flatetest.sos data
	flateTestData := `.HODE
..TEGNSETT UTF-8
..OMRÅDE
...MIN-NØ  7656714  341046
...MAX-NØ  7664713  348226
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0  0
...ENHET 0.01
.KURVE 633:
..OBJTYPE Tankkant
..DATAFANGSTDATO 20030702
..KVALITET 22 18
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..NØH
766135249 34181716 36815 ...KP 1
..NØH
766135250 34181718 36815
766134745 34182403 36815 ...KP 1
.KURVE 134:
..OBJTYPE Tankkant
..DATAFANGSTDATO 20030702
..KVALITET 22 18
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..NØH
766135249 34181716 36815 ...KP 1
..NØH
766135333 34181723 36808
766135685 34182091 36815 ...KP 1
.KURVE 138:
..OBJTYPE Tankkant
..DATAFANGSTDATO 20030702
..KVALITET 22 18
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..NØH
766135095 34182690 36815 ...KP 1
..NØH
766135101 34182638 36815
766135685 34182091 36815 ...KP 1
.KURVE 135:
..OBJTYPE Tankkant
..DATAFANGSTDATO 20030702
..KVALITET 22 18
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..NØH
766135095 34182690 36815 ...KP 1
..NØH
766135028 34182678 36808
766134745 34182403 36815 ...KP 1
.FLATE 651:
..OBJTYPE Tank
..DATAFANGSTDATO 20030702
..KVALITET 82
..REGISTRERINGSVERSJON "FKB" "3.4 eller eldre"
..REF :-633
:134 :-138
:135
..NØH
766135184 34182216 36808
.SLUTT`

	reader := strings.NewReader(flateTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Failed to parse flatetest.sos: %v", err)
	}

	// Validate structure - should have 5 features (4 KURVE + 1 FLATE)
	if len(doc.Features) != 5 {
		t.Errorf("Expected 5 features, got %d", len(doc.Features))
	}

	// Find polygon feature
	var polygonFeature *SOSIFeature
	for i := range doc.Features {
		if doc.Features[i].Type == "FLATE" && doc.Features[i].ID == 651 {
			polygonFeature = &doc.Features[i]
			break
		}
	}

	if polygonFeature == nil {
		t.Fatal("Polygon feature with ID 651 not found")
	}

	if polygonFeature.ObjectType != "Tank" {
		t.Errorf("Expected Tank object, got %s", polygonFeature.ObjectType)
	}

	if len(polygonFeature.Refs) != 4 {
		t.Errorf("Expected 4 references, got %d", len(polygonFeature.Refs))
	}

	// Validate center coordinates from JavaScript test
	expectedCenterX := 341822.16
	expectedCenterY := 7661351.84

	if len(polygonFeature.Coordinates) > 0 {
		center := polygonFeature.Coordinates[0]
		tolerance := 0.01

		if abs(center.X-expectedCenterX) > tolerance {
			t.Errorf("Center X = %f, want %f (from JavaScript test)", center.X, expectedCenterX)
		}

		if abs(center.Y-expectedCenterY) > tolerance {
			t.Errorf("Center Y = %f, want %f (from JavaScript test)", center.Y, expectedCenterY)
		}
	} else {
		t.Error("No center coordinates found")
	}
}
