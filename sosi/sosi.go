// Package sosi provides SOSI format writing functionality
package sosi

import "fmt"

// SOSIConfig defines the configuration parameters for SOSI format conversion
type SOSIConfig struct {
	Producer         string // Producer identification
	AltitudeAccuracy int    // Number of decimal places for altitude (3)
	LatLongAccuracy  int    // Number of decimal places for lat/lon (7)
	CoordSystem      int    // Coordinate system (84 for WGS84)
	Version          string // SOSI version (4.0)
	Level            int    // SOSI level (2)

	// Performance configuration
	Workers           int  // Number of concurrent workers (0 = auto-detect)
	BufferSize        int  // Buffer size for string building (0 = default)
	EnableConcurrency bool // Enable concurrent processing
}

// SOSIFeature represents a SOSI feature with its properties and coordinates
// Extended to support both writing (existing) and parsing (new) functionality
type SOSIFeature struct {
	ID          int                    // Feature ID
	Type        string                 // PUNKT, KURVE, FLATE, BUEP, SVERM
	ObjectType  string                 // User-defined object type
	Coordinates []Coordinate           // Feature coordinates
	Properties  map[string]interface{} // Parsed attributes from SOSI (new for parsing)
	Refs        []int                  // Feature references for FLATE geometries (new for parsing)
	OuterRing   []int                  // Outer ring references for FLATE with holes (new for parsing)
	Holes       [][]int                // Inner ring (hole) references for FLATE geometries (new for parsing)
}

// BoundingBox represents the extent of a SOSI dataset
type BoundingBox struct {
	MinLat, MinLon, MaxLat, MaxLon float64
}

// ConversionError represents errors that occur during GeoJSON to SOSI conversion
type ConversionError struct {
	Type    string // Error type classification
	Message string // Human readable error message
	Err     error  // Underlying error, if any
}

// Error implements the error interface for ConversionError
func (e *ConversionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error for error wrapping compatibility
func (e *ConversionError) Unwrap() error {
	return e.Err
}

// DefaultConfig returns a default SOSIConfig that matches the JavaScript implementation
func DefaultConfig() SOSIConfig {
	return SOSIConfig{
		Producer:         "GeoJSONtoSOSI",
		AltitudeAccuracy: 3,
		LatLongAccuracy:  7,
		CoordSystem:      84,
		Version:          "4.0",
		Level:            2,

		// Performance defaults
		Workers:           0,     // Auto-detect based on CPU cores
		BufferSize:        1024,  // 1KB default buffer
		EnableConcurrency: false, // Disabled by default for backward compatibility
	}
}

// Writer handles the generation of SOSI format output
// Deprecated: Use SOSIBuilder instead for better composability and exact format matching
type Writer struct {
	config  SOSIConfig
	builder *SOSIBuilder
}

// NewWriter creates a new SOSI writer with the given configuration
// Deprecated: Use NewBuilder instead for better composability
func NewWriter(config SOSIConfig) *Writer {
	return &Writer{
		config:  config,
		builder: NewBuilder(config),
	}
}

// NewWriterWithDefaults creates a new SOSI writer with default configuration
// Deprecated: Use NewBuilder(DefaultConfig()) instead
func NewWriterWithDefaults() *Writer {
	return NewWriter(DefaultConfig())
}

// WriteHeader generates the SOSI header section
// Deprecated: Use GenerateHeader function instead
func (w *Writer) WriteHeader(bbox BoundingBox) string {
	return GenerateHeader(w.config, bbox)
}

// WriteFooter generates the SOSI footer section
// Deprecated: Footer is automatically included in SOSIBuilder.Build()
func (w *Writer) WriteFooter() string {
	return ".SLUTT\n"
}

// AddFeature adds a feature to the writer
func (w *Writer) AddFeature(feature SOSIFeature) {
	w.builder.AddFeature(feature)
}

// Build generates the complete SOSI output
func (w *Writer) Build() (string, error) {
	return w.builder.Build()
}

// SOSIDocument represents a complete parsed SOSI document
// This is the root type returned by the parser
type SOSIDocument struct {
	Header   SOSIHeader    // Parsed header information
	Features []SOSIFeature // All parsed features
	Bounds   BoundingBox   // Document bounding box
}

// SOSIHeader represents parsed SOSI header information (HODE section)
type SOSIHeader struct {
	CharacterSet     string                 // TEGNSETT (e.g., "UTF-8")
	Producer         string                 // PRODUSENT
	Version          string                 // SOSI-VERSJON (e.g., "4.5")
	Level            int                    // SOSI-NIVÅ
	CoordSystem      int                    // KOORDSYS (e.g., 84 for WGS84)
	GeosysParams     []string               // GEOSYS parameters if present
	Unit             float64                // ENHET coordinate unit
	HeightUnit       float64                // ENHET-H height unit
	DepthUnit        float64                // ENHET-D depth unit
	Origo            Coordinate             // ORIGO-NØ origin point
	Owner            string                 // EIER
	ObjectCatalog    string                 // OBJEKTKATALOG
	VerificationDate string                 // VERIFISERINGSDATO
	Quality          map[string]interface{} // KVALITET information
	Area             BoundingBox            // OMRÅDE (MIN-NØ, MAX-NØ)
}
