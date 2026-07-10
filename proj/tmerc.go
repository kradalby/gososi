package proj

import "math"

// TransverseMercator is a transverse-Mercator projection on a given ellipsoid.
//
// It implements the Karney-Krüger n-series to sixth order. This is the same
// method used by GeographicLib, PROJ's etmerc and Kartverket, accurate to a few
// nanometres within a few degrees of the central meridian and to sub-millimetre
// well beyond a UTM zone's nominal width.
type TransverseMercator struct {
	Ell    Ellipsoid
	Lon0   float64 // central meridian, degrees
	K0     float64 // scale factor on the central meridian
	FalseE float64 // false easting, metres
	FalseN float64 // false northing, metres

	// Precomputed series coefficients (derived from Ell in NewTransverseMercator).
	n     float64
	bigA  float64    // rectifying radius * k0-independent factor
	alpha [7]float64 // forward coefficients, index 1..6
	beta  [7]float64 // inverse coefficients, index 1..6
	delta [7]float64 // conformal->geodetic latitude coefficients, index 1..6
}

// NewTransverseMercator builds a transverse-Mercator projection and precomputes
// its series coefficients.
func NewTransverseMercator(ell Ellipsoid, lon0, k0, falseE, falseN float64) *TransverseMercator {
	t := &TransverseMercator{
		Ell:    ell,
		Lon0:   lon0,
		K0:     k0,
		FalseE: falseE,
		FalseN: falseN,
	}

	n := ell.thirdFlattening()
	t.n = n
	n2 := n * n
	n3 := n2 * n
	n4 := n3 * n
	n5 := n4 * n
	n6 := n5 * n

	// Rectifying radius factor A = a/(1+n) * (1 + n^2/4 + n^4/64 + n^6/256).
	t.bigA = ell.A / (1 + n) * (1 + n2/4 + n4/64 + n6/256)

	// Forward (Gauss-Krüger) coefficients, to n^6.
	t.alpha[1] = n/2 - 2*n2/3 + 5*n3/16 + 41*n4/180 - 127*n5/288 + 7891*n6/37800
	t.alpha[2] = 13*n2/48 - 3*n3/5 + 557*n4/1440 + 281*n5/630 - 1983433*n6/1935360
	t.alpha[3] = 61*n3/240 - 103*n4/140 + 15061*n5/26880 + 167603*n6/181440
	t.alpha[4] = 49561*n4/161280 - 179*n5/168 + 6601661*n6/7257600
	t.alpha[5] = 34729*n5/80640 - 3418889*n6/1995840
	t.alpha[6] = 212378941 * n6 / 319334400

	// Inverse coefficients, to n^6.
	t.beta[1] = n/2 - 2*n2/3 + 37*n3/96 - n4/360 - 81*n5/512 + 96199*n6/604800
	t.beta[2] = n2/48 + n3/15 - 437*n4/1440 + 46*n5/105 - 1118711*n6/3870720
	t.beta[3] = 17*n3/480 - 37*n4/840 - 209*n5/4480 + 5569*n6/90720
	t.beta[4] = 4397*n4/161280 - 11*n5/504 - 830251*n6/7257600
	t.beta[5] = 4583*n5/161280 - 108847*n6/3991680
	t.beta[6] = 20648693 * n6 / 638668800

	// Conformal -> geodetic latitude coefficients, to n^6.
	t.delta[1] = 2*n - 2*n2/3 - 2*n3 + 116*n4/45 + 26*n5/45 - 2854*n6/675
	t.delta[2] = 7*n2/3 - 8*n3/5 - 227*n4/45 + 2704*n5/315 + 2323*n6/945
	t.delta[3] = 56*n3/15 - 136*n4/35 - 1262*n5/105 + 73814*n6/2835
	t.delta[4] = 4279*n4/630 - 332*n5/35 - 399572*n6/14175
	t.delta[5] = 4174*n5/315 - 144838*n6/6237
	t.delta[6] = 601676 * n6 / 22275

	return t
}

// Forward projects geographic (lon, lat) in degrees to (easting, northing) in
// metres.
func (t *TransverseMercator) Forward(lon, lat float64) (easting, northing float64) {
	phi := lat * math.Pi / 180
	// Longitude relative to the central meridian, normalised to (-180, 180].
	dl := math.Mod(lon-t.Lon0, 360)
	if dl > 180 {
		dl -= 360
	} else if dl < -180 {
		dl += 360
	}
	lambda := dl * math.Pi / 180

	// Conformal latitude via the isometric latitude.
	e2n := 2 * math.Sqrt(t.n) / (1 + t.n)
	tau := math.Sinh(math.Atanh(math.Sin(phi)) - e2n*math.Atanh(e2n*math.Sin(phi)))

	xiP := math.Atan2(tau, math.Cos(lambda))
	etaP := math.Asinh(math.Sin(lambda) / math.Hypot(tau, math.Cos(lambda)))

	xi := xiP
	eta := etaP
	for j := 1; j <= 6; j++ {
		jj := float64(2 * j)
		xi += t.alpha[j] * math.Sin(jj*xiP) * math.Cosh(jj*etaP)
		eta += t.alpha[j] * math.Cos(jj*xiP) * math.Sinh(jj*etaP)
	}

	easting = t.FalseE + t.K0*t.bigA*eta
	northing = t.FalseN + t.K0*t.bigA*xi
	return easting, northing
}

// Inverse projects (easting, northing) in metres back to geographic (lon, lat)
// in degrees.
func (t *TransverseMercator) Inverse(easting, northing float64) (lon, lat float64) {
	xi := (northing - t.FalseN) / (t.K0 * t.bigA)
	eta := (easting - t.FalseE) / (t.K0 * t.bigA)

	xiP := xi
	etaP := eta
	for j := 1; j <= 6; j++ {
		jj := float64(2 * j)
		xiP -= t.beta[j] * math.Sin(jj*xi) * math.Cosh(jj*eta)
		etaP -= t.beta[j] * math.Cos(jj*xi) * math.Sinh(jj*eta)
	}

	// Conformal latitude, then geodetic latitude via the delta series.
	chi := math.Asin(math.Sin(xiP) / math.Cosh(etaP))
	phi := chi
	for j := 1; j <= 6; j++ {
		phi += t.delta[j] * math.Sin(float64(2*j)*chi)
	}

	lambda := math.Atan2(math.Sinh(etaP), math.Cos(xiP))

	lat = phi * 180 / math.Pi
	lon = t.Lon0 + lambda*180/math.Pi
	return lon, lat
}
