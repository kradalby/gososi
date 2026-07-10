package geojson

import "math"

// earthRadiusM is the mean Earth radius in meters (Veness convention).
const earthRadiusM = 6_371_000

// DistanceHaversine returns the great-circle distance to other in meters,
// computed with the haversine formula on a sphere of radius 6 371 000 m.
// It is 2D: Depth is ignored. Spherical distance differs from the WGS84
// geodesic by up to ~0.5%.
func (p Point) DistanceHaversine(other Point) float64 {
	lat1 := p.Lat * math.Pi / 180
	lat2 := other.Lat * math.Pi / 180
	dLat := lat2 - lat1
	dLon := (other.Lon - p.Lon) * math.Pi / 180

	sinLat := math.Sin(dLat / 2)
	sinLon := math.Sin(dLon / 2)
	a := sinLat*sinLat + math.Cos(lat1)*math.Cos(lat2)*sinLon*sinLon

	return 2 * earthRadiusM * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// LengthHaversine returns the length of the line in meters as the sum of
// the haversine distances between consecutive points. Lines with fewer
// than two points have length 0. It is 2D: Depth is ignored.
func (ls LineString) LengthHaversine() float64 {
	var sum float64
	for i := 1; i < len(ls); i++ {
		sum += ls[i-1].DistanceHaversine(ls[i])
	}

	return sum
}
