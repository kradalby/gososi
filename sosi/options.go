package sosi

import "github.com/kradalby/gososi/proj"

// KoordSys is a SOSI coordinate-system code. It is re-exported from the proj
// package so callers configure output without importing proj directly.
type KoordSys = proj.KoordSys

// Supported coordinate systems. The UTM codes reproject WGS84 lon/lat into
// projected metres; KoordSysWGS84Geographic emits raw lon/lat with no
// projection (the legacy behaviour).
const (
	KoordSysUTM31           = proj.KoordSysUTM31
	KoordSysUTM32           = proj.KoordSysUTM32
	KoordSysUTM33           = proj.KoordSysUTM33
	KoordSysUTM34           = proj.KoordSysUTM34
	KoordSysUTM35           = proj.KoordSysUTM35
	KoordSysUTM36           = proj.KoordSysUTM36
	KoordSysWGS84Geographic = proj.KoordSysWGS84Geographic
)

// koordSysAuto is the zero value used internally to mean "auto-select the UTM
// zone from the data". It is not a valid KOORDSYS code.
const koordSysAuto KoordSys = 0

// VertDatum is a SOSI vertical datum (the ...VERT-DATUM directive).
type VertDatum string

const (
	VertDatumNone   VertDatum = "" // omit the directive
	VertDatumNN2000 VertDatum = "NN2000"
	VertDatumNN54   VertDatum = "NN54"
)

// SOSIVersion is a SOSI format version (the ..SOSI-VERSJON directive).
type SOSIVersion string

const (
	SOSIVersion34 SOSIVersion = "3.4"
	SOSIVersion40 SOSIVersion = "4.0"
	SOSIVersion45 SOSIVersion = "4.5"
	SOSIVersion50 SOSIVersion = "5.0"
)

// SOSILevel is a SOSI conformance level (the ..SOSI-NIVÅ directive).
type SOSILevel int

const (
	SOSILevel1 SOSILevel = 1
	SOSILevel2 SOSILevel = 2
	SOSILevel3 SOSILevel = 3
	SOSILevel4 SOSILevel = 4
)

// writeOptions holds resolved output settings. Zero value is not usable; use
// defaultWriteOptions.
type writeOptions struct {
	koordsys  KoordSys // koordSysAuto means auto-select from the data
	vertDatum VertDatum
	version   SOSIVersion
	level     SOSILevel
	producer  string
	latlonAcc int // decimals for planar/lon-lat coordinates (drives ENHET)
	altAcc    int // decimals for height (drives ENHET-H when it differs)

	attributes bool // emit non-objtype properties as ..key value
	featureID  bool // emit the GeoJSON feature id as ..id
}

// defaultWriteOptions is the Norwegian-government-compliant default: projected
// UTM (auto zone), NN2000 heights, current SOSI version, centimetre precision,
// and full attribute emission.
func defaultWriteOptions() writeOptions {
	return writeOptions{
		koordsys:   koordSysAuto,
		vertDatum:  VertDatumNN2000,
		version:    SOSIVersion50,
		level:      SOSILevel4,
		producer:   "",
		latlonAcc:  2,
		altAcc:     2,
		attributes: true,
		featureID:  true,
	}
}

// Option configures GeoJSONToSOSI. Options compose in the style of the
// encoding/json/v2 option functions.
type Option func(*writeOptions)

// WithKoordSys sets the target coordinate system. A UTM code reprojects the
// data; KoordSysWGS84Geographic emits raw lon/lat. When unset, the UTM zone is
// auto-selected from the data.
func WithKoordSys(k KoordSys) Option {
	return func(o *writeOptions) { o.koordsys = k }
}

// WithVertDatum sets the vertical datum. VertDatumNone omits the directive.
func WithVertDatum(v VertDatum) Option {
	return func(o *writeOptions) { o.vertDatum = v }
}

// WithSOSIVersion sets the SOSI format version.
func WithSOSIVersion(v SOSIVersion) Option {
	return func(o *writeOptions) { o.version = v }
}

// WithSOSILevel sets the SOSI conformance level.
func WithSOSILevel(l SOSILevel) Option {
	return func(o *writeOptions) { o.level = l }
}

// WithProducer sets the PRODUSENT string. Empty (the default) omits it.
func WithProducer(p string) Option {
	return func(o *writeOptions) { o.producer = p }
}

// WithAccuracy sets the number of decimal places retained for planar/lon-lat
// coordinates and for height, driving ENHET (and ENHET-H when they differ).
func WithAccuracy(latlon, alt int) Option {
	return func(o *writeOptions) {
		o.latlonAcc = latlon
		o.altAcc = alt
	}
}

// WithAttributes controls whether GeoJSON properties (other than objtype) are
// emitted as SOSI attributes. Default true.
func WithAttributes(on bool) Option {
	return func(o *writeOptions) { o.attributes = on }
}

// WithFeatureID controls whether the GeoJSON feature id is emitted as ..id.
// Default true.
func WithFeatureID(on bool) Option {
	return func(o *writeOptions) { o.featureID = on }
}
