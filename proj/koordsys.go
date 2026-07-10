package proj

import "math"

// KoordSys is a Norwegian SOSI coordinate-system code (the value of the
// ...KOORDSYS header directive).
type KoordSys int

// Recognised KOORDSYS codes.
//
// The 21-26 block is EUREF89/WGS84 UTM (zones 31-36), which is what modern SOSI
// deliveries use and what this package projects. KoordSysWGS84Geographic (84)
// is raw geographic longitude/latitude — no projection is applied.
const (
	KoordSysUTM31 KoordSys = 21
	KoordSysUTM32 KoordSys = 22
	KoordSysUTM33 KoordSys = 23
	KoordSysUTM34 KoordSys = 24
	KoordSysUTM35 KoordSys = 25
	KoordSysUTM36 KoordSys = 26

	KoordSysWGS84Geographic KoordSys = 84
)

// IsProjected reports whether the code denotes a projected system that requires
// reprojection (as opposed to raw geographic lon/lat).
func (k KoordSys) IsProjected() bool {
	_, ok := utmZoneForKoordSys[k]
	return ok
}

// utmZoneForKoordSys maps the supported EUREF89/WGS84 UTM codes to their zone
// number. Note code 26 -> zone 36 (EPSG:32636), matching the SRID in the SOSI
// code table.
var utmZoneForKoordSys = map[KoordSys]int{
	KoordSysUTM31: 31,
	KoordSysUTM32: 32,
	KoordSysUTM33: 33,
	KoordSysUTM34: 34,
	KoordSysUTM35: 35,
	KoordSysUTM36: 36,
}

// legacyDatumCodes are recognised SOSI codes whose datum needs a Helmert
// transform that is not yet implemented: NGO1948 (1-9) and ED50 (31-36).
func isLegacyDatumCode(code KoordSys) bool {
	if code >= 1 && code <= 9 {
		return true // NGO1948
	}
	if code >= 31 && code <= 36 {
		return true // ED50
	}
	return false
}

// FromKOORDSYS returns the projection for a SOSI KOORDSYS code.
//
// It supports the EUREF89/WGS84 UTM codes 21-26 (all northern hemisphere).
// The zone is taken from the code, never re-derived from longitude, because
// Norwegian KOORDSYS-22/23 data legitimately extends beyond the nominal 6° zone
// band. Legacy-datum codes return ErrUnsupportedDatum; anything else returns
// ErrUnknownKOORDSYS.
func FromKOORDSYS(code KoordSys) (CRS, error) {
	if zone, ok := utmZoneForKoordSys[code]; ok {
		return UTM(zone, true), nil
	}
	if isLegacyDatumCode(code) {
		return nil, ErrUnsupportedDatum
	}
	return nil, ErrUnknownKOORDSYS
}

// UTMZoneForLon returns the nominal UTM zone (1-60) for a longitude in degrees,
// applying the Norwegian/Svalbard exceptions around zone 32V and 31-37. It is
// intended for auto-selecting a zone when the caller has no explicit KOORDSYS;
// projection by an explicit code must use FromKOORDSYS instead.
//
// lat is needed only to apply the latitude-banded exceptions; pass the data
// centroid.
func UTMZoneForLon(lon, lat float64) int {
	// Normalise longitude to [-180, 180).
	lon = math.Mod(lon+180, 360)
	if lon < 0 {
		lon += 360
	}
	lon -= 180

	zone := int((lon+180)/6) + 1

	// Southwest Norway: zone 32 is widened west to 3°E over 56°-64°N.
	if lat >= 56 && lat < 64 && lon >= 3 && lon < 12 {
		zone = 32
	}

	// Svalbard: 74°-84°N uses only odd-numbered widened zones.
	if lat >= 72 && lat < 84 {
		switch {
		case lon >= 0 && lon < 9:
			zone = 31
		case lon >= 9 && lon < 21:
			zone = 33
		case lon >= 21 && lon < 33:
			zone = 35
		case lon >= 33 && lon < 42:
			zone = 37
		}
	}

	return zone
}

// KoordSysForZone returns the EUREF89/WGS84 KoordSys code for a northern-
// hemisphere UTM zone, and whether such a code exists in the supported 21-26
// range.
func KoordSysForZone(zone int) (KoordSys, bool) {
	for code, z := range utmZoneForKoordSys {
		if z == zone {
			return code, true
		}
	}
	return 0, false
}
