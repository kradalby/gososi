package sosi

import (
	"fmt"
	"slices"

	"github.com/kradalby/gososi/geojson"
)

// ToGeoJSON converts a SOSIDocument to a GeoJSON FeatureCollection.
func (doc *SOSIDocument) ToGeoJSON() (*geojson.FeatureCollection, error) {
	fc := geojson.NewFeatureCollection()

	for _, feature := range doc.Features {
		geoJSONFeature, err := convertSOSIFeatureToGeoJSON(&feature, &doc.Header)
		if err != nil {
			return nil, fmt.Errorf("converting feature %d: %w", feature.ID, err)
		}
		fc.Append(geoJSONFeature)
	}

	return fc, nil
}

// convertSOSIFeatureToGeoJSON converts a single SOSIFeature to GeoJSON Feature.
func convertSOSIFeatureToGeoJSON(feature *SOSIFeature, header *SOSIHeader) (*geojson.Feature, error) {
	// Convert geometry based on SOSI type
	var geometry geojson.Geometry
	var err error

	switch feature.Type {
	case "PUNKT":
		geometry, err = convertPointGeometry(feature)
	case "KURVE":
		geometry, err = convertLineStringGeometry(feature)
	case "FLATE":
		geometry, err = convertPolygonGeometry(feature)
	case "BUEP":
		geometry, err = convertArcGeometry(feature)
	case "TEKST":
		// TEKST is treated as a point with additional text attributes
		geometry, err = convertPointGeometry(feature)
	default:
		return nil, fmt.Errorf("unsupported geometry type: %s", feature.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("converting %s geometry: %w", feature.Type, err)
	}

	// Create GeoJSON feature with properties
	geoFeature := geojson.NewFeature(geometry)

	// Add SOSI-specific properties
	if geoFeature.Properties == nil {
		geoFeature.Properties = make(map[string]any)
	}

	// Add feature ID
	geoFeature.Properties["sosi_id"] = feature.ID

	// Add object type
	if feature.ObjectType != "" {
		geoFeature.Properties["objtype"] = feature.ObjectType
	}

	// Add parsed properties from SOSI
	for k, v := range feature.Properties {
		geoFeature.Properties[k] = cloneValue(v)
	}

	// Add coordinate system information from header
	if header != nil && header.CoordSystem != 0 {
		geoFeature.Properties["coord_system"] = header.CoordSystem
		if srid := GetSRIDFromCoordSystem(header.CoordSystem); srid != "" {
			geoFeature.Properties["srid"] = srid
		}
	}

	return geoFeature, nil
}

// convertPointGeometry converts SOSI PUNKT to GeoJSON Point.
func convertPointGeometry(feature *SOSIFeature) (geojson.Geometry, error) {
	if len(feature.Coordinates) == 0 {
		return nil, fmt.Errorf("point feature has no coordinates")
	}

	coord := feature.Coordinates[0]
	return geojson.Point{Lon: coord.X, Lat: coord.Y, Depth: coord.Z}, nil
}

// convertLineStringGeometry converts SOSI KURVE to GeoJSON LineString.
func convertLineStringGeometry(feature *SOSIFeature) (geojson.Geometry, error) {
	if len(feature.Coordinates) < 2 {
		return nil, fmt.Errorf("linestring feature needs at least 2 coordinates, got %d", len(feature.Coordinates))
	}

	lineString := make(geojson.LineString, len(feature.Coordinates))
	for i, coord := range feature.Coordinates {
		lineString[i] = geojson.Point{Lon: coord.X, Lat: coord.Y, Depth: coord.Z}
	}

	return lineString, nil
}

// convertPolygonGeometry converts SOSI FLATE to GeoJSON Polygon.
// Handles both simple polygons and polygons with holes.
func convertPolygonGeometry(feature *SOSIFeature) (geojson.Geometry, error) {
	// For simple polygon with direct coordinates
	if len(feature.Coordinates) > 0 && len(feature.Refs) == 0 {
		return convertSimplePolygon(feature.Coordinates), nil
	}

	// For polygon with references - this would need access to referenced features
	// This is a complex case that requires the full document context
	if len(feature.Refs) > 0 {
		return nil, fmt.Errorf("polygon with references requires document context - use ToGeoJSONWithReferences")
	}

	return nil, fmt.Errorf("polygon feature has no coordinates or references")
}

// convertArcGeometry converts SOSI BUEP (arc) to GeoJSON LineString.
// The arc should already be interpolated to linestring coordinates.
func convertArcGeometry(feature *SOSIFeature) (geojson.Geometry, error) {
	if len(feature.Coordinates) < 2 {
		return nil, fmt.Errorf("arc feature needs at least 2 coordinates after interpolation, got %d", len(feature.Coordinates))
	}

	lineString := make(geojson.LineString, len(feature.Coordinates))
	for i, coord := range feature.Coordinates {
		lineString[i] = geojson.Point{Lon: coord.X, Lat: coord.Y, Depth: coord.Z}
	}

	return lineString, nil
}

// convertSimplePolygon converts coordinates to a simple polygon.
func convertSimplePolygon(coordinates []Coordinate) geojson.Polygon {
	ring := make(geojson.Ring, len(coordinates))
	for i, coord := range coordinates {
		ring[i] = geojson.Point{Lon: coord.X, Lat: coord.Y, Depth: coord.Z}
	}

	// Ring will be auto-closed by geojson.Polygon.MarshalJSON
	return geojson.Polygon{ring}
}

// ToGeoJSONWithReferences converts SOSIDocument to GeoJSON with full reference resolution.
// This handles complex polygons that reference other features.
func (doc *SOSIDocument) ToGeoJSONWithReferences() (*geojson.FeatureCollection, error) {
	fc := geojson.NewFeatureCollection()

	// Build lookup map for referenced features
	featureMap := make(map[int]*SOSIFeature)
	for i := range doc.Features {
		featureMap[doc.Features[i].ID] = &doc.Features[i]
	}

	for _, feature := range doc.Features {
		geoJSONFeature, err := convertSOSIFeatureWithReferences(&feature, &doc.Header, featureMap)
		if err != nil {
			return nil, fmt.Errorf("converting feature %d with references: %w", feature.ID, err)
		}
		fc.Append(geoJSONFeature)
	}

	return fc, nil
}

// convertSOSIFeatureWithReferences converts SOSI feature with reference resolution support.
func convertSOSIFeatureWithReferences(feature *SOSIFeature, header *SOSIHeader, featureMap map[int]*SOSIFeature) (*geojson.Feature, error) {
	var geometry geojson.Geometry
	var err error

	switch feature.Type {
	case "PUNKT":
		geometry, err = convertPointGeometry(feature)
	case "KURVE":
		geometry, err = convertLineStringGeometry(feature)
	case "FLATE":
		geometry, err = convertPolygonWithReferences(feature, featureMap)
	case "BUEP":
		geometry, err = convertArcGeometry(feature)
	case "TEKST":
		// TEKST is treated as a point with additional text attributes
		geometry, err = convertPointGeometry(feature)
	default:
		return nil, fmt.Errorf("unsupported geometry type: %s", feature.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("converting %s geometry: %w", feature.Type, err)
	}

	// Create GeoJSON feature with properties
	geoFeature := geojson.NewFeature(geometry)

	// Add SOSI-specific properties
	if geoFeature.Properties == nil {
		geoFeature.Properties = make(map[string]any)
	}

	// Add feature ID
	geoFeature.Properties["sosi_id"] = feature.ID

	// Add object type
	if feature.ObjectType != "" {
		geoFeature.Properties["objtype"] = feature.ObjectType
	}

	// Add parsed properties from SOSI
	for k, v := range feature.Properties {
		geoFeature.Properties[k] = cloneValue(v)
	}

	// Add coordinate system information
	if header != nil && header.CoordSystem != 0 {
		geoFeature.Properties["coord_system"] = header.CoordSystem
		if srid := GetSRIDFromCoordSystem(header.CoordSystem); srid != "" {
			geoFeature.Properties["srid"] = srid
		}
	}

	return geoFeature, nil
}

// convertPolygonWithReferences converts FLATE with reference resolution.
func convertPolygonWithReferences(feature *SOSIFeature, featureMap map[int]*SOSIFeature) (geojson.Geometry, error) {
	r := refResolver{features: featureMap, resolving: make(map[int]bool)}
	return r.polygon(feature)
}

// refResolver expands FLATE references. A FLATE may reference another FLATE,
// so resolving tracks the FLATE IDs on the current expansion path: revisiting
// one means the input references itself, which would otherwise recurse forever.
type refResolver struct {
	features  map[int]*SOSIFeature
	resolving map[int]bool
}

func (r refResolver) polygon(feature *SOSIFeature) (geojson.Polygon, error) {
	if r.resolving[feature.ID] {
		return nil, fmt.Errorf("reference cycle through FLATE %d", feature.ID)
	}
	r.resolving[feature.ID] = true
	defer delete(r.resolving, feature.ID)

	// Simple polygon case
	if len(feature.Coordinates) > 0 && len(feature.Refs) == 0 {
		return convertSimplePolygon(feature.Coordinates), nil
	}

	if len(feature.Refs) == 0 {
		return nil, fmt.Errorf("polygon feature has no coordinates or references")
	}

	// Simple polygon from references
	if len(feature.OuterRing) == 0 {
		ring, err := r.ring(feature.Refs)
		if err != nil {
			return nil, fmt.Errorf("building outer ring: %w", err)
		}
		return geojson.Polygon{ring}, nil
	}

	outerRing, err := r.ring(feature.OuterRing)
	if err != nil {
		return nil, fmt.Errorf("building outer ring: %w", err)
	}

	polygon := geojson.Polygon{outerRing}
	for i, holeRefs := range feature.Holes {
		hole, err := r.ring(holeRefs)
		if err != nil {
			return nil, fmt.Errorf("building hole %d: %w", i, err)
		}
		polygon = append(polygon, hole)
	}

	return polygon, nil
}

// ring constructs a ring from referenced KURVE or FLATE features.
func (r refResolver) ring(refs []int) (geojson.Ring, error) {
	ring := geojson.Ring{}

	for _, refID := range refs {
		// Negative references in SOSI indicate reverse direction
		absRefID := refID
		if refID < 0 {
			absRefID = -refID
		}

		refFeature, exists := r.features[absRefID]
		if !exists {
			// Return error for missing references - this indicates data integrity issues
			return nil, fmt.Errorf("referenced feature %d not found - this may indicate incomplete or corrupted SOSI data", absRefID)
		}

		switch refFeature.Type {
		case "KURVE":
			coords := refFeature.Coordinates
			if refID < 0 {
				for _, coord := range slices.Backward(coords) {
					ring = append(ring, geojson.Point{Lon: coord.X, Lat: coord.Y, Depth: coord.Z})
				}
			} else {
				for _, coord := range coords {
					ring = append(ring, geojson.Point{Lon: coord.X, Lat: coord.Y, Depth: coord.Z})
				}
			}
		case "FLATE":
			// A hole may be defined by another polygon; use its outer ring.
			polygon, err := r.polygon(refFeature)
			if err != nil {
				return nil, fmt.Errorf("failed to convert referenced FLATE %d: %w", refID, err)
			}

			if len(polygon) > 0 {
				outerRing := polygon[0]
				if refID < 0 {
					for _, pt := range slices.Backward(outerRing) {
						ring = append(ring, pt)
					}
				} else {
					ring = append(ring, outerRing...)
				}
			}
		default:
			return nil, fmt.Errorf("referenced feature %d is not KURVE or FLATE, got %s", refID, refFeature.Type)
		}
	}

	// Ring will be auto-closed by geojson.Polygon.MarshalJSON if needed
	return ring, nil
}

// cloneValue deep-copies the nested shapes the parser produces (KVALITET and
// REGISTRERINGSVERSJON maps, repeated values as []any), so the GeoJSON output
// shares no mutable state with the document.
func cloneValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		if x == nil {
			return x
		}
		m := make(map[string]any, len(x))
		for k, e := range x {
			m[k] = cloneValue(e)
		}
		return m
	case []any:
		if x == nil {
			return x
		}
		s := make([]any, len(x))
		for i, e := range x {
			s[i] = cloneValue(e)
		}
		return s
	default:
		return v
	}
}
