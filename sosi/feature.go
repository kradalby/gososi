package sosi

import (
	"fmt"
	"strings"
)

// Feature represents a SOSI feature with its properties and coordinates
type Feature struct {
	ID          int          // Feature ID
	Type        string       // PUNKT, KURVE, SVERM
	ObjectType  string       // User-defined object type
	Coordinates []Coordinate // Feature coordinates
}

// GenerateFeature converts a SOSI feature to its string representation
func GenerateFeature(feature SOSIFeature, config SOSIConfig) string {
	var builder strings.Builder

	// Feature header: .[TYPE] [ID]:
	builder.WriteString(fmt.Sprintf(".%s %d:\n", feature.Type, feature.ID))
	builder.WriteString(fmt.Sprintf("..OBJTYPE %s\n", feature.ObjectType))
	builder.WriteString("..NØH\n")

	// Transform and add coordinates
	for _, coord := range feature.Coordinates {
		coordString := TransformCoordinate(coord, config)
		builder.WriteString(coordString + "\n")
	}

	return builder.String()
}

// ConvertGeometryTypeToSOSI maps GeoJSON geometry types to SOSI types
func ConvertGeometryTypeToSOSI(geojsonType string) (string, error) {
	typeMap := map[string]string{
		"Point":      "PUNKT",
		"MultiPoint": "SVERM",
		"LineString": "KURVE",
		"Polygon":    "FLATE",
	}

	sosiType, exists := typeMap[geojsonType]
	if !exists {
		return "", &ConversionError{
			Type:    "UnsupportedGeometry",
			Message: fmt.Sprintf("Cannot convert geometry type '%s' to SOSI", geojsonType),
		}
	}

	return sosiType, nil
}
