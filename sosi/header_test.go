package sosi

import (
	"strings"
	"testing"
)

func TestGenerateHeader(t *testing.T) {
	config := DefaultConfig()
	bbox := BoundingBox{
		MinLat: 63.485960004751576,
		MinLon: 10.919962082656106,
		MaxLat: 63.486702373293724,
		MaxLon: 10.921292458550036,
	}

	result := GenerateHeader(config, bbox)

	expectedLines := []string{
		".HODE 0:",
		"..TEGNSETT UTF-8",
		"..TRANSPAR",
		"...KOORDSYS 84",
		"...ORIGO-NØ 0 0",
		"...ENHET 0.0000001",
		"...ENHET-H 0.001",
		"...ENHET-D 0.001",
		"..PRODUSENT \"GeoJSONtoSOSI\"",
		"..SOSI-VERSJON 4.0",
		"..SOSI-NIVÅ 2",
		"..OMRÅDE",
		"...MIN-NØ 63 10",
		"...MAX-NØ 64 11",
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")

	if len(lines) != len(expectedLines) {
		t.Fatalf("Expected %d lines, got %d", len(expectedLines), len(lines))
	}

	for i, expected := range expectedLines {
		if lines[i] != expected {
			t.Errorf("Line %d: got %q, want %q", i+1, lines[i], expected)
		}
	}
}

func TestGenerateHeader_CustomConfig(t *testing.T) {
	config := SOSIConfig{
		Producer:         "CustomProducer",
		AltitudeAccuracy: 3,
		LatLongAccuracy:  7,
		CoordSystem:      25832,
		Version:          "4.5",
		Level:            3,
	}

	bbox := BoundingBox{
		MinLat: 60.0,
		MinLon: 5.0,
		MaxLat: 61.0,
		MaxLon: 6.0,
	}

	result := GenerateHeader(config, bbox)

	if !strings.Contains(result, "...KOORDSYS 25832") {
		t.Errorf("Expected custom coordinate system 25832, got: %s", result)
	}

	if !strings.Contains(result, "..PRODUSENT \"CustomProducer\"") {
		t.Errorf("Expected custom producer, got: %s", result)
	}

	if !strings.Contains(result, "..SOSI-VERSJON 4.5") {
		t.Errorf("Expected custom version 4.5, got: %s", result)
	}

	if !strings.Contains(result, "..SOSI-NIVÅ 3") {
		t.Errorf("Expected custom level 3, got: %s", result)
	}

	if !strings.Contains(result, "...MIN-NØ 60 5") {
		t.Errorf("Expected MIN-NØ 60 5, got: %s", result)
	}

	if !strings.Contains(result, "...MAX-NØ 61 6") {
		t.Errorf("Expected MAX-NØ 61 6, got: %s", result)
	}
}
