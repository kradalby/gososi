package sosi

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/kradalby/gososi/geojson"
	"github.com/kradalby/gososi/proj"
)

// GeoJSONToSOSI converts a GeoJSON FeatureCollection to SOSI. By default it
// produces Norwegian-government-compliant output: the data is reprojected into
// an auto-selected UTM zone (KOORDSYS 21-26), heights referenced to NN2000, at
// centimetre precision, with GeoJSON properties emitted as SOSI attributes. Use
// the Option functions to change any of this, including WithKoordSys to pick a
// zone explicitly or KoordSysWGS84Geographic for raw lon/lat.
//
// objectTypes maps each feature (by its GeoJSON id, or "c<index>" when it has
// none) to a SOSI object type emitted as OBJTYPE.
func GeoJSONToSOSI(data []byte, objectTypes map[string]string, opts ...Option) ([]byte, error) {
	o := defaultWriteOptions()
	for _, opt := range opts {
		opt(&o)
	}

	fc, err := geojson.UnmarshalFeatureCollection(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GeoJSON: %w", err)
	}
	if len(fc.Features) == 0 {
		return nil, &ConversionError{Type: "EmptyFeatureCollection", Message: "no features provided"}
	}

	crs, koordsys, err := resolveCRS(fc, o)
	if err != nil {
		return nil, err
	}

	features, bbox, err := buildFeatures(fc, objectTypes, crs)
	if err != nil {
		return nil, err
	}

	var b strings.Builder
	writeHeader(&b, o, koordsys, bbox)
	for _, f := range features {
		writeFeature(&b, f, o)
	}
	b.WriteString(".SLUTT\n")

	return []byte(b.String()), nil
}

// outFeature is a feature ready to serialise.
type outFeature struct {
	seq        int
	sosiType   string
	objType    string
	idString   string
	properties map[string]any
	coords     []Coordinate
}

// resolveCRS determines the target coordinate system and, if projected, the
// projection to apply. A nil CRS means geographic (no reprojection).
func resolveCRS(fc *geojson.FeatureCollection, o writeOptions) (proj.CRS, KoordSys, error) {
	koordsys := o.koordsys
	if koordsys == koordSysAuto {
		lon, lat, ok := centroid(fc)
		if !ok {
			return nil, 0, &ConversionError{Type: "InvalidGeometry", Message: "no coordinates to auto-select a coordinate system"}
		}
		zone := proj.UTMZoneForLon(lon, lat)
		code, ok := proj.KoordSysForZone(zone)
		if !ok {
			return nil, 0, fmt.Errorf("cannot auto-select KOORDSYS for UTM zone %d; set one explicitly with WithKoordSys", zone)
		}
		koordsys = code
	}

	if !koordsys.IsProjected() {
		return nil, koordsys, nil // geographic
	}

	crs, err := proj.FromKOORDSYS(koordsys)
	if err != nil {
		return nil, 0, fmt.Errorf("KOORDSYS %d: %w", int(koordsys), err)
	}
	return crs, koordsys, nil
}

// centroid returns the mean lon/lat over every coordinate in the collection.
func centroid(fc *geojson.FeatureCollection) (lon, lat float64, ok bool) {
	var sumLon, sumLat float64
	var n int
	for _, f := range fc.Features {
		if f.Geometry == nil {
			continue
		}
		coords, err := convertGeoJSONGeometryToCoordinates(f.Geometry)
		if err != nil {
			continue
		}
		for _, c := range coords {
			sumLon += c.X
			sumLat += c.Y
			n++
		}
	}
	if n == 0 {
		return 0, 0, false
	}
	return sumLon / float64(n), sumLat / float64(n), true
}

// buildFeatures converts each GeoJSON feature into an outFeature, reprojecting
// coordinates through crs when non-nil, and accumulates the bounding box in the
// output coordinate space.
func buildFeatures(fc *geojson.FeatureCollection, objectTypes map[string]string, crs proj.CRS) ([]outFeature, BoundingBox, error) {
	bbox := BoundingBox{
		MinLat: math.Inf(1), MinLon: math.Inf(1),
		MaxLat: math.Inf(-1), MaxLon: math.Inf(-1),
	}

	out := make([]outFeature, 0, len(fc.Features))
	for i, f := range fc.Features {
		if f.Geometry == nil {
			return nil, bbox, fmt.Errorf("feature %d has no geometry", i)
		}

		idString := ""
		featureKey := fmt.Sprintf("c%d", i)
		if f.ID != nil {
			idString = featureIDString(f.ID)
			if s, ok := f.ID.(string); ok {
				featureKey = s
			}
		}

		objType, ok := objectTypes[featureKey]
		if !ok {
			return nil, bbox, fmt.Errorf("no object type provided for feature %s", featureKey)
		}

		sosiType, err := ConvertGeometryTypeToSOSI(string(f.Geometry.GeoJSONType()))
		if err != nil {
			return nil, bbox, err
		}

		coords, err := convertGeoJSONGeometryToCoordinates(f.Geometry)
		if err != nil {
			return nil, bbox, fmt.Errorf("feature %s: %w", featureKey, err)
		}
		coords = dropConsecutiveDuplicates(coords)

		if crs != nil {
			for j := range coords {
				e, n := crs.Forward(coords[j].X, coords[j].Y)
				coords[j].X, coords[j].Y = e, n
			}
		}
		UpdateBoundingBox(&bbox, coords)

		out = append(out, outFeature{
			seq:        i + 1,
			sosiType:   sosiType,
			objType:    objType,
			idString:   idString,
			properties: f.Properties,
			coords:     coords,
		})
	}

	return out, bbox, nil
}

// dropConsecutiveDuplicates removes coordinates identical to their predecessor,
// eliminating zero-length segments (SOSI producers reject repeated vertices).
func dropConsecutiveDuplicates(coords []Coordinate) []Coordinate {
	if len(coords) < 2 {
		return coords
	}
	out := coords[:1]
	for _, c := range coords[1:] {
		prev := out[len(out)-1]
		if c.X == prev.X && c.Y == prev.Y && c.Z == prev.Z {
			continue
		}
		out = append(out, c)
	}
	return out
}

func writeHeader(b *strings.Builder, o writeOptions, koordsys KoordSys, bbox BoundingBox) {
	b.WriteString(".HODE\n")
	b.WriteString("..TEGNSETT UTF-8\n")
	fmt.Fprintf(b, "..SOSI-VERSJON %s\n", o.version)
	fmt.Fprintf(b, "..SOSI-NIVÅ %d\n", o.level)
	if o.producer != "" {
		fmt.Fprintf(b, "..PRODUSENT %q\n", o.producer)
	}

	b.WriteString("..TRANSPAR\n")
	fmt.Fprintf(b, "...ENHET %s\n", formatEnhet(o.latlonAcc))
	if o.altAcc != o.latlonAcc {
		fmt.Fprintf(b, "...ENHET-H %s\n", formatEnhet(o.altAcc))
	}
	fmt.Fprintf(b, "...KOORDSYS %d\n", int(koordsys))
	if o.vertDatum != VertDatumNone {
		fmt.Fprintf(b, "...VERT-DATUM %s\n", o.vertDatum)
	}
	b.WriteString("...ORIGO-NØ 0 0\n")

	b.WriteString("..OMRÅDE\n")
	fmt.Fprintf(b, "...MIN-NØ %d %d\n", int(math.Floor(bbox.MinLat)), int(math.Floor(bbox.MinLon)))
	fmt.Fprintf(b, "...MAX-NØ %d %d\n", int(math.Ceil(bbox.MaxLat)), int(math.Ceil(bbox.MaxLon)))
}

func writeFeature(b *strings.Builder, f outFeature, o writeOptions) {
	fmt.Fprintf(b, ".%s %d:\n", f.sosiType, f.seq)
	fmt.Fprintf(b, "..OBJTYPE %s\n", f.objType)

	if o.attributes {
		for _, kv := range attributeLines(f.properties) {
			fmt.Fprintf(b, "..%s %s\n", kv[0], kv[1])
		}
	}
	if o.featureID && f.idString != "" {
		fmt.Fprintf(b, "..id %s\n", f.idString)
	}

	b.WriteString("..NØH\n")
	for _, c := range f.coords {
		fmt.Fprintf(
			b, "%s %s %s\n",
			FormatCoordinateToSOSI(c.Y, o.latlonAcc),
			FormatCoordinateToSOSI(c.X, o.latlonAcc),
			FormatCoordinateToSOSI(c.Z, o.altAcc),
		)
	}
}

// attributeLines returns [key, value] pairs for every scalar property except
// objtype (which becomes OBJTYPE), in deterministic key order.
func attributeLines(props map[string]any) [][2]string {
	if len(props) == 0 {
		return nil
	}
	keys := make([]string, 0, len(props))
	for k := range props {
		if k == "objtype" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	lines := make([][2]string, 0, len(keys))
	for _, k := range keys {
		lines = append(lines, [2]string{k, formatAttrValue(props[k])})
	}
	return lines
}

// formatEnhet formats a coordinate unit (10^-accuracy) with enough decimals to
// represent it, but at least six so cm precision reads as "0.010000".
func formatEnhet(accuracy int) string {
	decimals := max(accuracy, 6)
	return strconv.FormatFloat(math.Pow(10, -float64(accuracy)), 'f', decimals, 64)
}

func formatAttrValue(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		if x == math.Trunc(x) && !math.IsInf(x, 0) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'g', -1, 64)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case bool:
		return strconv.FormatBool(x)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func featureIDString(id any) string {
	switch x := id.(type) {
	case string:
		return x
	case float64:
		if x == math.Trunc(x) && !math.IsInf(x, 0) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'g', -1, 64)
	default:
		return fmt.Sprintf("%v", id)
	}
}
