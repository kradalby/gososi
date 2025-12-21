package sosi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kradalby/gososi/geojson"
)

// GeoJSONtoSOSI converts GeoJSON data to SOSI format
func GeoJSONtoSOSI(geojsonData []byte, objectTypes map[string]string) ([]byte, error) {
	// Parse GeoJSON
	fc, err := geojson.UnmarshalFeatureCollection(geojsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GeoJSON: %w", err)
	}

	// Create SOSI builder with default config
	builder := NewBuilder(DefaultConfig())

	// Convert each feature
	for i, feature := range fc.Features {
		sosiFeature, err := convertGeoJSONFeatureToSOSI(feature, i+1, objectTypes)
		if err != nil {
			return nil, fmt.Errorf("failed to convert feature %d: %w", i, err)
		}
		builder.AddFeature(sosiFeature)
	}

	// Build final SOSI string
	sosiString, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build SOSI: %w", err)
	}

	return []byte(sosiString), nil
}

// SOSItoGeoJSON converts SOSI data to GeoJSON format
func SOSItoGeoJSON(sosiData []byte) ([]byte, error) {
	// Parse SOSI
	parser := NewParser()
	doc, err := parser.Parse(strings.NewReader(string(sosiData)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SOSI: %w", err)
	}

	// Convert to GeoJSON
	fc, err := doc.ToGeoJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to GeoJSON: %w", err)
	}

	// Marshal to JSON
	geojsonData, err := json.Marshal(fc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GeoJSON: %w", err)
	}

	return geojsonData, nil
}

// AnalyzeGeoJSON analyzes GeoJSON data and returns feature descriptions
func AnalyzeGeoJSON(geojsonData []byte) (map[string]string, error) {
	// Parse GeoJSON
	fc, err := geojson.UnmarshalFeatureCollection(geojsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GeoJSON: %w", err)
	}

	analysis := make(map[string]string)

	// Analyze each feature
	for i, feature := range fc.Features {
		// Create feature ID
		featureID := fmt.Sprintf("c%d", i)
		if feature.ID != nil {
			if id, ok := feature.ID.(string); ok {
				featureID = id
			}
		}

		// Get geometry type
		geomType := string(feature.Geometry.GeoJSONType())
		
		// Get label from properties
		label := "unlabeled"
		if labelProp, exists := feature.Properties["label"]; exists {
			if labelStr, ok := labelProp.(string); ok {
				label = labelStr
			}
		}

		// Count coordinates
		coordCount := countCoordinates(feature.Geometry)

		// Create description matching JavaScript format
		description := fmt.Sprintf("%s: %s %s = %d coords", featureID, label, geomType, coordCount)
		analysis[featureID] = description
	}

	return analysis, nil
}

// convertGeoJSONFeatureToSOSI converts a single GeoJSON feature to SOSI
func convertGeoJSONFeatureToSOSI(feature *geojson.Feature, id int, objectTypes map[string]string) (SOSIFeature, error) {
	// Get feature ID - default to c[index] pattern
	featureID := fmt.Sprintf("c%d", id-1)
	if feature.ID != nil {
		if idStr, ok := feature.ID.(string); ok {
			featureID = idStr
		}
	}

	objectType, exists := objectTypes[featureID]
	if !exists {
		return SOSIFeature{}, fmt.Errorf("no object type provided for feature %s", featureID)
	}

	// Convert geometry type
	geomType := string(feature.Geometry.GeoJSONType())
	sosiType, err := ConvertGeometryTypeToSOSI(geomType)
	if err != nil {
		return SOSIFeature{}, err
	}

	// Convert coordinates
	coordinates, err := convertGeoJSONGeometryToCoordinates(feature.Geometry)
	if err != nil {
		return SOSIFeature{}, fmt.Errorf("failed to convert geometry: %w", err)
	}

	return SOSIFeature{
		ID:          id,
		Type:        sosiType,
		ObjectType:  objectType,
		Coordinates: coordinates,
		Properties:  feature.Properties,
	}, nil
}

// convertGeoJSONGeometryToCoordinates converts GeoJSON geometry to SOSI coordinates.
func convertGeoJSONGeometryToCoordinates(geom geojson.Geometry) ([]Coordinate, error) {
	switch g := geom.(type) {
	case geojson.Point:
		return []Coordinate{{X: g.Lon, Y: g.Lat, Z: g.Depth}}, nil
	case geojson.LineString:
		coords := make([]Coordinate, len(g))
		for i, point := range g {
			coords[i] = Coordinate{X: point.Lon, Y: point.Lat, Z: point.Depth}
		}
		return coords, nil
	case geojson.Polygon:
		if len(g) == 0 {
			return nil, fmt.Errorf("empty polygon")
		}
		// Use outer ring coordinates
		ring := g[0]
		coords := make([]Coordinate, len(ring))
		for i, point := range ring {
			coords[i] = Coordinate{X: point.Lon, Y: point.Lat, Z: point.Depth}
		}
		return coords, nil
	default:
		return nil, fmt.Errorf("unsupported geometry type: %T", geom)
	}
}

// countCoordinates counts the number of coordinate points in a geometry.
func countCoordinates(geom geojson.Geometry) int {
	switch g := geom.(type) {
	case geojson.Point:
		return 1
	case geojson.LineString:
		return len(g)
	case geojson.Polygon:
		total := 0
		for _, ring := range g {
			total += len(ring)
		}
		return total
	case geojson.MultiPoint:
		return len(g)
	case geojson.MultiLineString:
		total := 0
		for _, line := range g {
			total += len(line)
		}
		return total
	case geojson.MultiPolygon:
		total := 0
		for _, polygon := range g {
			for _, ring := range polygon {
				total += len(ring)
			}
		}
		return total
	default:
		return 0
	}
}