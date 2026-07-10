package sosi_test

import (
	"fmt"

	"github.com/kradalby/gososi/sosi"
)

// ExampleGeoJSONToSOSI converts a GeoJSON LineString to government-compliant
// SOSI. With no options the data is reprojected into an auto-selected UTM zone
// (here KOORDSYS 22 / UTM32), heights are referenced to NN2000, and properties
// are emitted as attributes.
func ExampleGeoJSONToSOSI() {
	data := []byte(`{
      "type": "FeatureCollection",
      "features": [{
        "type": "Feature",
        "id": "tube-1",
        "geometry": {"type": "LineString", "coordinates": [
          [10.210588982, 59.171243033, 145.811],
          [10.210562953, 59.171222462, 145.624]
        ]},
        "properties": {"objtype": "Roer", "diameter_mm": 16}
      }]
    }`)

	out, err := sosi.GeoJSONToSOSI(data, map[string]string{"tube-1": "Roer"},
		sosi.WithSOSIVersion(sosi.SOSIVersion34))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(string(out))
	// Output:
	// .HODE
	// ..TEGNSETT UTF-8
	// ..SOSI-VERSJON 3.4
	// ..SOSI-NIVÅ 4
	// ..TRANSPAR
	// ...ENHET 0.010000
	// ...KOORDSYS 22
	// ...VERT-DATUM NN2000
	// ...ORIGO-NØ 0 0
	// ..OMRÅDE
	// ...MIN-NØ 6559745 569201
	// ...MAX-NØ 6559749 569204
	// .KURVE 1:
	// ..OBJTYPE Roer
	// ..diameter_mm 16
	// ..id tube-1
	// ..NØH
	// 655974810 56920304 14581
	// 655974579 56920159 14562
	// .SLUTT
}

// ExampleGeoJSONToSOSI_geographic emits raw WGS84 lon/lat instead of projecting,
// by selecting KoordSysWGS84Geographic.
func ExampleGeoJSONToSOSI_geographic() {
	data := []byte(`{
      "type": "FeatureCollection",
      "features": [{
        "type": "Feature",
        "id": "p1",
        "geometry": {"type": "Point", "coordinates": [10.921292458550036, 63.4856654467292]},
        "properties": {}
      }]
    }`)

	out, err := sosi.GeoJSONToSOSI(
		data, map[string]string{"p1": "Fordelingsskap"},
		sosi.WithKoordSys(sosi.KoordSysWGS84Geographic),
		sosi.WithAccuracy(7, 3),
		sosi.WithVertDatum(sosi.VertDatumNone),
		sosi.WithAttributes(false),
		sosi.WithFeatureID(false),
	)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Print(string(out))
	// Output:
	// .HODE
	// ..TEGNSETT UTF-8
	// ..SOSI-VERSJON 5.0
	// ..SOSI-NIVÅ 4
	// ..TRANSPAR
	// ...ENHET 0.0000001
	// ...ENHET-H 0.001000
	// ...KOORDSYS 84
	// ...ORIGO-NØ 0 0
	// ..OMRÅDE
	// ...MIN-NØ 63 10
	// ...MAX-NØ 64 11
	// .PUNKT 1:
	// ..OBJTYPE Fordelingsskap
	// ..NØH
	// 634856654 109212925 0000
	// .SLUTT
}

// ExampleConvertGeometryTypeToSOSI shows how GeoJSON geometry types map to SOSI types.
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
