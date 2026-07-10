package proj

// Datum couples an ellipsoid with a Helmert 7-parameter transform to WGS84.
//
// This is scaffolding for future support of legacy Norwegian datums (NGO1948,
// ED50), which need a datum shift before projection. The transform itself is
// not implemented yet; the type exists so the API can grow without breaking.
type Datum struct {
	Ell     Ellipsoid
	ToWGS84 [7]float64 // dx, dy, dz (m); rx, ry, rz (arcsec); ds (ppm)
}
