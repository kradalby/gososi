package proj

// Ellipsoid is a reference ellipsoid defined by its semi-major axis A (metres)
// and flattening F.
type Ellipsoid struct {
	A float64 // semi-major axis, metres
	F float64 // flattening
}

// WGS84 returns the WGS84 reference ellipsoid. It and GRS80 differ only in
// flattening, by less than 0.1 mm on the ground; both are provided so callers
// can match a specific datum definition exactly. They are functions rather than
// variables so no importer can redefine an ellipsoid for the whole process.
func WGS84() Ellipsoid { return Ellipsoid{A: 6378137.0, F: 1.0 / 298.257223563} }

// GRS80 returns the GRS80 reference ellipsoid.
func GRS80() Ellipsoid { return Ellipsoid{A: 6378137.0, F: 1.0 / 298.257222101} }

// thirdFlattening returns n = f / (2 - f), the parameter the Krüger series is
// expressed in.
func (e Ellipsoid) thirdFlattening() float64 {
	return e.F / (2 - e.F)
}
