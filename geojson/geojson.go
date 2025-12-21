// Package geojson provides GeoJSON geometry types with 3D coordinate support.
//
// This package is based on github.com/paulmach/orb/geojson but with:
// - Ergonomic struct-based Point type with named Lon, Lat, Depth fields
// - Full 3D coordinate support (depth/altitude) on all geometry types
// - Database integration via Null* types with sql.Scanner/driver.Valuer
// - JSON-only serialization using json/v2 (no BSON)
package geojson

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/go-json-experiment/json"
)

var (
	errUnexpectedPointType       = errors.New("unexpected geometry type, expected Point")
	errUnexpectedLineStringType  = errors.New("unexpected geometry type, expected LineString")
	errInsufficientCoordinates   = errors.New("point requires at least 2 coordinates")
	errUnsupportedScanType       = errors.New("cannot scan type into NullPoint")
	errUnsupportedLineStringScan = errors.New("cannot scan type into NullLineString")
)

// Point represents a 3D geographic point.
// Coordinates follow GeoJSON order: longitude, latitude, altitude/depth.
type Point struct {
	Lon   float64
	Lat   float64
	Depth float64 // Depth/altitude in meters (0 if not set)
}

// pointJSON is the GeoJSON representation.
type pointJSON struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// MarshalJSON implements json.Marshaler.
// Produces GeoJSON with optional depth coordinate (omitted if zero).
func (p Point) MarshalJSON() ([]byte, error) {
	coords := []float64{p.Lon, p.Lat}
	if p.Depth != 0 {
		coords = append(coords, p.Depth)
	}

	return json.Marshal(pointJSON{
		Type:        "Point",
		Coordinates: coords,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Point) UnmarshalJSON(data []byte) error {
	var pj pointJSON

	err := json.Unmarshal(data, &pj)
	if err != nil {
		return err
	}

	if pj.Type != "Point" {
		return fmt.Errorf("%w: got %q", errUnexpectedPointType, pj.Type)
	}

	if len(pj.Coordinates) < 2 {
		return fmt.Errorf("%w: got %d", errInsufficientCoordinates, len(pj.Coordinates))
	}

	p.Lon = pj.Coordinates[0]
	p.Lat = pj.Coordinates[1]

	if len(pj.Coordinates) >= 3 {
		p.Depth = pj.Coordinates[2]
	} else {
		p.Depth = 0
	}

	return nil
}

// NullPoint represents a nullable Point for database storage.
// Stored as JSON string in the database.
//
//nolint:recvcheck // intentionally mixed receivers for interface compliance
type NullPoint struct {
	Point Point
	Valid bool
}

// Scan implements sql.Scanner. Reads JSON string from database.
func (np *NullPoint) Scan(value any) error {
	if value == nil {
		np.Point = Point{}
		np.Valid = false

		return nil
	}

	var data []byte

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("%w: %T", errUnsupportedScanType, value)
	}

	if len(data) == 0 {
		np.Point = Point{}
		np.Valid = false

		return nil
	}

	err := json.Unmarshal(data, &np.Point)
	if err != nil {
		return fmt.Errorf("unmarshaling point: %w", err)
	}

	np.Valid = true

	return nil
}

// Value implements driver.Valuer. Writes JSON string to database.
// Uses value receiver to work with non-pointer struct fields.
func (np NullPoint) Value() (driver.Value, error) {
	if !np.Valid {
		return nil, nil //nolint:nilnil // driver.Valuer requires nil,nil for NULL
	}

	data, err := json.Marshal(np.Point)
	if err != nil {
		return nil, fmt.Errorf("marshaling point: %w", err)
	}

	return string(data), nil
}

// Equal compares two NullPoint values.
func (np NullPoint) Equal(other NullPoint) bool {
	if !np.Valid && !other.Valid {
		return true
	}

	if np.Valid != other.Valid {
		return false
	}

	return np.Point.Lon == other.Point.Lon &&
		np.Point.Lat == other.Point.Lat &&
		np.Point.Depth == other.Point.Depth
}

// LineString represents a series of connected 3D points.
// Stored as GeoJSON LineString with optional altitude/depth per point.
type LineString []Point

// lineStringJSON is the GeoJSON representation.
type lineStringJSON struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

// MarshalJSON implements json.Marshaler.
func (ls LineString) MarshalJSON() ([]byte, error) {
	coords := make([][]float64, len(ls))

	for i, p := range ls {
		if p.Depth != 0 {
			coords[i] = []float64{p.Lon, p.Lat, p.Depth}
		} else {
			coords[i] = []float64{p.Lon, p.Lat}
		}
	}

	return json.Marshal(lineStringJSON{
		Type:        "LineString",
		Coordinates: coords,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (ls *LineString) UnmarshalJSON(data []byte) error {
	var lsj lineStringJSON

	err := json.Unmarshal(data, &lsj)
	if err != nil {
		return err
	}

	if lsj.Type != "LineString" {
		return fmt.Errorf("%w: got %q", errUnexpectedLineStringType, lsj.Type)
	}

	*ls = make(LineString, len(lsj.Coordinates))

	for i, coord := range lsj.Coordinates {
		if len(coord) < 2 {
			return fmt.Errorf("%w: got %d", errInsufficientCoordinates, len(coord))
		}

		(*ls)[i] = Point{
			Lon: coord[0],
			Lat: coord[1],
		}

		if len(coord) >= 3 {
			(*ls)[i].Depth = coord[2]
		}
	}

	return nil
}

// NullLineString represents a nullable LineString for database storage.
//
//nolint:recvcheck // intentionally mixed receivers for interface compliance
type NullLineString struct {
	LineString LineString
	Valid      bool
}

// Scan implements sql.Scanner.
func (nls *NullLineString) Scan(value any) error {
	if value == nil {
		nls.LineString = nil
		nls.Valid = false

		return nil
	}

	var data []byte

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("%w: %T", errUnsupportedLineStringScan, value)
	}

	if len(data) == 0 {
		nls.LineString = nil
		nls.Valid = false

		return nil
	}

	err := json.Unmarshal(data, &nls.LineString)
	if err != nil {
		return fmt.Errorf("unmarshaling linestring: %w", err)
	}

	nls.Valid = true

	return nil
}

// Value implements driver.Valuer.
func (nls NullLineString) Value() (driver.Value, error) {
	if !nls.Valid {
		return nil, nil //nolint:nilnil // driver.Valuer requires nil,nil for NULL
	}

	data, err := json.Marshal(nls.LineString)
	if err != nil {
		return nil, fmt.Errorf("marshaling linestring: %w", err)
	}

	return string(data), nil
}

// Equal compares two NullLineString values.
func (nls NullLineString) Equal(other NullLineString) bool {
	if !nls.Valid && !other.Valid {
		return true
	}

	if nls.Valid != other.Valid {
		return false
	}

	if len(nls.LineString) != len(other.LineString) {
		return false
	}

	for i := range nls.LineString {
		if nls.LineString[i].Lon != other.LineString[i].Lon ||
			nls.LineString[i].Lat != other.LineString[i].Lat ||
			nls.LineString[i].Depth != other.LineString[i].Depth {
			return false
		}
	}

	return true
}
