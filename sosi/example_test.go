package sosi_test

import (
	"fmt"

	"github.com/kradalby/gososi/sosi"
)

// ExampleSOSIBuilder demonstrates how to use the SOSI builder to create SOSI format output
func ExampleSOSIBuilder() {
	// Create a builder with default configuration
	config := sosi.DefaultConfig()
	builder := sosi.NewBuilder(config)

	// Add a point feature
	pointFeature := sosi.SOSIFeature{
		ID:         1,
		Type:       "PUNKT",
		ObjectType: "Fordelingsskap",
		Coordinates: []sosi.Coordinate{
			sosi.ConvertCoordinate(63.4856654467292, 10.921292458550036, 0.0),
		},
	}
	builder.AddFeature(pointFeature)

	// Add a line feature
	lineFeature := sosi.SOSIFeature{
		ID:         2,
		Type:       "KURVE",
		ObjectType: "TeleFibertrase",
		Coordinates: []sosi.Coordinate{
			sosi.ConvertCoordinate(63.48565826234929, 10.92128709413111, 0.0),
			sosi.ConvertCoordinate(63.48581871307087, 10.921147619239004, 0.0),
		},
	}
	builder.AddFeature(lineFeature)

	// Build the final SOSI output
	sosiOutput, err := builder.Build()
	if err != nil {
		fmt.Printf("Error building SOSI: %v\n", err)
		return
	}

	fmt.Print(sosiOutput)
	// Output:
	// .HODE 0:
	// ..TEGNSETT UTF-8
	// ..TRANSPAR
	// ...KOORDSYS 84
	// ...ORIGO-NØ 0 0
	// ...ENHET 0.0000001
	// ...ENHET-H 0.001
	// ...ENHET-D 0.001
	// ..PRODUSENT "GeoJSONtoSOSI"
	// ..SOSI-VERSJON 4.0
	// ..SOSI-NIVÅ 2
	// ..OMRÅDE
	// ...MIN-NØ 63 10
	// ...MAX-NØ 64 11
	// .PUNKT 1:
	// ..OBJTYPE Fordelingsskap
	// ..NØH
	// 634856654 109212925 0000
	// .KURVE 2:
	// ..OBJTYPE TeleFibertrase
	// ..NØH
	// 634856583 109212871 0000
	// 634858187 109211476 0000
	//
	// .SLUTT
}

// ExampleTransformCoordinate demonstrates coordinate transformation from GeoJSON to SOSI format
func ExampleTransformCoordinate() {
	config := sosi.DefaultConfig()

	// Convert from GeoJSON coordinate format (lat, lon, alt) to internal format
	coord := sosi.ConvertCoordinate(63.4856654467292, 10.921292458550036, 0.0)

	// Transform to SOSI format: "latitude longitude altitude" (with decimal points removed)
	sosiCoord := sosi.TransformCoordinate(coord, config)
	fmt.Println(sosiCoord)
	// Output: 634856654 109212925 0000
}

// ExampleConvertGeometryTypeToSOSI shows how GeoJSON geometry types map to SOSI types
func ExampleConvertGeometryTypeToSOSI() {
	types := []string{"Point", "LineString", "MultiPoint", "Polygon"}

	for _, geoJSONType := range types {
		sosiType, err := sosi.ConvertGeometryTypeToSOSI(geoJSONType)
		if err != nil {
			fmt.Printf("%s -> Error: %v\n", geoJSONType, err)
		} else {
			fmt.Printf("%s -> %s\n", geoJSONType, sosiType)
		}
	}
	// Output:
	// Point -> PUNKT
	// LineString -> KURVE
	// MultiPoint -> SVERM
	// Polygon -> FLATE
}

// ExampleGenerateHeader shows how to generate just the SOSI header
func ExampleGenerateHeader() {
	config := sosi.SOSIConfig{
		Producer:         "MyApp",
		AltitudeAccuracy: 3,
		LatLongAccuracy:  7,
		CoordSystem:      25832, // UTM Zone 32N
		Version:          "4.5",
		Level:            3,
	}

	bbox := sosi.BoundingBox{
		MinLat: 60.0,
		MinLon: 5.0,
		MaxLat: 61.0,
		MaxLon: 6.0,
	}

	header := sosi.GenerateHeader(config, bbox)
	fmt.Print(header)
	// Output:
	// .HODE 0:
	// ..TEGNSETT UTF-8
	// ..TRANSPAR
	// ...KOORDSYS 25832
	// ...ORIGO-NØ 0 0
	// ...ENHET 0.0000001
	// ...ENHET-H 0.001
	// ...ENHET-D 0.001
	// ..PRODUSENT "MyApp"
	// ..SOSI-VERSJON 4.5
	// ..SOSI-NIVÅ 3
	// ..OMRÅDE
	// ...MIN-NØ 60 5
	// ...MAX-NØ 61 6
}
