package sosi

import (
	"strings"
	"testing"
)

func TestSOSIBuilder(t *testing.T) {
	config := DefaultConfig()
	builder := NewBuilder(config)

	// Add features from the example test data
	feature1 := SOSIFeature{
		ID:         1,
		Type:       "PUNKT",
		ObjectType: "Fordelingsskap",
		Coordinates: []Coordinate{
			{Y: 63.4856654467292, X: 10.921292458550036, Z: 0.0},
		},
	}

	feature2 := SOSIFeature{
		ID:         2,
		Type:       "PUNKT",
		ObjectType: "Fordelingsskap",
		Coordinates: []Coordinate{
			{Y: 63.486702373293724, X: 10.920455609197402, Z: 0.0},
		},
	}

	feature3 := SOSIFeature{
		ID:         3,
		Type:       "KURVE",
		ObjectType: "TeleFibertrase",
		Coordinates: []Coordinate{
			{Y: 63.48565826234929, X: 10.92128709413111, Z: 0.0},
			{Y: 63.48581871307087, X: 10.921147619239004, Z: 0.0},
			{Y: 63.486024663423365, X: 10.920975957833337, Z: 0.0},
			{Y: 63.486216243814425, X: 10.920825754103376, Z: 0.0},
			{Y: 63.48636232299978, X: 10.920713101305905, Z: 0.0},
			{Y: 63.486539532812685, X: 10.920584355251654, Z: 0.0},
			{Y: 63.48665447913092, X: 10.920493160129892, Z: 0.0},
			{Y: 63.486702373293724, X: 10.920444880359549, Z: 0.0},
		},
	}

	builder.AddFeature(feature1)
	builder.AddFeature(feature2)
	builder.AddFeature(feature3)

	result, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() returned error: %v", err)
	}

	// Check that the result contains expected components
	if !strings.Contains(result, ".HODE 0:") {
		t.Errorf("Result missing header")
	}

	if !strings.Contains(result, "..TEGNSETT UTF-8") {
		t.Errorf("Result missing character set")
	}

	if !strings.Contains(result, "..TRANSPAR") {
		t.Errorf("Result missing TRANSPAR")
	}

	if !strings.Contains(result, "...KOORDSYS 84") {
		t.Errorf("Result missing coordinate system")
	}

	if !strings.Contains(result, "..PRODUSENT \"GeoJSONtoSOSI\"") {
		t.Errorf("Result missing producer")
	}

	if !strings.Contains(result, "..SOSI-VERSJON 4.0") {
		t.Errorf("Result missing SOSI version")
	}

	if !strings.Contains(result, "..SOSI-NIVÅ 2") {
		t.Errorf("Result missing SOSI level")
	}

	if !strings.Contains(result, "..OMRÅDE") {
		t.Errorf("Result missing area section")
	}

	if !strings.Contains(result, "...MIN-NØ 63 10") {
		t.Errorf("Result missing or incorrect MIN-NØ")
	}

	if !strings.Contains(result, "...MAX-NØ 64 11") {
		t.Errorf("Result missing or incorrect MAX-NØ")
	}

	if !strings.Contains(result, ".PUNKT 1:") {
		t.Errorf("Result missing first feature")
	}

	if !strings.Contains(result, "..OBJTYPE Fordelingsskap") {
		t.Errorf("Result missing object type")
	}

	if !strings.Contains(result, "634856654 109212925 0000") {
		t.Errorf("Result missing coordinate transformation")
	}

	if !strings.Contains(result, ".KURVE 3:") {
		t.Errorf("Result missing line feature")
	}

	if !strings.Contains(result, "..OBJTYPE TeleFibertrase") {
		t.Errorf("Result missing line object type")
	}

	if !strings.Contains(result, ".SLUTT") {
		t.Errorf("Result missing footer")
	}

	// Verify feature count
	if builder.GetFeatureCount() != 3 {
		t.Errorf("GetFeatureCount() = %d, want 3", builder.GetFeatureCount())
	}

	// Verify bounding box
	bbox := builder.GetBoundingBox()
	if bbox.MinLat >= bbox.MaxLat || bbox.MinLon >= bbox.MaxLon {
		t.Errorf("Invalid bounding box: %+v", bbox)
	}
}

func TestSOSIBuilder_EmptyFeatures(t *testing.T) {
	config := DefaultConfig()
	builder := NewBuilder(config)

	_, err := builder.Build()
	if err == nil {
		t.Errorf("Build() should return error for empty feature collection")
	}

	convErr, ok := err.(*ConversionError)
	if !ok {
		t.Errorf("Expected ConversionError, got %T", err)
	}

	if convErr.Type != "EmptyFeatureCollection" {
		t.Errorf("Expected EmptyFeatureCollection error, got %s", convErr.Type)
	}
}
