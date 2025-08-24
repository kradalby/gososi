package sosi

import (
	"math"
	"testing"
)

func TestCalculateBoundingBox(t *testing.T) {
	tests := []struct {
		name        string
		coordinates []Coordinate
		expected    BoundingBox
	}{
		{
			name: "single coordinate",
			coordinates: []Coordinate{
				{Y: 63.4856654467292, X: 10.921292458550036, Z: 0.0},
			},
			expected: BoundingBox{
				MinLat: 63.4856654467292,
				MinLon: 10.921292458550036,
				MaxLat: 63.4856654467292,
				MaxLon: 10.921292458550036,
			},
		},
		{
			name: "multiple coordinates",
			coordinates: []Coordinate{
				{Y: 63.4856654467292, X: 10.921292458550036, Z: 0.0},
				{Y: 63.486702373293724, X: 10.920455609197402, Z: 0.0},
				{Y: 63.485960004751576, X: 10.919962082656106, Z: 0.0},
			},
			expected: BoundingBox{
				MinLat: 63.4856654467292,   // Actually the lowest lat value
				MinLon: 10.919962082656106, // Lowest lon value
				MaxLat: 63.486702373293724, // Highest lat value
				MaxLon: 10.921292458550036, // Highest lon value
			},
		},
		{
			name:        "empty coordinates",
			coordinates: []Coordinate{},
			expected:    BoundingBox{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBoundingBox(tt.coordinates)

			if len(tt.coordinates) == 0 {
				// For empty coordinates, just check that we get an empty bbox
				if result.MinLat != 0 || result.MinLon != 0 || result.MaxLat != 0 || result.MaxLon != 0 {
					t.Errorf("CalculateBoundingBox() for empty coords = %v, want empty bbox", result)
				}
				return
			}

			if result.MinLat != tt.expected.MinLat {
				t.Errorf("CalculateBoundingBox() MinLat = %v, want %v", result.MinLat, tt.expected.MinLat)
			}
			if result.MinLon != tt.expected.MinLon {
				t.Errorf("CalculateBoundingBox() MinLon = %v, want %v", result.MinLon, tt.expected.MinLon)
			}
			if result.MaxLat != tt.expected.MaxLat {
				t.Errorf("CalculateBoundingBox() MaxLat = %v, want %v", result.MaxLat, tt.expected.MaxLat)
			}
			if result.MaxLon != tt.expected.MaxLon {
				t.Errorf("CalculateBoundingBox() MaxLon = %v, want %v", result.MaxLon, tt.expected.MaxLon)
			}
		})
	}
}

func TestUpdateBoundingBox(t *testing.T) {
	// Start with initial bounding box
	bbox := BoundingBox{
		MinLat: math.Inf(1),
		MinLon: math.Inf(1),
		MaxLat: math.Inf(-1),
		MaxLon: math.Inf(-1),
	}

	// First update
	coords1 := []Coordinate{
		{Y: 63.4856654467292, X: 10.921292458550036, Z: 0.0},
	}

	UpdateBoundingBox(&bbox, coords1)

	if bbox.MinLat != 63.4856654467292 || bbox.MaxLat != 63.4856654467292 {
		t.Errorf("First update failed: got MinLat=%v, MaxLat=%v", bbox.MinLat, bbox.MaxLat)
	}

	// Second update with expanding coordinates
	coords2 := []Coordinate{
		{Y: 63.486702373293724, X: 10.920455609197402, Z: 0.0}, // Higher lat, lower lon
		{Y: 63.485960004751576, X: 10.919962082656106, Z: 0.0}, // Lower lat, even lower lon
	}

	UpdateBoundingBox(&bbox, coords2)

	expectedMinLat := 63.4856654467292
	expectedMaxLat := 63.486702373293724
	expectedMinLon := 10.919962082656106
	expectedMaxLon := 10.921292458550036

	if bbox.MinLat != expectedMinLat {
		t.Errorf("UpdateBoundingBox() MinLat = %v, want %v", bbox.MinLat, expectedMinLat)
	}
	if bbox.MaxLat != expectedMaxLat {
		t.Errorf("UpdateBoundingBox() MaxLat = %v, want %v", bbox.MaxLat, expectedMaxLat)
	}
	if bbox.MinLon != expectedMinLon {
		t.Errorf("UpdateBoundingBox() MinLon = %v, want %v", bbox.MinLon, expectedMinLon)
	}
	if bbox.MaxLon != expectedMaxLon {
		t.Errorf("UpdateBoundingBox() MaxLon = %v, want %v", bbox.MaxLon, expectedMaxLon)
	}
}
