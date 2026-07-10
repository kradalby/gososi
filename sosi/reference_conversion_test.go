package sosi

import (
	"bufio"
	"os"
	"strings"
	"testing"
)

// TestReferenceConversion is an end-to-end test that runs GeoJSONToSOSI against
// a known-good reference conversion of the same input
// (testdata/reference_pipe.geojson -> testdata/reference_pipe.sos, produced by
// an external authoritative tool) and asserts our output matches it: projected
// coordinates (KOORDSYS 22 / UTM32), cm ENHET, VERT-DATUM, and GeoJSON
// properties emitted as attributes.
//
// The reference is SOSI 3.4; everything else (auto-selected zone 22, NN2000, cm
// precision, attributes) comes from the government-compliant defaults.
func TestReferenceConversion(t *testing.T) {
	inputGeoJSON, err := os.ReadFile("testdata/reference_pipe.geojson")
	if err != nil {
		t.Fatalf("reading input geojson: %v", err)
	}
	reference, err := os.ReadFile("testdata/reference_pipe.sos")
	if err != nil {
		t.Fatalf("reading reference sosi: %v", err)
	}

	// The single feature has id "tube-56" and properties.objtype "Roer".
	got, err := GeoJSONToSOSI(inputGeoJSON, map[string]string{"tube-56": "Roer"},
		WithSOSIVersion(SOSIVersion34))
	if err != nil {
		t.Fatalf("GeoJSONToSOSI: %v", err)
	}

	gotDir := directives(string(got))
	wantDir := directives(string(reference))

	t.Run("Header", func(t *testing.T) {
		// key -> the value the reference conversion uses.
		for _, tc := range []struct{ key, want string }{
			{"KOORDSYS", "22"},
			{"ENHET", "0.010000"},
			{"VERT-DATUM", "NN2000"},
			{"SOSI-VERSJON", "3.4"},
			{"SOSI-NIVÅ", "4"},
		} {
			g := gotDir[tc.key]
			if g != tc.want {
				t.Errorf("%s = %q, reference = %q", tc.key, g, tc.want)
			}
			if w := wantDir[tc.key]; w != tc.want {
				t.Fatalf("reference fixture changed: %s = %q, expected %q", tc.key, w, tc.want)
			}
		}
	})

	t.Run("FeatureAttributes", func(t *testing.T) {
		// Reference emits GeoJSON properties as SOSI attributes on the feature.
		for _, attr := range []string{"diameter_mm", "id"} {
			if _, ok := wantDir[attr]; !ok {
				t.Fatalf("reference fixture changed: expected attribute %q", attr)
			}
			if _, ok := gotDir[attr]; !ok {
				t.Errorf("our output drops feature attribute %q (reference = %q)", attr, wantDir[attr])
			}
		}
	})

	t.Run("Coordinates", func(t *testing.T) {
		gotCoords := coordinateLines(string(got))
		wantCoords := coordinateLines(string(reference))

		if len(gotCoords) == 0 || len(wantCoords) == 0 {
			t.Fatalf("no coordinates extracted (got %d, reference %d)", len(gotCoords), len(wantCoords))
		}

		// Proves the projection gap: our first coordinate is raw lon/lat, the
		// reference's is projected metres in KOORDSYS 22.
		if gotCoords[0] != wantCoords[0] {
			t.Errorf("first coordinate = %q, reference = %q (no reprojection to KOORDSYS %s)",
				gotCoords[0], wantCoords[0], wantDir["KOORDSYS"])
		}

		// Point count also differs: the reference dropped a few coincident points.
		if len(gotCoords) != len(wantCoords) {
			t.Errorf("coordinate count = %d, reference = %d", len(gotCoords), len(wantCoords))
		}
	})
}

// directives parses SOSI directive lines ("..KEY value", "...KEY value") into a
// map of KEY -> value (last write wins). Lines without a value map to "".
func directives(sosi string) map[string]string {
	out := make(map[string]string)
	sc := bufio.NewScanner(strings.NewReader(sosi))
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), " \t\r")
		trimmed := strings.TrimLeft(line, ".")
		if trimmed == line || trimmed == "" {
			continue // not a directive line, or empty
		}
		key, value, _ := strings.Cut(trimmed, " ")
		out[key] = strings.TrimSpace(value)
	}
	return out
}

// coordinateLines returns the NØH coordinate rows (lines beginning with a digit).
func coordinateLines(sosi string) []string {
	var out []string
	sc := bufio.NewScanner(strings.NewReader(sosi))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" && line[0] >= '0' && line[0] <= '9' {
			out = append(out, line)
		}
	}
	return out
}
