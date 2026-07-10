package proj

// Ellipsoid is a reference ellipsoid defined by its semi-major axis A (metres)
// and flattening F.
type Ellipsoid struct {
	A float64 // semi-major axis, metres
	F float64 // flattening
}

// Reference ellipsoids. WGS84 and GRS80 differ only in flattening, by less than
// 0.1 mm on the ground; both are provided so callers can match a specific datum
// definition exactly.
var (
	WGS84 = Ellipsoid{A: 6378137.0, F: 1.0 / 298.257223563}
	GRS80 = Ellipsoid{A: 6378137.0, F: 1.0 / 298.257222101}
)

// thirdFlattening returns n = f / (2 - f), the parameter the Krüger series is
// expressed in.
func (e Ellipsoid) thirdFlattening() float64 {
	return e.F / (2 - e.F)
}
