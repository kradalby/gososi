package sosi

import (
	"strings"
	"testing"
)

// TestParser_Parse tests the main parsing functionality
// Ports JavaScript parser-test.js functionality using Go table-driven test pattern
func TestParser_Parse(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name           string
		filename       string
		wantFeatures   int
		wantHeader     SOSIHeader
		wantFeatureIDs []int
		wantError      bool
	}{
		{
			name:         "basic SOSI file parsing",
			filename:     "testfile1.sos",
			wantFeatures: 5,
			wantHeader: SOSIHeader{
				CharacterSet:     "UTF-8",
				Producer:         "SØRKART A/S",
				Owner:            "Statens kartverk",
				ObjectCatalog:    "Eksempel 4.5",
				Version:          "4.5",
				Level:            5,
				CoordSystem:      5,
				Unit:             0.01,
				VerificationDate: "19890623",
				Origo: Coordinate{
					X: 10000,  // Ø (longitude)
					Y: 100000, // N (latitude)
					Z: 0,
				},
				Area: BoundingBox{
					MinLat: 100000,
					MinLon: 10000,
					MaxLat: 102400,
					MaxLon: 13200,
				},
			},
			wantFeatureIDs: []int{1, 250, 223, 312, 298},
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test file
			sosiData := getTestData(t, tt.filename)
			reader := strings.NewReader(sosiData)

			// Parse SOSI data
			doc, err := parser.Parse(reader)

			if tt.wantError {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Test feature count
			if len(doc.Features) != tt.wantFeatures {
				t.Errorf("Parse() features count = %d, want %d", len(doc.Features), tt.wantFeatures)
			}

			// Test header fields
			testHeader(t, doc.Header, tt.wantHeader)

			// Test feature IDs
			testFeatureIDs(t, doc.Features, tt.wantFeatureIDs)
		})
	}
}

// TestParser_HeaderParsing tests header parsing specifically
// Equivalent to JavaScript "should read header" test
func TestParser_HeaderParsing(t *testing.T) {
	parser := NewParser()
	sosiData := getTestData(t, "testfile1.sos")
	reader := strings.NewReader(sosiData)

	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	header := doc.Header

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"Producer", header.Producer, "SØRKART A/S"},
		{"Owner", header.Owner, "Statens kartverk"},
		{"ObjectCatalog", header.ObjectCatalog, "Eksempel 4.5"},
		{"Version", header.Version, "4.5"},
		{"Level", header.Level, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("Header.%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

// TestParser_QualityParsing tests quality attribute parsing
// Equivalent to JavaScript "should get kvalitet" test
func TestParser_QualityParsing(t *testing.T) {
	parser := NewParser()
	sosiData := getTestData(t, "testfile1.sos")
	reader := strings.NewReader(sosiData)

	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Check header quality - should now work properly
	if len(doc.Header.Quality) > 0 {
		t.Logf("Quality data: %+v", doc.Header.Quality)

		// Verify the expected values from testfile1.sos: "..KVALITET 11 300"
		if målemetode := doc.Header.Quality["målemetode"]; målemetode != 11 {
			t.Errorf("Expected målemetode=11, got %v", målemetode)
		}
		if nøyaktighet := doc.Header.Quality["nøyaktighet"]; nøyaktighet != 300 {
			t.Errorf("Expected nøyaktighet=300, got %v", nøyaktighet)
		}

		t.Log("✅ Header KVALITET parsing implemented and working")
	} else {
		t.Error("Header KVALITET parsing failed - Quality is nil or empty")
	}
}

// TestParser_BoundsParsing tests bounding box parsing
// Equivalent to JavaScript "should get bounds" test
func TestParser_BoundsParsing(t *testing.T) {
	parser := NewParser()
	sosiData := getTestData(t, "testfile1.sos")
	reader := strings.NewReader(sosiData)

	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := BoundingBox{
		MinLat: 100000,
		MinLon: 10000,
		MaxLat: 102400,
		MaxLon: 13200,
	}

	if doc.Header.Area != expected {
		t.Errorf("Header.Area = %+v, want %+v", doc.Header.Area, expected)
	}
}

// TestParser_OrigoParsing tests origin point parsing
// Equivalent to JavaScript "should get origo" test
func TestParser_OrigoParsing(t *testing.T) {
	parser := NewParser()
	sosiData := getTestData(t, "testfile1.sos")
	reader := strings.NewReader(sosiData)

	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := Coordinate{
		X: 10000,  // Ø (longitude)
		Y: 100000, // N (latitude)
		Z: 0,
	}

	if doc.Header.Origo != expected {
		t.Errorf("Header.Origo = %+v, want %+v", doc.Header.Origo, expected)
	}
}

// TestParser_UnitParsing tests coordinate unit parsing
// Equivalent to JavaScript "should get enhet" test
func TestParser_UnitParsing(t *testing.T) {
	parser := NewParser()
	sosiData := getTestData(t, "testfile1.sos")
	reader := strings.NewReader(sosiData)

	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := 0.01
	if doc.Header.Unit != expected {
		t.Errorf("Header.Unit = %v, want %v", doc.Header.Unit, expected)
	}
}

// TestParser_SRIDParsing tests SRID extraction from coordinate system
// Equivalent to JavaScript "should get srid" test
func TestParser_SRIDParsing(t *testing.T) {
	parser := NewParser()
	sosiData := getTestData(t, "testfile1.sos")
	reader := strings.NewReader(sosiData)

	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// KOORDSYS 5 should map to EPSG:27395
	expected := 5
	if doc.Header.CoordSystem != expected {
		t.Errorf("Header.CoordSystem = %v, want %v", doc.Header.CoordSystem, expected)
	}

	// Test SRID lookup
	if coordSys, exists := parser.coordSystems[doc.Header.CoordSystem]; exists {
		expectedSRID := "EPSG:27395"
		if coordSys.SRID != expectedSRID {
			t.Errorf("CoordSystem SRID = %v, want %v", coordSys.SRID, expectedSRID)
		}
	} else {
		t.Errorf("Coordinate system %d not found in lookup table", doc.Header.CoordSystem)
	}
}

// TestParser_FeatureParsing tests individual feature parsing
func TestParser_FeatureParsing(t *testing.T) {
	parser := NewParser()
	sosiData := getTestData(t, "testfile1.sos")
	reader := strings.NewReader(sosiData)

	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(doc.Features) == 0 {
		t.Fatal("No features parsed")
	}

	// Test first feature (PUNKT 1)
	feature := doc.Features[0]

	expectedFeature := struct {
		ID         int
		Type       string
		ObjectType string
		HasCoords  bool
	}{
		ID:         1,
		Type:       "PUNKT",
		ObjectType: "Fastmerke",
		HasCoords:  true,
	}

	if feature.ID != expectedFeature.ID {
		t.Errorf("Feature.ID = %d, want %d", feature.ID, expectedFeature.ID)
	}

	if feature.Type != expectedFeature.Type {
		t.Errorf("Feature.Type = %s, want %s", feature.Type, expectedFeature.Type)
	}

	if feature.ObjectType != expectedFeature.ObjectType {
		t.Errorf("Feature.ObjectType = %s, want %s", feature.ObjectType, expectedFeature.ObjectType)
	}

	if len(feature.Coordinates) == 0 && expectedFeature.HasCoords {
		t.Error("Feature.Coordinates is empty, expected coordinates")
	}
}

// TestParser_GeometryTypes tests parsing of different geometry types
func TestParser_GeometryTypes(t *testing.T) {
	parser := NewParser()
	sosiData := getTestData(t, "testfile1.sos")
	reader := strings.NewReader(sosiData)

	doc, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Expected geometry types from testfile1.sos
	expectedTypes := map[int]string{
		1:   "PUNKT", // Point
		250: "KURVE", // LineString
		223: "KURVE", // LineString with height
		312: "BUEP",  // Arc
		298: "TEKST", // Text (treated as Point)
	}

	featureMap := make(map[int]SOSIFeature)
	for _, feature := range doc.Features {
		featureMap[feature.ID] = feature
	}

	for id, expectedType := range expectedTypes {
		if feature, exists := featureMap[id]; exists {
			if feature.Type != expectedType {
				t.Errorf("Feature %d type = %s, want %s", id, feature.Type, expectedType)
			}
		} else {
			t.Errorf("Feature %d not found", id)
		}
	}
}

// TestParser_LineCleaning tests SOSI line cleaning functionality
func TestParser_LineCleaning(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "remove comments and empty lines",
			input: []string{
				".HODE                    !SOSI-filas hode.",
				"..TRANSPAR",
				"! This is a comment",
				"...KOORDSYS 5",
				"",
				"   ",
				"...ORIGO-NØ 100000 10000 !inline comment",
			},
			expected: []string{
				".HODE",
				"..TRANSPAR",
				"...KOORDSYS 5",
				"...ORIGO-NØ 100000 10000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.splitOnNewline(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("splitOnNewline() length = %d, want %d", len(result), len(tt.expected))
				t.Errorf("Got: %v", result)
				t.Errorf("Want: %v", tt.expected)
				return
			}

			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("splitOnNewline()[%d] = %q, want %q", i, line, tt.expected[i])
				}
			}
		})
	}
}

// Helper functions

// testHeader compares parsed header with expected values
func testHeader(t *testing.T, got, want SOSIHeader) {
	t.Helper()

	if got.Producer != want.Producer {
		t.Errorf("Header.Producer = %q, want %q", got.Producer, want.Producer)
	}

	if got.Owner != want.Owner {
		t.Errorf("Header.Owner = %q, want %q", got.Owner, want.Owner)
	}

	if got.ObjectCatalog != want.ObjectCatalog {
		t.Errorf("Header.ObjectCatalog = %q, want %q", got.ObjectCatalog, want.ObjectCatalog)
	}

	if got.Version != want.Version {
		t.Errorf("Header.Version = %q, want %q", got.Version, want.Version)
	}

	if got.Level != want.Level {
		t.Errorf("Header.Level = %d, want %d", got.Level, want.Level)
	}

	if got.CoordSystem != want.CoordSystem {
		t.Errorf("Header.CoordSystem = %d, want %d", got.CoordSystem, want.CoordSystem)
	}

	if got.Unit != want.Unit {
		t.Errorf("Header.Unit = %f, want %f", got.Unit, want.Unit)
	}

	if got.Origo != want.Origo {
		t.Errorf("Header.Origo = %+v, want %+v", got.Origo, want.Origo)
	}

	if got.Area != want.Area {
		t.Errorf("Header.Area = %+v, want %+v", got.Area, want.Area)
	}
}

// testFeatureIDs checks that all expected feature IDs are present
func testFeatureIDs(t *testing.T, features []SOSIFeature, wantIDs []int) {
	t.Helper()

	gotIDs := make([]int, len(features))
	for i, feature := range features {
		gotIDs[i] = feature.ID
	}

	if len(gotIDs) != len(wantIDs) {
		t.Errorf("Feature count = %d, want %d", len(gotIDs), len(wantIDs))
		t.Errorf("Got IDs: %v", gotIDs)
		t.Errorf("Want IDs: %v", wantIDs)
		return
	}

	// Create maps for comparison (order doesn't matter)
	gotMap := make(map[int]bool)
	wantMap := make(map[int]bool)

	for _, id := range gotIDs {
		gotMap[id] = true
	}

	for _, id := range wantIDs {
		wantMap[id] = true
	}

	// Check all expected IDs are present
	for id := range wantMap {
		if !gotMap[id] {
			t.Errorf("Missing feature ID: %d", id)
		}
	}

	// Check no unexpected IDs are present
	for id := range gotMap {
		if !wantMap[id] {
			t.Errorf("Unexpected feature ID: %d", id)
		}
	}
}

// getTestData reads test data from testdata directory
func getTestData(t *testing.T, filename string) string {
	t.Helper()

	// For now, return the known testfile1.sos content directly
	// In a real implementation, this would read from testdata/filename
	if filename == "testfile1.sos" {
		return `.HODE                                           !SOSI-filas hode.
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
!
.KURVE 223:                                     !Geometritype KURVE.
..OBJTYPE ElvBekk
..VANNBR 1
..KVALITET 51 200
..NØH
23456 2345 123    ...KP 1                       !den ene enden er knutepunkt.
..NØ
23460 2360 123
!
.BUEP 312:                                      !Geometritype BUEP.
..OBJTYPE EiendomsGrense
..NØ
23470 2355
..NØ
23456 2345
23480 2367
!
.TEKST 298:                                     !Kartografisk tekstelement TEKST
..STRENG "Valbjørg-vatnet"
..NØ
23467 2350
23400 2400                                      !Teksten definert med STRENG
                                                !Skal skrives ved punkt 2.
!
.SLUTT                                          !Slutt på data`
	}

	t.Fatalf("Test data file not found: %s", filename)
	return ""
}
