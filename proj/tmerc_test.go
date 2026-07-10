package proj

import (
	"math"
	"testing"
)

// TestFromKOORDSYS22Reference checks the two endpoints of the reference pipe
// conversion (sosi/testdata/reference_pipe) against ground truth, to the
// centimetre. This is the primary correctness gate for the Krüger series.
func TestFromKOORDSYS22Reference(t *testing.T) {
	crs, err := FromKOORDSYS(KoordSysUTM32)
	if err != nil {
		t.Fatalf("FromKOORDSYS(22): %v", err)
	}

	tests := []struct {
		name         string
		lon, lat     float64
		wantE, wantN float64
		tol          float64
	}{
		{"first vertex", 10.210588982, 59.171243033, 569203.04, 6559748.10, 0.01},
		{"last vertex", 10.199476078333333, 59.15564586333333, 568599.04, 6558000.01, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, n := crs.Forward(tt.lon, tt.lat)
			if math.Abs(e-tt.wantE) > tt.tol {
				t.Errorf("easting = %.4f, want %.4f (±%.2f)", e, tt.wantE, tt.tol)
			}
			if math.Abs(n-tt.wantN) > tt.tol {
				t.Errorf("northing = %.4f, want %.4f (±%.2f)", n, tt.wantN, tt.tol)
			}
		})
	}
}

// TestRoundTripGrid checks Inverse(Forward(x)) == x to sub-millimetre across a
// grid spanning each supported zone's valid extent.
func TestRoundTripGrid(t *testing.T) {
	for code, zone := range utmZoneForKoordSys {
		crs, err := FromKOORDSYS(code)
		if err != nil {
			t.Fatalf("FromKOORDSYS(%d): %v", code, err)
		}
		lon0 := float64(6*zone - 183)

		for dlon := -6.0; dlon <= 6.0; dlon += 1.0 {
			for lat := 55.0; lat <= 72.0; lat += 1.0 {
				lon := lon0 + dlon
				e, n := crs.Forward(lon, lat)
				gotLon, gotLat := crs.Inverse(e, n)

				// Compare in metres on the ground: ~1e-9 deg ≈ 0.1 mm.
				if math.Abs(gotLon-lon) > 1e-9 || math.Abs(gotLat-lat) > 1e-9 {
					t.Errorf("code %d roundtrip (%.4f,%.4f) -> (%.4f,%.4f) [Δlon=%.2e Δlat=%.2e]",
						code, lon, lat, gotLon, gotLat, gotLon-lon, gotLat-lat)
				}
			}
		}
	}
}

func TestFromKOORDSYSErrors(t *testing.T) {
	if _, err := FromKOORDSYS(84); err != ErrUnknownKOORDSYS {
		t.Errorf("code 84: got %v, want ErrUnknownKOORDSYS", err)
	}
	if _, err := FromKOORDSYS(5); err != ErrUnsupportedDatum {
		t.Errorf("code 5 (NGO1948): got %v, want ErrUnsupportedDatum", err)
	}
	if _, err := FromKOORDSYS(32); err != ErrUnsupportedDatum {
		t.Errorf("code 32 (ED50): got %v, want ErrUnsupportedDatum", err)
	}
	if _, err := FromKOORDSYS(999); err != ErrUnknownKOORDSYS {
		t.Errorf("code 999: got %v, want ErrUnknownKOORDSYS", err)
	}
}
