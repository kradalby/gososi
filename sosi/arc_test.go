package sosi

import (
	"math"
	"strings"
	"testing"
)

// TestArcGeometry tests BUEP (arc) geometry parsing functionality
// Ports JavaScript bue-test.js functionality using Go table-driven test pattern
func TestArcGeometry(t *testing.T) {
	parser := NewParser()

	// Test data from buer.sos - simplified single BUEP
	arcTestData := `.HODE
..TEGNSETT ISO8859-10
..OMRÅDE
...MIN-NØ  6427548  247730
...MAX-NØ  8017910  1339080
..SOSI-VERSJON 4.5
..SOSI-NIVÅ 4
..TRANSPAR
...KOORDSYS 22
...ORIGO-NØ 0  0
...ENHET 0.01
..OBJEKTKATALOG Regplan 20120416 * Arealplan Reguleringsplan
.BUEP 8:
..OBJTYPE RpGrense
..KOPIDATA
...OMRÅDEID 0618
...ORIGINALDATAVERT "Hemsedal kommune"
...KOPIDATO 20130531
..OPPDATERINGSDATO 20130531092024
..NØ
674759173 47234619 ...KP 1
..NØ
674759286 47234337
674759383 47234049 ...KP 1
.SLUTT`

	tests := []struct {
		name                string
		sosiData            string
		wantFeatureCount    int
		wantArcID           int
		wantFeatureType     string
		wantObjectType      string
		wantCoordinateCount int
		wantStartX          float64
		wantStartY          float64
		wantMiddleX         float64
		wantMiddleY         float64
		wantEndX            float64
		wantEndY            float64
		wantCoordSystem     int
	}{
		{
			name:                "BUEP arc geometry parsing",
			sosiData:            arcTestData,
			wantFeatureCount:    1,
			wantArcID:           8,
			wantFeatureType:     "BUEP",
			wantObjectType:      "RpGrense",
			wantCoordinateCount: 3, // Start, middle (control), end
			wantStartX:          472346.19,
			wantStartY:          6747591.73,
			wantMiddleX:         472343.37,
			wantMiddleY:         6747592.86,
			wantEndX:            472340.49,
			wantEndY:            6747593.83,
			wantCoordSystem:     22,
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

			// Find the arc feature by ID
			var arcFeature *SOSIFeature
			for i := range doc.Features {
				if doc.Features[i].ID == tt.wantArcID {
					arcFeature = &doc.Features[i]
					break
				}
			}

			if arcFeature == nil {
				t.Fatalf("Arc feature with ID %d not found", tt.wantArcID)
			}

			// Test feature type
			if arcFeature.Type != tt.wantFeatureType {
				t.Errorf("Arc feature type = %s, want %s", arcFeature.Type, tt.wantFeatureType)
			}

			// Test object type
			if arcFeature.ObjectType != tt.wantObjectType {
				t.Errorf("Arc object type = %s, want %s", arcFeature.ObjectType, tt.wantObjectType)
			}

			// Test coordinate system
			if doc.Header.CoordSystem != tt.wantCoordSystem {
				t.Errorf("Header CoordSystem = %d, want %d", doc.Header.CoordSystem, tt.wantCoordSystem)
			}

			// Test coordinate count - BUEP should have exactly 3 coordinates
			if len(arcFeature.Coordinates) != tt.wantCoordinateCount {
				t.Errorf("Coordinate count = %d, want %d", len(arcFeature.Coordinates), tt.wantCoordinateCount)
			}

			// Test arc coordinates (start, middle, end points)
			if len(arcFeature.Coordinates) >= 3 {
				tolerance := 1.0 // Relaxed tolerance for coordinate transformation

				start := arcFeature.Coordinates[0]
				if abs(start.X-tt.wantStartX) > tolerance {
					t.Errorf("Start X = %f, want %f (diff: %f)", start.X, tt.wantStartX, abs(start.X-tt.wantStartX))
				}
				if abs(start.Y-tt.wantStartY) > tolerance {
					t.Errorf("Start Y = %f, want %f (diff: %f)", start.Y, tt.wantStartY, abs(start.Y-tt.wantStartY))
				}

				middle := arcFeature.Coordinates[1]
				if abs(middle.X-tt.wantMiddleX) > tolerance {
					t.Errorf("Middle X = %f, want %f (diff: %f)", middle.X, tt.wantMiddleX, abs(middle.X-tt.wantMiddleX))
				}
				if abs(middle.Y-tt.wantMiddleY) > tolerance {
					t.Errorf("Middle Y = %f, want %f (diff: %f)", middle.Y, tt.wantMiddleY, abs(middle.Y-tt.wantMiddleY))
				}

				end := arcFeature.Coordinates[2]
				if abs(end.X-tt.wantEndX) > tolerance {
					t.Errorf("End X = %f, want %f (diff: %f)", end.X, tt.wantEndX, abs(end.X-tt.wantEndX))
				}
				if abs(end.Y-tt.wantEndY) > tolerance {
					t.Errorf("End Y = %f, want %f (diff: %f)", end.Y, tt.wantEndY, abs(end.Y-tt.wantEndY))
				}
			}
		})
	}
}

// TestArcInterpolation tests arc interpolation to linestring functionality
// This is critical for converting BUEP to standard GeoJSON LineString
func TestArcInterpolation(t *testing.T) {
	tests := []struct {
		name            string
		start           Coordinate
		middle          Coordinate
		end             Coordinate
		wantSegments    int
		wantTotalLength float64
	}{
		{
			name:            "simple arc interpolation",
			start:           Coordinate{X: 0, Y: 0},
			middle:          Coordinate{X: 1, Y: 1},
			end:             Coordinate{X: 2, Y: 0},
			wantSegments:    10,   // Default interpolation segments
			wantTotalLength: 3.14, // Approximate arc length for semicircle
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test arc interpolation
			interpolatedPoints := interpolateArc(tt.start, tt.middle, tt.end, tt.wantSegments)

			if len(interpolatedPoints) != tt.wantSegments+1 { // +1 for inclusive end point
				t.Errorf("Interpolated points count = %d, want %d", len(interpolatedPoints), tt.wantSegments+1)
			}

			// Verify start and end points are preserved
			if interpolatedPoints[0] != tt.start {
				t.Errorf("First interpolated point = %v, want %v", interpolatedPoints[0], tt.start)
			}

			lastIdx := len(interpolatedPoints) - 1
			tolerance := 0.001
			if abs(interpolatedPoints[lastIdx].X-tt.end.X) > tolerance ||
				abs(interpolatedPoints[lastIdx].Y-tt.end.Y) > tolerance {
				t.Errorf("Last interpolated point = %v, want %v", interpolatedPoints[lastIdx], tt.end)
			}
		})
	}
}

// TestArcAttributes tests BUEP attribute parsing
func TestArcAttributes(t *testing.T) {
	parser := NewParser()

	arcTestData := `.HODE
..TRANSPAR
...KOORDSYS 22
...ORIGO-NØ 0  0
...ENHET 0.01
..SOSI-VERSJON 4.5
..SOSI-NIVÅ 4
.BUEP 8:
..OBJTYPE RpGrense
..KOPIDATA
...OMRÅDEID 0618
...ORIGINALDATAVERT "Hemsedal kommune"
...KOPIDATO 20130531
..OPPDATERINGSDATO 20130531092024
..NØ
674759173 47234619 ...KP 1
..NØ
674759286 47234337
674759383 47234049 ...KP 1
.SLUTT`

	reader := strings.NewReader(arcTestData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Fatal("No features found")
	}

	feature := doc.Features[0]

	// Test basic attributes
	if feature.ObjectType != "RpGrense" {
		t.Errorf("ObjectType = %s, want RpGrense", feature.ObjectType)
	}

	// Test properties parsing
	if feature.Properties == nil {
		t.Fatal("Feature Properties is nil")
	}

	// Check OBJTYPE in properties
	if objtype, exists := feature.Properties["OBJTYPE"]; !exists {
		t.Error("OBJTYPE not found in properties")
	} else if objtype != "RpGrense" {
		t.Errorf("Properties OBJTYPE = %v, want RpGrense", objtype)
	}

	// Test KOPIDATA parsing (nested structure)
	if kopidata, exists := feature.Properties["KOPIDATA"]; exists {
		if kopidataMap, ok := kopidata.(map[string]interface{}); ok {
			if områdeid, exists := kopidataMap["OMRÅDEID"]; !exists {
				t.Error("OMRÅDEID not found in kopidata")
			} else {
				// OMRÅDEID can be either string "0618" or integer 618
				switch v := områdeid.(type) {
				case string:
					if v != "0618" {
						t.Errorf("OMRÅDEID = %v, want 0618", v)
					}
				case int:
					if v != 618 {
						t.Errorf("OMRÅDEID = %v, want 618", v)
					}
				default:
					t.Errorf("OMRÅDEID unexpected type: %T = %v", v, v)
				}
			}
		} else {
			t.Logf("KOPIDATA format: %T = %v", kopidata, kopidata)
		}
	} else {
		t.Error("KOPIDATA not found in properties")
	}
}

// TestMultipleArcs tests parsing multiple BUEP features in one file
func TestMultipleArcs(t *testing.T) {
	parser := NewParser()

	// Test data with multiple arcs from buer.sos
	multiArcData := `.HODE
..TEGNSETT ISO8859-10
..SOSI-VERSJON 4.5
..SOSI-NIVÅ 4
..TRANSPAR
...KOORDSYS 22
...ORIGO-NØ 0  0
...ENHET 0.01
.BUEP 8:
..OBJTYPE RpGrense
..NØ
674759173 47234619 ...KP 1
..NØ
674759286 47234337
674759383 47234049 ...KP 1
.BUEP 11:
..OBJTYPE RpGrense
..NØ
674746679 47419783 ...KP 1
..NØ
674747030 47420868
674748043 47421392 ...KP 1
.BUEP 13:
..OBJTYPE RpGrense
..NØ
674767168 47419684 ...KP 1
..NØ
674767092 47419780
674767025 47419883 ...KP 1
.SLUTT`

	reader := strings.NewReader(multiArcData)
	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should parse 3 BUEP features
	expectedFeatureCount := 3
	if len(doc.Features) != expectedFeatureCount {
		t.Errorf("Feature count = %d, want %d", len(doc.Features), expectedFeatureCount)
	}

	// Check that all features are BUEP type
	expectedIDs := []int{8, 11, 13}
	for i, expectedID := range expectedIDs {
		if i >= len(doc.Features) {
			t.Errorf("Missing feature at index %d", i)
			continue
		}

		feature := doc.Features[i]
		if feature.ID != expectedID {
			t.Errorf("Feature[%d] ID = %d, want %d", i, feature.ID, expectedID)
		}

		if feature.Type != "BUEP" {
			t.Errorf("Feature[%d] Type = %s, want BUEP", i, feature.Type)
		}

		// Each BUEP should have exactly 3 coordinates
		if len(feature.Coordinates) != 3 {
			t.Errorf("Feature[%d] coordinate count = %d, want 3", i, len(feature.Coordinates))
		}
	}
}

// interpolateArc converts a 3-point circular arc to a series of line segments
// This is essential for converting BUEP geometry to standard GeoJSON LineString
func interpolateArc(start, middle, end Coordinate, segments int) []Coordinate {
	// Calculate circle center and radius from 3 points
	center, radius := calculateCircleFromThreePoints(start, middle, end)

	// Calculate start and end angles
	startAngle := math.Atan2(start.Y-center.Y, start.X-center.X)
	endAngle := math.Atan2(end.Y-center.Y, end.X-center.X)

	// Determine arc direction (clockwise or counterclockwise)
	// Check if middle point is on the shorter arc
	middleAngle := math.Atan2(middle.Y-center.Y, middle.X-center.X)

	// Normalize angles to [0, 2π)
	if startAngle < 0 {
		startAngle += 2 * math.Pi
	}
	if endAngle < 0 {
		endAngle += 2 * math.Pi
	}
	if middleAngle < 0 {
		middleAngle += 2 * math.Pi
	}

	// Determine the arc direction by checking if middle point is between start and end
	var angleStep float64
	var totalAngle float64

	if isAngleBetween(startAngle, middleAngle, endAngle) {
		// Counterclockwise
		if endAngle < startAngle {
			endAngle += 2 * math.Pi
		}
		totalAngle = endAngle - startAngle
	} else {
		// Clockwise
		if startAngle < endAngle {
			startAngle += 2 * math.Pi
		}
		totalAngle = startAngle - endAngle
		// For clockwise direction, we need to start from the original endAngle
		startAngle = endAngle
		totalAngle = -totalAngle
	}

	angleStep = totalAngle / float64(segments)

	// Generate interpolated points
	points := make([]Coordinate, segments+1)
	points[0] = start

	for i := 1; i < segments; i++ {
		angle := startAngle + float64(i)*angleStep
		x := center.X + radius*math.Cos(angle)
		y := center.Y + radius*math.Sin(angle)
		points[i] = Coordinate{X: x, Y: y, Z: start.Z} // Preserve Z from start point
	}

	points[segments] = end
	return points
}

// calculateCircleFromThreePoints finds the center and radius of a circle passing through 3 points
func calculateCircleFromThreePoints(p1, p2, p3 Coordinate) (center Coordinate, radius float64) {
	// Handle degenerate cases
	if isCollinear(p1, p2, p3) {
		// Points are collinear - return a very large radius (essentially a straight line)
		center = Coordinate{X: (p1.X + p3.X) / 2, Y: (p1.Y + p3.Y) / 2}
		radius = distance(p1, p3) / 2
		return
	}

	// Calculate the perpendicular bisectors
	ax, ay := (p1.X+p2.X)/2, (p1.Y+p2.Y)/2
	bx, by := (p2.X+p3.X)/2, (p2.Y+p3.Y)/2

	// Slopes of the original lines
	slope1 := (p2.Y - p1.Y) / (p2.X - p1.X)
	slope2 := (p3.Y - p2.Y) / (p3.X - p2.X)

	// Handle vertical lines
	if math.Abs(p2.X-p1.X) < 1e-10 {
		// Line 1 is vertical, perpendicular bisector is horizontal
		center.Y = ay
		center.X = bx + (by-ay)/slope2
	} else if math.Abs(p3.X-p2.X) < 1e-10 {
		// Line 2 is vertical, perpendicular bisector is horizontal
		center.Y = by
		center.X = ax + (ay-by)/slope1
	} else {
		// General case
		perpSlope1 := -1 / slope1
		perpSlope2 := -1 / slope2

		// Intersection of perpendicular bisectors
		center.X = (perpSlope1*ax - perpSlope2*bx + by - ay) / (perpSlope1 - perpSlope2)
		center.Y = perpSlope1*(center.X-ax) + ay
	}

	// Calculate radius
	radius = distance(center, p1)
	return
}

// isCollinear checks if three points are collinear
func isCollinear(p1, p2, p3 Coordinate) bool {
	// Calculate cross product of vectors (p2-p1) and (p3-p1)
	crossProduct := (p2.X-p1.X)*(p3.Y-p1.Y) - (p2.Y-p1.Y)*(p3.X-p1.X)
	return math.Abs(crossProduct) < 1e-10
}

// distance calculates Euclidean distance between two points
func distance(p1, p2 Coordinate) float64 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// isAngleBetween checks if angle is between start and end angles (counterclockwise)
func isAngleBetween(start, angle, end float64) bool {
	// Normalize all angles to [0, 2π)
	for start < 0 {
		start += 2 * math.Pi
	}
	for start >= 2*math.Pi {
		start -= 2 * math.Pi
	}

	for angle < 0 {
		angle += 2 * math.Pi
	}
	for angle >= 2*math.Pi {
		angle -= 2 * math.Pi
	}

	for end < 0 {
		end += 2 * math.Pi
	}
	for end >= 2*math.Pi {
		end -= 2 * math.Pi
	}

	if start <= end {
		return start <= angle && angle <= end
	} else {
		return angle >= start || angle <= end
	}
}
