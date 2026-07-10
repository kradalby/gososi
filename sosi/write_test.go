package sosi

import (
	"strings"
	"testing"
)

const twoPointLine = `{
  "type": "FeatureCollection",
  "features": [{
    "type": "Feature",
    "id": "f1",
    "geometry": {"type": "LineString", "coordinates": [
      [10.921292458550036, 63.4856654467292, 0.0],
      [10.920455609197402, 63.486702373293724, 0.0]
    ]},
    "properties": {"objtype": "TeleFibertrase", "status": "planned"}
  }]
}`

func convert(t *testing.T, opts ...Option) string {
	t.Helper()
	out, err := GeoJSONToSOSI([]byte(twoPointLine), map[string]string{"f1": "TeleFibertrase"}, opts...)
	if err != nil {
		t.Fatalf("GeoJSONToSOSI: %v", err)
	}
	return string(out)
}

func TestWriterDefaultsAreCompliant(t *testing.T) {
	out := convert(t)
	for _, want := range []string{
		".HODE\n",
		"..SOSI-VERSJON 5.0",
		"..SOSI-NIVÅ 4",
		"...VERT-DATUM NN2000",
		"...KOORDSYS 22", // auto-selected UTM zone 32 (code 22) for ~10.9°E, 63.5°N
		"...ENHET 0.010000",
		"..OBJTYPE TeleFibertrase",
		"..status planned", // property emitted as attribute
		"..id f1",
		".SLUTT\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("compliant default output missing %q\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, ".HODE 0:") {
		t.Error("header should not carry the legacy \".HODE 0:\" form")
	}
}

func TestWriterOptionsReachLegacyBehaviour(t *testing.T) {
	// The pre-existing capabilities (raw lon/lat, custom producer/version/level,
	// no attributes) remain reachable through typed options.
	out := convert(
		t,
		WithKoordSys(KoordSysWGS84Geographic),
		WithAccuracy(7, 3),
		WithVertDatum(VertDatumNone),
		WithProducer("GeoJSONtoSOSI"),
		WithSOSIVersion(SOSIVersion40),
		WithSOSILevel(SOSILevel2),
		WithAttributes(false),
		WithFeatureID(false),
	)
	for _, want := range []string{
		"...KOORDSYS 84",
		"...ENHET 0.0000001",
		"..PRODUSENT \"GeoJSONtoSOSI\"",
		"..SOSI-VERSJON 4.0",
		"..SOSI-NIVÅ 2",
		"634856654 109212925 0000", // raw lat/lon, unprojected
	} {
		if !strings.Contains(out, want) {
			t.Errorf("legacy-mode output missing %q\n---\n%s", want, out)
		}
	}
	for _, unwanted := range []string{"VERT-DATUM", "..status", "..id "} {
		if strings.Contains(out, unwanted) {
			t.Errorf("legacy-mode output should not contain %q", unwanted)
		}
	}
}

func TestWriterExplicitKoordSysProjects(t *testing.T) {
	// KoordSysUTM33 is SOSI code 23; the directive carries the code, not the zone.
	out := convert(t, WithKoordSys(KoordSysUTM33))
	if !strings.Contains(out, "...KOORDSYS 23") {
		t.Errorf("explicit KOORDSYS 23 (UTM33) not honoured\n%s", out)
	}
}

func TestWriterMissingObjectType(t *testing.T) {
	_, err := GeoJSONToSOSI([]byte(twoPointLine), map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing object type")
	}
}
