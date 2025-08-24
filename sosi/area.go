package sosi

import "math"

// Coordinate represents a single coordinate point with latitude, longitude, and optional altitude
type Coordinate struct {
	X, Y, Z      float64 // X=Longitude, Y=Latitude, Z=Altitude
	TiePointCode int     // Knutepunktkode (tie point code) for SOSI parsing, 0 if not a tie point
}

// CalculateBoundingBox computes the bounding box from a collection of coordinates
func CalculateBoundingBox(coordinates []Coordinate) BoundingBox {
	if len(coordinates) == 0 {
		return BoundingBox{}
	}

	bbox := BoundingBox{
		MinLat: math.Inf(1),  // Positive infinity
		MinLon: math.Inf(1),  // Positive infinity
		MaxLat: math.Inf(-1), // Negative infinity
		MaxLon: math.Inf(-1), // Negative infinity
	}

	for _, coord := range coordinates {
		// Update minimums
		if coord.Y < bbox.MinLat {
			bbox.MinLat = coord.Y
		}
		if coord.X < bbox.MinLon {
			bbox.MinLon = coord.X
		}

		// Update maximums
		if coord.Y > bbox.MaxLat {
			bbox.MaxLat = coord.Y
		}
		if coord.X > bbox.MaxLon {
			bbox.MaxLon = coord.X
		}
	}

	return bbox
}

// UpdateBoundingBox extends an existing bounding box with new coordinates
func UpdateBoundingBox(bbox *BoundingBox, coordinates []Coordinate) {
	for _, coord := range coordinates {
		// Update minimums
		if coord.Y < bbox.MinLat {
			bbox.MinLat = coord.Y
		}
		if coord.X < bbox.MinLon {
			bbox.MinLon = coord.X
		}

		// Update maximums
		if coord.Y > bbox.MaxLat {
			bbox.MaxLat = coord.Y
		}
		if coord.X > bbox.MaxLon {
			bbox.MaxLon = coord.X
		}
	}
}
