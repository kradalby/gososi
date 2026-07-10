// Package proj provides pure-Go coordinate projection between geographic
// (WGS84 longitude/latitude, in degrees) and projected transverse-Mercator /
// UTM coordinates (easting/northing, in metres).
//
// It is dependency-free (standard library only) so it can be reused across
// projects. The transverse-Mercator implementation uses the Karney-Krüger
// n-series to sixth order, matching GeographicLib / PROJ etmerc / Kartverket to
// sub-micrometre accuracy over each zone's valid extent.
//
// Coordinate order follows GeoJSON: longitude first, then latitude.
package proj

import "errors"

var (
	// ErrUnknownKOORDSYS is returned by FromKOORDSYS for a code that is not a
	// recognised SOSI coordinate-system code.
	ErrUnknownKOORDSYS = errors.New("proj: unknown KOORDSYS code")

	// ErrUnsupportedDatum is returned by FromKOORDSYS for a recognised code
	// whose datum needs a Helmert transform that is not yet implemented
	// (NGO1948 codes 1-9, ED50 codes 31-36).
	ErrUnsupportedDatum = errors.New("proj: KOORDSYS datum not supported (needs Helmert transform)")
)

// CRS is a coordinate reference system that can project geographic coordinates
// to a plane and back.
//
// Forward maps geographic (lon, lat) in degrees to projected (easting,
// northing) in metres. Inverse is its inverse. For every point within the
// system's valid domain, Inverse(Forward(lon, lat)) reproduces (lon, lat) to
// sub-millimetre accuracy.
type CRS interface {
	Forward(lon, lat float64) (easting, northing float64)
	Inverse(easting, northing float64) (lon, lat float64)
}
