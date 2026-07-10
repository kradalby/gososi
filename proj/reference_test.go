package proj

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

// referenceVector is a projection sample harvested from an independent
// implementation (see testdata/vectors.json, generated with wroge/wgs84).
type referenceVector struct {
	KoordSys int     `json:"koordsys"`
	Lon      float64 `json:"lon"`
	Lat      float64 `json:"lat"`
	E        float64 `json:"e"`
	N        float64 `json:"n"`
	Source   string  `json:"source"`
}

// TestReferenceVectors cross-checks Forward and Inverse against vectors produced
// by an independent projection library. Agreement to the millimetre over a wide
// grid is strong evidence the Krüger series is implemented correctly. The two
// libraries are separate implementations, so this catches coefficient or
// bookkeeping mistakes that a self-consistent round-trip test cannot.
func TestReferenceVectors(t *testing.T) {
	data, err := os.ReadFile("testdata/vectors.json")
	if err != nil {
		t.Fatalf("reading vectors: %v", err)
	}
	var vectors []referenceVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatalf("parsing vectors: %v", err)
	}
	if len(vectors) == 0 {
		t.Fatal("no reference vectors loaded")
	}

	const tolFwd = 1e-4 // 0.1 mm — the limit of the oracle's printed precision
	const tolInv = 1e-8 // ~1 mm in degrees

	for _, v := range vectors {
		crs, err := FromKOORDSYS(KoordSys(v.KoordSys))
		if err != nil {
			t.Fatalf("FromKOORDSYS(%d): %v", v.KoordSys, err)
		}

		e, n := crs.Forward(v.Lon, v.Lat)
		if math.Abs(e-v.E) > tolFwd || math.Abs(n-v.N) > tolFwd {
			t.Errorf("Forward(%.6f,%.6f) code %d = (%.4f,%.4f), want (%.4f,%.4f) [%s]",
				v.Lon, v.Lat, v.KoordSys, e, n, v.E, v.N, v.Source)
		}

		lon, lat := crs.Inverse(v.E, v.N)
		if math.Abs(lon-v.Lon) > tolInv || math.Abs(lat-v.Lat) > tolInv {
			t.Errorf("Inverse(%.4f,%.4f) code %d = (%.8f,%.8f), want (%.8f,%.8f) [%s]",
				v.E, v.N, v.KoordSys, lon, lat, v.Lon, v.Lat, v.Source)
		}
	}
}
