package sosi

import (
	"math"
	"strings"
)

// SOSIBuilder provides a builder pattern for assembling SOSI format output
type SOSIBuilder struct {
	config   SOSIConfig
	features []SOSIFeature
	bbox     BoundingBox
}

// NewBuilder creates a new SOSI builder with the given configuration
func NewBuilder(config SOSIConfig) *SOSIBuilder {
	return &SOSIBuilder{
		config: config,
		bbox: BoundingBox{
			MinLat: math.Inf(1),  // Positive infinity
			MinLon: math.Inf(1),  // Positive infinity
			MaxLat: math.Inf(-1), // Negative infinity
			MaxLon: math.Inf(-1), // Negative infinity
		},
	}
}

// AddFeature adds a feature to the builder and updates the bounding box
func (b *SOSIBuilder) AddFeature(feature SOSIFeature) {
	b.features = append(b.features, feature)
	UpdateBoundingBox(&b.bbox, feature.Coordinates)
}

// Build assembles all components into the final SOSI format string
func (b *SOSIBuilder) Build() (string, error) {
	if len(b.features) == 0 {
		return "", &ConversionError{
			Type:    "EmptyFeatureCollection",
			Message: "Cannot generate SOSI output: no features provided",
		}
	}

	// Validate bounding box was properly calculated
	if math.IsInf(b.bbox.MinLat, 0) || math.IsInf(b.bbox.MinLon, 0) {
		return "", &ConversionError{
			Type:    "InvalidBoundingBox",
			Message: "Cannot generate SOSI area: no valid coordinates processed",
		}
	}

	var builder strings.Builder

	// Generate header
	header := GenerateHeader(b.config, b.bbox)
	builder.WriteString(header)

	// Generate features
	for _, feature := range b.features {
		featureString := GenerateFeature(feature, b.config)
		builder.WriteString(featureString)
	}

	// Generate footer
	builder.WriteString("\n.SLUTT")

	return builder.String(), nil
}

// GetBoundingBox returns the current bounding box
func (b *SOSIBuilder) GetBoundingBox() BoundingBox {
	return b.bbox
}

// GetFeatureCount returns the number of features added to the builder
func (b *SOSIBuilder) GetFeatureCount() int {
	return len(b.features)
}
