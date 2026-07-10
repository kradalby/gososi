package sosi

import "fmt"

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
