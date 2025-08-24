package sosi

import (
	"strings"
	"testing"
)

// TestNonLinearOrder tests parsing SOSI files with non-linear feature references
// Ports JavaScript non-linear-order-test.js functionality using Go table-driven test pattern
func TestNonLinearOrder(t *testing.T) {
	parser := NewParser()

	// Test data from non-linear.sos - FLATE references KURVE features defined later
	nonLinearTestData := `.HODE
..TEGNSETT UTF-8
..OMRÅDE
...MIN-NØ  7000000  300000
...MAX-NØ  8000000  400000
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0 0
...ENHET 1
.FLATE 500:
..OBJTYPE Mahogney
..KVALITET 82
..REF :200 :201 :202 :203
..NØ
7000015 300015
.KURVE 100:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7000000 300000 ...KP 1
..NØ
7001000 300000 ...KP 1
.KURVE 200:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7000010 300010 ...KP 1
..NØ
7000020 300010 ...KP 1
.KURVE 201:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7000020 300010 ...KP 1
..NØ
7000020 300020 ...KP 1
.KURVE 202:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7000020 300020 ...KP 1
..NØ
7000010 300020 ...KP 1
.KURVE 203:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7000010 300020 ...KP 1
..NØ
7000010 300010 ...KP 1
.SLUTT`

	tests := []struct {
		name             string
		sosiData         string
		wantFeatureCount int
		wantPolygonID    int
		wantFeatureType  string
		wantObjectType   string
		wantRefsCount    int
		wantCenterX      float64
		wantCenterY      float64
	}{
		{
			name:             "non-linear forward references",
			sosiData:         nonLinearTestData,
			wantFeatureCount: 6, // 1 FLATE + 5 KURVE (100, 200, 201, 202, 203)
			wantPolygonID:    500,
			wantFeatureType:  "FLATE",
			wantObjectType:   "Mahogney",
			wantRefsCount:    4, // References to 4 KURVE features
			wantCenterX:      300015,
			wantCenterY:      7000015,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.sosiData)

			// Parse SOSI data - this should handle forward references
			doc, err := parser.Parse(reader)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Test total feature count
			if len(doc.Features) != tt.wantFeatureCount {
				t.Errorf("Feature count = %d, want %d", len(doc.Features), tt.wantFeatureCount)
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

			// Test references count - the key test for non-linear parsing
			if len(polygonFeature.Refs) != tt.wantRefsCount {
				t.Errorf("References count = %d, want %d", len(polygonFeature.Refs), tt.wantRefsCount)
			}

			// Test specific forward references
			expectedRefs := []int{200, 201, 202, 203}
			for i, expectedRef := range expectedRefs {
				if i < len(polygonFeature.Refs) {
					if polygonFeature.Refs[i] != expectedRef {
						t.Errorf("Reference[%d] = %d, want %d", i, polygonFeature.Refs[i], expectedRef)
					}
				}
			}

			// Test that referenced KURVE features were parsed correctly
			referencedFeatures := make(map[int]*SOSIFeature)
			for i := range doc.Features {
				feature := &doc.Features[i]
				if feature.Type == "KURVE" {
					referencedFeatures[feature.ID] = feature
				}
			}

			// Verify all referenced features exist
			for _, refID := range polygonFeature.Refs {
				if _, exists := referencedFeatures[refID]; !exists {
					t.Errorf("Referenced KURVE feature %d not found", refID)
				}
			}

			t.Logf("Successfully parsed non-linear references: %v", polygonFeature.Refs)
			t.Logf("Found %d KURVE features: %v", len(referencedFeatures), getFeatureIDs(referencedFeatures))

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
			}
		})
	}
}

// TestComplexNonLinearReferences tests complex cases with mixed forward/backward references
func TestComplexNonLinearReferences(t *testing.T) {
	parser := NewParser()

	// Complex case with both forward and backward references in the same FLATE
	complexTestData := `.HODE
..TEGNSETT UTF-8
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0 0
...ENHET 1
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
.KURVE 100:
..OBJTYPE Bambus
..NØ
7000000 300000 ...KP 1
..NØ
7001000 300000 ...KP 1
.FLATE 600:
..OBJTYPE Mahogney
..REF :100 :101 :102 :103 (:200 :201)
..NØ
7000010 300010
.KURVE 101:
..OBJTYPE Bambus
..NØ
7001000 300000 ...KP 1
..NØ
7001000 301000 ...KP 1
.KURVE 102:
..OBJTYPE Bambus
..NØ
7001000 301000 ...KP 1
..NØ
7000000 301000 ...KP 1
.KURVE 103:
..OBJTYPE Bambus
..NØ
7000000 301000 ...KP 1
..NØ
7000000 300000 ...KP 1
.KURVE 200:
..OBJTYPE Bambus
..NØ
7000010 300010 ...KP 1
..NØ
7000020 300010 ...KP 1
.KURVE 201:
..OBJTYPE Bambus
..NØ
7000020 300010 ...KP 1
..NØ
7000020 300020 ...KP 1
.SLUTT`

	reader := strings.NewReader(complexTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find polygon feature
	var polygonFeature *SOSIFeature
	for i := range doc.Features {
		if doc.Features[i].ID == 600 {
			polygonFeature = &doc.Features[i]
			break
		}
	}

	if polygonFeature == nil {
		t.Fatal("Polygon feature with ID 600 not found")
	}

	// Test that both backward (100) and forward (101, 102, 103, 200, 201) references work
	expectedOuterRing := []int{100, 101, 102, 103}
	expectedHoles := [][]int{{200, 201}}
	expectedTotalRefs := 6 // 4 outer + 2 hole

	if len(polygonFeature.OuterRing) != len(expectedOuterRing) {
		t.Errorf("Outer ring count = %d, want %d", len(polygonFeature.OuterRing), len(expectedOuterRing))
	}

	if len(polygonFeature.Holes) != len(expectedHoles) {
		t.Errorf("Holes count = %d, want %d", len(polygonFeature.Holes), len(expectedHoles))
	}

	if len(polygonFeature.Refs) != expectedTotalRefs {
		t.Errorf("Total references count = %d, want %d", len(polygonFeature.Refs), expectedTotalRefs)
	}

	t.Logf("Complex non-linear outer ring: %v", polygonFeature.OuterRing)
	t.Logf("Complex non-linear holes: %v", polygonFeature.Holes)
	t.Logf("Complex non-linear all refs: %v", polygonFeature.Refs)
}

// TestFeatureOrdering tests that features are parsed in document order regardless of references
func TestFeatureOrdering(t *testing.T) {
	parser := NewParser()

	testData := `.HODE
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0 0
...ENHET 1
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
.FLATE 500:
..OBJTYPE First
..REF :200
..NØ
7000015 300015
.KURVE 100:
..OBJTYPE Second
..NØ
7000000 300000 ...KP 1
..NØ
7001000 300000 ...KP 1
.KURVE 200:
..OBJTYPE Third
..NØ
7000010 300010 ...KP 1
..NØ
7000020 300010 ...KP 1
.SLUTT`

	reader := strings.NewReader(testData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Test that features are in consistent ID-sorted order
	if len(doc.Features) != 3 {
		t.Errorf("Feature count = %d, want 3", len(doc.Features))
	}

	// Features are sorted by ID for deterministic parsing
	expectedOrder := []struct {
		ID   int
		Type string
		Obj  string
	}{
		{100, "KURVE", "Second"},
		{200, "KURVE", "Third"},
		{500, "FLATE", "First"},
	}

	for i, expected := range expectedOrder {
		if i >= len(doc.Features) {
			t.Fatalf("Missing feature at index %d", i)
		}

		feature := doc.Features[i]
		if feature.ID != expected.ID {
			t.Errorf("Feature[%d] ID = %d, want %d", i, feature.ID, expected.ID)
		}

		if feature.Type != expected.Type {
			t.Errorf("Feature[%d] Type = %s, want %s", i, feature.Type, expected.Type)
		}

		if feature.ObjectType != expected.Obj {
			t.Errorf("Feature[%d] ObjectType = %s, want %s", i, feature.ObjectType, expected.Obj)
		}
	}
}

// getFeatureIDs extracts feature IDs from a map for logging
func getFeatureIDs(features map[int]*SOSIFeature) []int {
	var ids []int
	for id := range features {
		ids = append(ids, id)
	}
	return ids
}
