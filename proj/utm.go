package proj

// UTM constants.
const (
	utmK0      = 0.9996
	utmFalseE  = 500000.0
	utmFalseNS = 10000000.0 // false northing on the southern hemisphere
)

// UTM returns the Universal Transverse Mercator projection for the given zone
// (1-60) on the WGS84 ellipsoid. north selects the northern (false northing 0)
// or southern (false northing 10 000 000 m) hemisphere aspect.
func UTM(zone int, north bool) *TransverseMercator {
	lon0 := float64(6*zone - 183)
	falseN := 0.0
	if !north {
		falseN = utmFalseNS
	}
	return NewTransverseMercator(WGS84(), lon0, utmK0, utmFalseE, falseN)
}
