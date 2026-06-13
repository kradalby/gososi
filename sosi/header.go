package sosi

import (
	"fmt"
	"math"
	"strings"
)

// Header represents SOSI header information
type Header struct {
	CharacterSet     string
	Version          string
	Level            int
	CoordSystem      int
	AltitudeAccuracy int
	LatLongAccuracy  int
	Producer         string
	Area             Area
}

// Area represents SOSI area information
type Area struct {
	MinX, MinY, MaxX, MaxY float64
}

// GenerateHeader creates the SOSI header section exactly matching the JavaScript output
func GenerateHeader(config SOSIConfig, bbox BoundingBox) string {
	var builder strings.Builder

	// Calculate ENHET values as in JavaScript
	sosiUnit := fmt.Sprintf("%.7f", math.Pow(10, -float64(config.LatLongAccuracy)))
	sosiHeightUnit := fmt.Sprintf("%.3f", math.Pow(10, -float64(config.AltitudeAccuracy)))

	builder.WriteString(".HODE 0:\n")
	builder.WriteString("..TEGNSETT UTF-8\n")
	builder.WriteString("..TRANSPAR\n")
	fmt.Fprintf(&builder, "...KOORDSYS %d\n", config.CoordSystem)
	builder.WriteString("...ORIGO-NØ 0 0\n")
	fmt.Fprintf(&builder, "...ENHET %s\n", sosiUnit)
	fmt.Fprintf(&builder, "...ENHET-H %s\n", sosiHeightUnit)
	fmt.Fprintf(&builder, "...ENHET-D %s\n", sosiHeightUnit)
	fmt.Fprintf(&builder, "..PRODUSENT \"%s\"\n", config.Producer)
	fmt.Fprintf(&builder, "..SOSI-VERSJON %s\n", config.Version)
	fmt.Fprintf(&builder, "..SOSI-NIVÅ %d\n", config.Level)
	builder.WriteString("..OMRÅDE\n")

	// Format bounding box as integers (floor for min, ceil for max)
	minLat := int(math.Floor(bbox.MinLat))
	minLon := int(math.Floor(bbox.MinLon))
	maxLat := int(math.Ceil(bbox.MaxLat))
	maxLon := int(math.Ceil(bbox.MaxLon))

	fmt.Fprintf(&builder, "...MIN-NØ %d %d\n", minLat, minLon)
	fmt.Fprintf(&builder, "...MAX-NØ %d %d\n", maxLat, maxLon)

	return builder.String()
}
