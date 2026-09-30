package sosi

import "testing"

func TestAnalyzeGeoJSONNullGeometry(t *testing.T) {
	in := `{"type":"FeatureCollection","features":[{"type":"Feature","id":"a","geometry":null,"properties":{"label":"Skap"}}]}`
	got, err := AnalyzeGeoJSON([]byte(in))
	if err != nil {
		t.Fatalf("AnalyzeGeoJSON: %v", err)
	}
	if want := "a: Skap null = 0 coords"; got["a"] != want {
		t.Errorf("analysis = %q, want %q", got["a"], want)
	}
}
