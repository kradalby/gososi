package sosi

import (
	"encoding/json"
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

// pointWith converts a single Point feature carrying id and properties.
func pointWith(id, props, objType string, opts ...Option) (string, error) {
	in := `{"type":"FeatureCollection","features":[{"type":"Feature","id":` + id +
		`,"geometry":{"type":"Point","coordinates":[10.7,59.9]},"properties":` + props + `}]}`
	out, err := GeoJSONToSOSI([]byte(in), map[string]string{"a": objType}, opts...)
	return string(out), err
}

func TestWriterAttributeValues(t *testing.T) {
	tests := []struct {
		name  string
		props string
		want  string
	}{
		{"plain token", `{"k":"planned"}`, "..k planned\n"},
		{"space", `{"k":"hello world"}`, "..k \"hello world\"\n"},
		{"double quote", `{"k":"say \"hi\""}`, "..k 'say \"hi\"'\n"},
		{"single quote", `{"k":"it's"}`, "..k \"it's\"\n"},
		{"empty", `{"k":""}`, "..k \"\"\n"},
		{"leading dot", `{"k":".SLUTT"}`, "..k \".SLUTT\"\n"},
		{"comment marker", `{"k":"a!b"}`, "..k \"a!b\"\n"},
		{"colon", `{"k":":5"}`, "..k \":5\"\n"},
		{"integer", `{"k":40}`, "..k 40\n"},
		{"fraction", `{"k":0.5}`, "..k 0.5\n"},
		{"beyond int64", `{"k":1e21}`, "..k 1000000000000000000000\n"},
		{"bool", `{"k":true}`, "..k true\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := pointWith(`"a"`, tt.props, "Kum")
			if err != nil {
				t.Fatalf("GeoJSONToSOSI: %v", err)
			}
			if !strings.Contains(out, tt.want) {
				t.Errorf("output missing %q\n---\n%s", tt.want, out)
			}
		})
	}
}

func TestWriterOmitsNullAndGeometryAttributes(t *testing.T) {
	out, err := pointWith(`"a"`, `{"gone":null,"NØH":"1 2 3","NØ":"1 2","REF":":1","OBJTYPE":"X"}`, "Kum")
	if err != nil {
		t.Fatalf("GeoJSONToSOSI: %v", err)
	}
	for _, unwanted := range []string{"..gone", "..NØ ", "..NØH ", "..REF", "..OBJTYPE X"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("output should not contain %q\n---\n%s", unwanted, out)
		}
	}
	if n := strings.Count(out, "..NØH"); n != 1 {
		t.Errorf("want one ..NØH block, got %d\n---\n%s", n, out)
	}
}

func TestWriterRejectsUnrepresentable(t *testing.T) {
	tests := []struct {
		name    string
		id      string // JSON; also the objectTypes key
		props   string
		objType string
		opts    []Option
	}{
		{"newline in value", `"a"`, `{"k":"hi\n.SLUTT\n.PUNKT 99:"}`, "Kum", nil},
		{"carriage return in value", `"a"`, `{"k":"a\rb"}`, "Kum", nil},
		{"tab in value", `"a"`, `{"k":"a\tb"}`, "Kum", nil},
		{"both quote kinds", `"a"`, `{"k":"'\""}`, "Kum", nil},
		{"nested object", `"a"`, `{"k":{"x":1}}`, "Kum", nil},
		{"array", `"a"`, `{"k":[1,2]}`, "Kum", nil},
		{"newline in key", `"a"`, `{"a\n.SLUTT":1}`, "Kum", nil},
		{"space in key", `"a"`, `{"a b":1}`, "Kum", nil},
		{"leading dot in key", `"a"`, `{".k":1}`, "Kum", nil},
		{"empty key", `"a"`, `{"":1}`, "Kum", nil},
		{"newline in id", `"a\n.SLUTT"`, `{}`, "Kum", nil},
		{"newline in objtype", `"a"`, `{}`, "Kum\n.SLUTT", nil},
		{"space in objtype", `"a"`, `{}`, "Kum Evil", nil},
		{"empty objtype", `"a"`, `{}`, "", nil},
		{"newline in producer", `"a"`, `{}`, "Kum", []Option{WithProducer("x\n.SLUTT")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var key string
			if err := json.Unmarshal([]byte(tt.id), &key); err != nil {
				t.Fatal(err)
			}
			in := `{"type":"FeatureCollection","features":[{"type":"Feature","id":` + tt.id +
				`,"geometry":{"type":"Point","coordinates":[10.7,59.9]},"properties":` + tt.props + `}]}`
			out, err := GeoJSONToSOSI([]byte(in), map[string]string{key: tt.objType}, tt.opts...)
			if err == nil {
				t.Fatalf("want error, got output\n---\n%s", out)
			}
		})
	}
}

func TestWriterAttributesRoundTrip(t *testing.T) {
	out, err := pointWith(`"a"`, `{"note":"x .SLUTT .PUNKT 99"}`, "Kum")
	if err != nil {
		t.Fatalf("GeoJSONToSOSI: %v", err)
	}
	doc, err := NewParser().Parse(strings.NewReader(out))
	if err != nil {
		t.Fatalf("Parse: %v\n---\n%s", err, out)
	}
	if len(doc.Features) != 1 {
		t.Fatalf("want 1 feature, got %d\n---\n%s", len(doc.Features), out)
	}
	if got := doc.Features[0].Properties["note"]; got != "x .SLUTT .PUNKT 99" {
		t.Errorf("note = %q\n---\n%s", got, out)
	}
}
