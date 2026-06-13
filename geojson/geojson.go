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
	"github.com/go-json-experiment/json/jsontext"
)

var (
	errUnexpectedPointType           = errors.New("unexpected geometry type, expected Point")
	errUnexpectedLineStringType      = errors.New("unexpected geometry type, expected LineString")
	errUnexpectedPolygonType         = errors.New("unexpected geometry type, expected Polygon")
	errUnexpectedMultiPointType      = errors.New("unexpected geometry type, expected MultiPoint")
	errUnexpectedMultiLineStringType = errors.New("unexpected geometry type, expected MultiLineString")
	errUnexpectedMultiPolygonType    = errors.New("unexpected geometry type, expected MultiPolygon")
	errUnknownGeometryType           = errors.New("unknown geometry type")
	errInsufficientCoordinates       = errors.New("point requires at least 2 coordinates")
	errUnsupportedScanType           = errors.New("cannot scan type into NullPoint")
	errUnsupportedLineStringScan     = errors.New("cannot scan type into NullLineString")
	errUnsupportedPolygonScan        = errors.New("cannot scan type into NullPolygon")
)

// Geometry is the interface implemented by all GeoJSON geometry types.
type Geometry interface {
	GeoJSONType() string
}

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

// GeoJSONType returns the GeoJSON type string.
func (p Point) GeoJSONType() string {
	return "Point"
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

// GeoJSONType returns the GeoJSON type string.
func (ls LineString) GeoJSONType() string {
	return "LineString"
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

// Ring represents a closed linear ring of 3D points.
// Used as part of Polygon (exterior ring and holes).
// The ring should be closed (first point equals last point).
type Ring []Point

// IsClosed returns true if the ring is closed (first point equals last point).
func (r Ring) IsClosed() bool {
	if len(r) < 2 {
		return false
	}

	first := r[0]
	last := r[len(r)-1]

	return first.Lon == last.Lon && first.Lat == last.Lat && first.Depth == last.Depth
}

// Close returns a new ring with the first point appended if not already closed.
// If the ring is already closed, returns the original ring.
func (r Ring) Close() Ring {
	if r.IsClosed() {
		return r
	}

	if len(r) == 0 {
		return r
	}

	closed := make(Ring, len(r)+1)
	copy(closed, r)
	closed[len(r)] = r[0]

	return closed
}

// GeoJSONType returns the GeoJSON type string.
// Note: Ring is not a standard GeoJSON type, but is used internally for Polygon construction.
func (r Ring) GeoJSONType() string {
	return "Ring"
}

// Equal compares two rings for equality.
func (r Ring) Equal(other Ring) bool {
	if len(r) != len(other) {
		return false
	}

	for i := range r {
		if r[i].Lon != other[i].Lon ||
			r[i].Lat != other[i].Lat ||
			r[i].Depth != other[i].Depth {
			return false
		}
	}

	return true
}

// Polygon represents a polygon with optional holes.
// The first ring is the exterior boundary, subsequent rings are holes.
// All rings should be closed (first point equals last point).
type Polygon []Ring

// polygonJSON is the GeoJSON representation.
type polygonJSON struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// MarshalJSON implements json.Marshaler.
func (p Polygon) MarshalJSON() ([]byte, error) {
	coords := make([][][]float64, len(p))

	for i, ring := range p {
		// Auto-close rings during marshaling
		closedRing := ring.Close()
		coords[i] = make([][]float64, len(closedRing))

		for j, pt := range closedRing {
			if pt.Depth != 0 {
				coords[i][j] = []float64{pt.Lon, pt.Lat, pt.Depth}
			} else {
				coords[i][j] = []float64{pt.Lon, pt.Lat}
			}
		}
	}

	return json.Marshal(polygonJSON{
		Type:        "Polygon",
		Coordinates: coords,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Polygon) UnmarshalJSON(data []byte) error {
	var pj polygonJSON

	err := json.Unmarshal(data, &pj)
	if err != nil {
		return err
	}

	if pj.Type != "Polygon" {
		return fmt.Errorf("%w: got %q", errUnexpectedPolygonType, pj.Type)
	}

	*p = make(Polygon, len(pj.Coordinates))

	for i, ringCoords := range pj.Coordinates {
		(*p)[i] = make(Ring, len(ringCoords))

		for j, coord := range ringCoords {
			if len(coord) < 2 {
				return fmt.Errorf("%w: got %d", errInsufficientCoordinates, len(coord))
			}

			(*p)[i][j] = Point{
				Lon: coord[0],
				Lat: coord[1],
			}

			if len(coord) >= 3 {
				(*p)[i][j].Depth = coord[2]
			}
		}
	}

	return nil
}

// GeoJSONType returns the GeoJSON type string.
func (p Polygon) GeoJSONType() string {
	return "Polygon"
}

// NullPolygon represents a nullable Polygon for database storage.
//
//nolint:recvcheck // intentionally mixed receivers for interface compliance
type NullPolygon struct {
	Polygon Polygon
	Valid   bool
}

// Scan implements sql.Scanner.
func (np *NullPolygon) Scan(value any) error {
	if value == nil {
		np.Polygon = nil
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
		return fmt.Errorf("%w: %T", errUnsupportedPolygonScan, value)
	}

	if len(data) == 0 {
		np.Polygon = nil
		np.Valid = false

		return nil
	}

	err := json.Unmarshal(data, &np.Polygon)
	if err != nil {
		return fmt.Errorf("unmarshaling polygon: %w", err)
	}

	np.Valid = true

	return nil
}

// Value implements driver.Valuer.
func (np NullPolygon) Value() (driver.Value, error) {
	if !np.Valid {
		return nil, nil //nolint:nilnil // driver.Valuer requires nil,nil for NULL
	}

	data, err := json.Marshal(np.Polygon)
	if err != nil {
		return nil, fmt.Errorf("marshaling polygon: %w", err)
	}

	return string(data), nil
}

// Equal compares two NullPolygon values.
func (np NullPolygon) Equal(other NullPolygon) bool {
	if !np.Valid && !other.Valid {
		return true
	}

	if np.Valid != other.Valid {
		return false
	}

	if len(np.Polygon) != len(other.Polygon) {
		return false
	}

	for i := range np.Polygon {
		if !np.Polygon[i].Equal(other.Polygon[i]) {
			return false
		}
	}

	return true
}

// MultiPoint represents a collection of points.
type MultiPoint []Point

// multiPointJSON is the GeoJSON representation.
type multiPointJSON struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

// MarshalJSON implements json.Marshaler.
func (mp MultiPoint) MarshalJSON() ([]byte, error) {
	coords := make([][]float64, len(mp))

	for i, p := range mp {
		if p.Depth != 0 {
			coords[i] = []float64{p.Lon, p.Lat, p.Depth}
		} else {
			coords[i] = []float64{p.Lon, p.Lat}
		}
	}

	return json.Marshal(multiPointJSON{
		Type:        "MultiPoint",
		Coordinates: coords,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (mp *MultiPoint) UnmarshalJSON(data []byte) error {
	var mpj multiPointJSON

	err := json.Unmarshal(data, &mpj)
	if err != nil {
		return err
	}

	if mpj.Type != "MultiPoint" {
		return fmt.Errorf("%w: got %q", errUnexpectedMultiPointType, mpj.Type)
	}

	*mp = make(MultiPoint, len(mpj.Coordinates))

	for i, coord := range mpj.Coordinates {
		if len(coord) < 2 {
			return fmt.Errorf("%w: got %d", errInsufficientCoordinates, len(coord))
		}

		(*mp)[i] = Point{
			Lon: coord[0],
			Lat: coord[1],
		}

		if len(coord) >= 3 {
			(*mp)[i].Depth = coord[2]
		}
	}

	return nil
}

// GeoJSONType returns the GeoJSON type string.
func (mp MultiPoint) GeoJSONType() string {
	return "MultiPoint"
}

// MultiLineString represents a collection of line strings.
type MultiLineString []LineString

// multiLineStringJSON is the GeoJSON representation.
type multiLineStringJSON struct {
	Type        string        `json:"type"`
	Coordinates [][][]float64 `json:"coordinates"`
}

// MarshalJSON implements json.Marshaler.
func (mls MultiLineString) MarshalJSON() ([]byte, error) {
	coords := make([][][]float64, len(mls))

	for i, ls := range mls {
		coords[i] = make([][]float64, len(ls))

		for j, p := range ls {
			if p.Depth != 0 {
				coords[i][j] = []float64{p.Lon, p.Lat, p.Depth}
			} else {
				coords[i][j] = []float64{p.Lon, p.Lat}
			}
		}
	}

	return json.Marshal(multiLineStringJSON{
		Type:        "MultiLineString",
		Coordinates: coords,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (mls *MultiLineString) UnmarshalJSON(data []byte) error {
	var mlsj multiLineStringJSON

	err := json.Unmarshal(data, &mlsj)
	if err != nil {
		return err
	}

	if mlsj.Type != "MultiLineString" {
		return fmt.Errorf("%w: got %q", errUnexpectedMultiLineStringType, mlsj.Type)
	}

	*mls = make(MultiLineString, len(mlsj.Coordinates))

	for i, lineCoords := range mlsj.Coordinates {
		(*mls)[i] = make(LineString, len(lineCoords))

		for j, coord := range lineCoords {
			if len(coord) < 2 {
				return fmt.Errorf("%w: got %d", errInsufficientCoordinates, len(coord))
			}

			(*mls)[i][j] = Point{
				Lon: coord[0],
				Lat: coord[1],
			}

			if len(coord) >= 3 {
				(*mls)[i][j].Depth = coord[2]
			}
		}
	}

	return nil
}

// GeoJSONType returns the GeoJSON type string.
func (mls MultiLineString) GeoJSONType() string {
	return "MultiLineString"
}

// MultiPolygon represents a collection of polygons.
type MultiPolygon []Polygon

// multiPolygonJSON is the GeoJSON representation.
type multiPolygonJSON struct {
	Type        string          `json:"type"`
	Coordinates [][][][]float64 `json:"coordinates"`
}

// MarshalJSON implements json.Marshaler.
func (mpg MultiPolygon) MarshalJSON() ([]byte, error) {
	coords := make([][][][]float64, len(mpg))

	for i, polygon := range mpg {
		coords[i] = make([][][]float64, len(polygon))

		for j, ring := range polygon {
			// Auto-close rings during marshaling
			closedRing := ring.Close()
			coords[i][j] = make([][]float64, len(closedRing))

			for k, pt := range closedRing {
				if pt.Depth != 0 {
					coords[i][j][k] = []float64{pt.Lon, pt.Lat, pt.Depth}
				} else {
					coords[i][j][k] = []float64{pt.Lon, pt.Lat}
				}
			}
		}
	}

	return json.Marshal(multiPolygonJSON{
		Type:        "MultiPolygon",
		Coordinates: coords,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (mpg *MultiPolygon) UnmarshalJSON(data []byte) error {
	var mpgj multiPolygonJSON

	err := json.Unmarshal(data, &mpgj)
	if err != nil {
		return err
	}

	if mpgj.Type != "MultiPolygon" {
		return fmt.Errorf("%w: got %q", errUnexpectedMultiPolygonType, mpgj.Type)
	}

	*mpg = make(MultiPolygon, len(mpgj.Coordinates))

	for i, polygonCoords := range mpgj.Coordinates {
		(*mpg)[i] = make(Polygon, len(polygonCoords))

		for j, ringCoords := range polygonCoords {
			(*mpg)[i][j] = make(Ring, len(ringCoords))

			for k, coord := range ringCoords {
				if len(coord) < 2 {
					return fmt.Errorf("%w: got %d", errInsufficientCoordinates, len(coord))
				}

				(*mpg)[i][j][k] = Point{
					Lon: coord[0],
					Lat: coord[1],
				}

				if len(coord) >= 3 {
					(*mpg)[i][j][k].Depth = coord[2]
				}
			}
		}
	}

	return nil
}

// GeoJSONType returns the GeoJSON type string.
func (mpg MultiPolygon) GeoJSONType() string {
	return "MultiPolygon"
}

// geometryTypeJSON is used to extract the type field from GeoJSON.
type geometryTypeJSON struct {
	Type string `json:"type"`
}

// UnmarshalGeometry parses GeoJSON and returns the appropriate Geometry type.
func UnmarshalGeometry(data []byte) (Geometry, error) {
	var gt geometryTypeJSON
	if err := json.Unmarshal(data, &gt); err != nil {
		return nil, err
	}

	switch gt.Type {
	case "Point":
		var p Point
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		return p, nil
	case "LineString":
		var ls LineString
		if err := json.Unmarshal(data, &ls); err != nil {
			return nil, err
		}
		return ls, nil
	case "Polygon":
		var p Polygon
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, err
		}
		return p, nil
	case "MultiPoint":
		var mp MultiPoint
		if err := json.Unmarshal(data, &mp); err != nil {
			return nil, err
		}
		return mp, nil
	case "MultiLineString":
		var mls MultiLineString
		if err := json.Unmarshal(data, &mls); err != nil {
			return nil, err
		}
		return mls, nil
	case "MultiPolygon":
		var mpg MultiPolygon
		if err := json.Unmarshal(data, &mpg); err != nil {
			return nil, err
		}
		return mpg, nil
	default:
		return nil, fmt.Errorf("%w: %q", errUnknownGeometryType, gt.Type)
	}
}

// Feature represents a GeoJSON Feature with geometry, properties, and optional ID.
type Feature struct {
	ID         any
	Geometry   Geometry
	Properties map[string]any
}

// featureJSON is the GeoJSON representation of a Feature.
type featureJSON struct {
	Type       string         `json:"type"`
	ID         any            `json:"id,omitzero"`
	Geometry   jsontext.Value `json:"geometry"`
	Properties map[string]any `json:"properties"`
}

// NewFeature creates a new Feature with the given geometry.
func NewFeature(geometry Geometry) *Feature {
	return &Feature{
		Geometry:   geometry,
		Properties: make(map[string]any),
	}
}

// MarshalJSON implements json.Marshaler.
func (f Feature) MarshalJSON() ([]byte, error) {
	var geomData jsontext.Value

	if f.Geometry != nil {
		data, err := json.Marshal(f.Geometry)
		if err != nil {
			return nil, fmt.Errorf("marshaling geometry: %w", err)
		}
		geomData = data
	} else {
		geomData = []byte("null")
	}

	props := f.Properties
	if props == nil {
		props = make(map[string]any)
	}

	return json.Marshal(featureJSON{
		Type:       "Feature",
		ID:         f.ID,
		Geometry:   geomData,
		Properties: props,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (f *Feature) UnmarshalJSON(data []byte) error {
	var fj featureJSON

	err := json.Unmarshal(data, &fj)
	if err != nil {
		return err
	}

	if fj.Type != "Feature" {
		return fmt.Errorf("expected type Feature, got %q", fj.Type)
	}

	f.ID = fj.ID
	f.Properties = fj.Properties

	// Handle null geometry
	if len(fj.Geometry) == 0 || string(fj.Geometry) == "null" {
		f.Geometry = nil
		return nil
	}

	geom, err := UnmarshalGeometry(fj.Geometry)
	if err != nil {
		return fmt.Errorf("unmarshaling geometry: %w", err)
	}

	f.Geometry = geom

	return nil
}

// UnmarshalFeature parses GeoJSON and returns a Feature.
func UnmarshalFeature(data []byte) (*Feature, error) {
	var f Feature
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// FeatureCollection represents a GeoJSON FeatureCollection.
type FeatureCollection struct {
	Features []*Feature
}

// featureCollectionJSON is the GeoJSON representation.
type featureCollectionJSON struct {
	Type     string           `json:"type"`
	Features []jsontext.Value `json:"features"`
}

// NewFeatureCollection creates a new empty FeatureCollection.
func NewFeatureCollection() *FeatureCollection {
	return &FeatureCollection{
		Features: make([]*Feature, 0),
	}
}

// Append adds a feature to the collection and returns the collection for chaining.
func (fc *FeatureCollection) Append(feature *Feature) *FeatureCollection {
	fc.Features = append(fc.Features, feature)
	return fc
}

// MarshalJSON implements json.Marshaler.
func (fc FeatureCollection) MarshalJSON() ([]byte, error) {
	features := make([]jsontext.Value, len(fc.Features))

	for i, f := range fc.Features {
		data, err := json.Marshal(f)
		if err != nil {
			return nil, fmt.Errorf("marshaling feature %d: %w", i, err)
		}
		features[i] = data
	}

	return json.Marshal(struct {
		Type     string           `json:"type"`
		Features []jsontext.Value `json:"features"`
	}{
		Type:     "FeatureCollection",
		Features: features,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (fc *FeatureCollection) UnmarshalJSON(data []byte) error {
	var fcj featureCollectionJSON

	err := json.Unmarshal(data, &fcj)
	if err != nil {
		return err
	}

	if fcj.Type != "FeatureCollection" {
		return fmt.Errorf("expected type FeatureCollection, got %q", fcj.Type)
	}

	fc.Features = make([]*Feature, len(fcj.Features))

	for i, rawFeature := range fcj.Features {
		var f Feature
		if err := json.Unmarshal(rawFeature, &f); err != nil {
			return fmt.Errorf("unmarshaling feature %d: %w", i, err)
		}
		fc.Features[i] = &f
	}

	return nil
}

// UnmarshalFeatureCollection parses GeoJSON and returns a FeatureCollection.
func UnmarshalFeatureCollection(data []byte) (*FeatureCollection, error) {
	var fc FeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, err
	}
	return &fc, nil
}
