package sosi

import (
	"strings"
	"testing"
)

// TestPolygonHoles tests complex polygon hole resolution functionality
// Ports JavaScript island-test.js functionality using Go table-driven test pattern
func TestPolygonHoles(t *testing.T) {
	parser := NewParser()

	// Test data from flate_oy.sos
	polygonHoleTestData := `.HODE
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
.KURVE 100:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7000000 300000 ...KP 1
..NØ
7001000 300000 ...KP 1
.KURVE 101:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7001000 300000 ...KP 1
..NØ
7001000 301000 ...KP 1
.KURVE 102:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7001000 301000 ...KP 1
..NØ
7000000 301000 ...KP 1
.KURVE 103:
..OBJTYPE Bambus
..KVALITET 22 18
..NØ
7000000 301000 ...KP 1
..NØ
7000000 300000 ...KP 1
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
.FLATE 500:
..OBJTYPE Mahogney
..KVALITET 82
..REF :200 :201 :202 :203
..NØ
7000015 300015
.FLATE 400:
..OBJTYPE Mahogney
..KVALITET 82
..REF :100 :101 :102 :103 (:500)
..NØ
7000010 300010
.FLATE 600:
..OBJTYPE Mahogney
..KVALITET 82
..REF :100 :101 :102 :103 (:200 :201 :202 :203)
..NØ
7000010 300010
.SLUTT`

	tests := []struct {
		name              string
		sosiData          string
		wantFeatureCount  int
		wantPolygonID     int
		wantFeatureType   string
		wantObjectType    string
		wantOuterRingRefs int
		wantHoleCount     int
		wantTotalRefs     int
		wantCenterX       float64
		wantCenterY       float64
	}{
		{
			name:              "polygon with FLATE hole reference",
			sosiData:          polygonHoleTestData,
			wantFeatureCount:  11, // 8 KURVE + 3 FLATE
			wantPolygonID:     400,
			wantFeatureType:   "FLATE",
			wantObjectType:    "Mahogney",
			wantOuterRingRefs: 4, // :100 :101 :102 :103
			wantHoleCount:     1, // (:500)
			wantTotalRefs:     5, // 4 outer + 1 hole reference
			wantCenterX:       300010,
			wantCenterY:       7000010,
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

			// Test total reference count
			if len(polygonFeature.Refs) != tt.wantTotalRefs {
				t.Errorf("Total references count = %d, want %d", len(polygonFeature.Refs), tt.wantTotalRefs)
			}

			// Test polygon hole structure
			if len(polygonFeature.OuterRing) != tt.wantOuterRingRefs {
				t.Errorf("Outer ring references count = %d, want %d", len(polygonFeature.OuterRing), tt.wantOuterRingRefs)
			}

			if len(polygonFeature.Holes) != tt.wantHoleCount {
				t.Errorf("Hole count = %d, want %d", len(polygonFeature.Holes), tt.wantHoleCount)
			}

			t.Logf("Outer ring references: %v", polygonFeature.OuterRing)
			t.Logf("Hole references: %v", polygonFeature.Holes)
			t.Logf("All references (backward compatibility): %v", polygonFeature.Refs)

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

// TestPolygonDirectHoleReference tests polygon with direct KURVE hole references
func TestPolygonDirectHoleReference(t *testing.T) {
	parser := NewParser()

	// Test FLATE 600 with direct KURVE hole references
	polygonDirectHoleData := `.HODE
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
.KURVE 202:
..OBJTYPE Bambus
..NØ
7000020 300020 ...KP 1
..NØ
7000010 300020 ...KP 1
.KURVE 203:
..OBJTYPE Bambus
..NØ
7000010 300020 ...KP 1
..NØ
7000010 300010 ...KP 1
.FLATE 600:
..OBJTYPE Mahogney
..KVALITET 82
..REF :100 :101 :102 :103 (:200 :201 :202 :203)
..NØ
7000010 300010
.SLUTT`

	reader := strings.NewReader(polygonDirectHoleData)
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

	// Test reference parsing - should have outer refs + hole refs
	expectedOuterRefs := 4 // :100 :101 :102 :103
	expectedHoles := 1     // (:200 :201 :202 :203)
	expectedTotalRefs := 8 // 4 outer + 4 hole references

	if len(polygonFeature.OuterRing) != expectedOuterRefs {
		t.Errorf("Outer ring reference count = %d, want %d", len(polygonFeature.OuterRing), expectedOuterRefs)
	}

	if len(polygonFeature.Holes) != expectedHoles {
		t.Errorf("Hole count = %d, want %d", len(polygonFeature.Holes), expectedHoles)
	}

	if len(polygonFeature.Refs) != expectedTotalRefs {
		t.Errorf("Total reference count = %d, want %d", len(polygonFeature.Refs), expectedTotalRefs)
	}

	// Test that the hole has 4 references
	if len(polygonFeature.Holes) > 0 && len(polygonFeature.Holes[0]) != 4 {
		t.Errorf("First hole reference count = %d, want 4", len(polygonFeature.Holes[0]))
	}

	t.Logf("Polygon 600 Outer Ring: %v", polygonFeature.OuterRing)
	t.Logf("Polygon 600 Holes: %v", polygonFeature.Holes)
	t.Logf("Polygon 600 All Refs: %v", polygonFeature.Refs)
}

// TestPolygonHoleStructure tests the enhanced SOSIFeature structure for holes
func TestPolygonHoleStructure(t *testing.T) {
	parser := NewParser()

	// Simple test with polygon having one hole
	testData := `.HODE
..TRANSPAR
...KOORDSYS 25
...ORIGO-NØ 0 0
...ENHET 1
..SOSI-VERSJON 4.0
..SOSI-NIVÅ 4
.FLATE 400:
..OBJTYPE Test
..REF :100 :101 :102 :103 (:500)
..NØ
7000010 300010
.SLUTT`

	reader := strings.NewReader(testData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Fatal("No features found")
	}

	feature := doc.Features[0]

	// Test that SOSIFeature has the expected structure
	if feature.OuterRing == nil {
		t.Error("OuterRing should not be nil")
	}

	if feature.Holes == nil {
		t.Error("Holes should not be nil (can be empty slice)")
	}

	// Test specific values
	expectedOuterRing := []int{100, 101, 102, 103}
	if !equalIntSlice(feature.OuterRing, expectedOuterRing) {
		t.Errorf("OuterRing = %v, want %v", feature.OuterRing, expectedOuterRing)
	}

	if len(feature.Holes) != 1 {
		t.Errorf("Holes count = %d, want 1", len(feature.Holes))
	}

	if len(feature.Holes) > 0 {
		expectedHole := []int{500}
		if !equalIntSlice(feature.Holes[0], expectedHole) {
			t.Errorf("Holes[0] = %v, want %v", feature.Holes[0], expectedHole)
		}
	}

	// Test backward compatibility - all refs should be present
	expectedAllRefs := []int{100, 101, 102, 103, 500}
	if !equalIntSlice(feature.Refs, expectedAllRefs) {
		t.Errorf("Refs (backward compatibility) = %v, want %v", feature.Refs, expectedAllRefs)
	}
}

// equalIntSlice compares two int slices for equality
func equalIntSlice(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
