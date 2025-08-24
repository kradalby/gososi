package sosi

import (
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

// ToGeoJSON converts a SOSIDocument to a GeoJSON FeatureCollection
// Uses github.com/paulmach/orb/geojson for standard-compliant GeoJSON output
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

// convertSOSIFeatureToGeoJSON converts a single SOSIFeature to GeoJSON Feature
func convertSOSIFeatureToGeoJSON(feature *SOSIFeature, header *SOSIHeader) (*geojson.Feature, error) {
	// Convert geometry based on SOSI type
	var geometry orb.Geometry
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
		geoFeature.Properties = make(map[string]interface{})
	}

	// Add feature ID
	geoFeature.Properties["sosi_id"] = feature.ID

	// Add object type
	if feature.ObjectType != "" {
		geoFeature.Properties["objtype"] = feature.ObjectType
	}

	// Add parsed properties from SOSI
	if feature.Properties != nil {
		for key, value := range feature.Properties {
			geoFeature.Properties[key] = value
		}
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

// convertPointGeometry converts SOSI PUNKT to GeoJSON Point
func convertPointGeometry(feature *SOSIFeature) (orb.Geometry, error) {
	if len(feature.Coordinates) == 0 {
		return nil, fmt.Errorf("point feature has no coordinates")
	}

	coord := feature.Coordinates[0]
	return orb.Point{coord.X, coord.Y}, nil
}

// convertLineStringGeometry converts SOSI KURVE to GeoJSON LineString
func convertLineStringGeometry(feature *SOSIFeature) (orb.Geometry, error) {
	if len(feature.Coordinates) < 2 {
		return nil, fmt.Errorf("linestring feature needs at least 2 coordinates, got %d", len(feature.Coordinates))
	}

	lineString := orb.LineString{}
	for _, coord := range feature.Coordinates {
		lineString = append(lineString, orb.Point{coord.X, coord.Y})
	}

	return lineString, nil
}

// convertPolygonGeometry converts SOSI FLATE to GeoJSON Polygon
// Handles both simple polygons and polygons with holes
func convertPolygonGeometry(feature *SOSIFeature) (orb.Geometry, error) {
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

// convertArcGeometry converts SOSI BUEP (arc) to GeoJSON LineString
// The arc should already be interpolated to linestring coordinates
func convertArcGeometry(feature *SOSIFeature) (orb.Geometry, error) {
	if len(feature.Coordinates) < 2 {
		return nil, fmt.Errorf("arc feature needs at least 2 coordinates after interpolation, got %d", len(feature.Coordinates))
	}

	lineString := orb.LineString{}
	for _, coord := range feature.Coordinates {
		lineString = append(lineString, orb.Point{coord.X, coord.Y})
	}

	return lineString, nil
}

// convertSimplePolygon converts coordinates to a simple polygon
func convertSimplePolygon(coordinates []Coordinate) orb.Polygon {
	ring := orb.Ring{}
	for _, coord := range coordinates {
		ring = append(ring, orb.Point{coord.X, coord.Y})
	}

	// Ensure ring is closed
	if len(ring) > 0 && !ring[0].Equal(ring[len(ring)-1]) {
		ring = append(ring, ring[0])
	}

	return orb.Polygon{ring}
}

// ToGeoJSONWithReferences converts SOSIDocument to GeoJSON with full reference resolution
// This handles complex polygons that reference other features
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

// convertSOSIFeatureWithReferences converts SOSI feature with reference resolution support
func convertSOSIFeatureWithReferences(feature *SOSIFeature, header *SOSIHeader, featureMap map[int]*SOSIFeature) (*geojson.Feature, error) {
	var geometry orb.Geometry
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
		geoFeature.Properties = make(map[string]interface{})
	}

	// Add feature ID
	geoFeature.Properties["sosi_id"] = feature.ID

	// Add object type
	if feature.ObjectType != "" {
		geoFeature.Properties["objtype"] = feature.ObjectType
	}

	// Add parsed properties from SOSI
	if feature.Properties != nil {
		for key, value := range feature.Properties {
			geoFeature.Properties[key] = value
		}
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

// convertPolygonWithReferences converts FLATE with reference resolution
func convertPolygonWithReferences(feature *SOSIFeature, featureMap map[int]*SOSIFeature) (orb.Geometry, error) {
	// Simple polygon case
	if len(feature.Coordinates) > 0 && len(feature.Refs) == 0 {
		return convertSimplePolygon(feature.Coordinates), nil
	}

	// Polygon with references
	if len(feature.Refs) > 0 {
		return buildPolygonFromReferences(feature, featureMap)
	}

	return nil, fmt.Errorf("polygon feature has no coordinates or references")
}

// buildPolygonFromReferences constructs polygon from referenced KURVE features
func buildPolygonFromReferences(feature *SOSIFeature, featureMap map[int]*SOSIFeature) (orb.Polygon, error) {
	// Handle polygon with holes structure
	if len(feature.OuterRing) > 0 {
		return buildPolygonWithHoles(feature, featureMap)
	}

	// Simple polygon from references
	ring, err := buildRingFromReferences(feature.Refs, featureMap)
	if err != nil {
		return nil, fmt.Errorf("building outer ring: %w", err)
	}

	return orb.Polygon{ring}, nil
}

// buildPolygonWithHoles constructs polygon with holes from references
func buildPolygonWithHoles(feature *SOSIFeature, featureMap map[int]*SOSIFeature) (orb.Polygon, error) {
	// Build outer ring
	outerRing, err := buildRingFromReferences(feature.OuterRing, featureMap)
	if err != nil {
		return nil, fmt.Errorf("building outer ring: %w", err)
	}

	polygon := orb.Polygon{outerRing}

	// Build holes
	for i, holeRefs := range feature.Holes {
		hole, err := buildRingFromReferences(holeRefs, featureMap)
		if err != nil {
			return nil, fmt.Errorf("building hole %d: %w", i, err)
		}
		polygon = append(polygon, hole)
	}

	return polygon, nil
}

// buildRingFromReferences constructs a ring from referenced KURVE features
func buildRingFromReferences(refs []int, featureMap map[int]*SOSIFeature) (orb.Ring, error) {
	ring := orb.Ring{}

	for _, refID := range refs {
		// Handle negative references by taking absolute value
		// Negative references in SOSI indicate reverse direction or special handling
		absRefID := refID
		if refID < 0 {
			absRefID = -refID
		}

		refFeature, exists := featureMap[absRefID]
		if !exists {
			// Return error for missing references - this indicates data integrity issues
			return nil, fmt.Errorf("referenced feature %d not found - this may indicate incomplete or corrupted SOSI data", absRefID)
		}

		switch refFeature.Type {
		case "KURVE":
			// Add coordinates from referenced KURVE
			// If reference was negative, reverse the coordinate order
			coords := refFeature.Coordinates
			if refID < 0 {
				// Reverse coordinate order for negative references
				for i := len(coords) - 1; i >= 0; i-- {
					ring = append(ring, orb.Point{coords[i].X, coords[i].Y})
				}
			} else {
				// Normal order for positive references
				for _, coord := range coords {
					ring = append(ring, orb.Point{coord.X, coord.Y})
				}
			}
		case "FLATE":
			// Handle FLATE references - use the outer ring of the referenced polygon
			// This is for cases where a polygon hole is defined by another polygon
			polygonGeometry, err := convertPolygonWithReferences(refFeature, featureMap)
			if err != nil {
				return nil, fmt.Errorf("failed to convert referenced FLATE %d: %w", refID, err)
			}

			// Get the outer ring (first ring) of the referenced polygon
			if polygon, ok := polygonGeometry.(orb.Polygon); ok && len(polygon) > 0 {
				outerRing := polygon[0]
				if refID < 0 {
					// Reverse ring direction for negative references
					for i := len(outerRing) - 1; i >= 0; i-- {
						ring = append(ring, outerRing[i])
					}
				} else {
					// Normal ring direction for positive references
					ring = append(ring, outerRing...)
				}
			}
		default:
			return nil, fmt.Errorf("referenced feature %d is not KURVE or FLATE, got %s", refID, refFeature.Type)
		}
	}

	// Ensure ring is closed
	if len(ring) > 0 && !ring[0].Equal(ring[len(ring)-1]) {
		ring = append(ring, ring[0])
	}

	return ring, nil
}
