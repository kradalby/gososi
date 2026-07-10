package geojson

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Expected values are recomputed at R=6371000 from each source's fixtures
// (sources use differing Earth radii: orb 6378137, Turf/mapado 6371008.8,
// s2 6371010, Veness 6371000 — never copy their numbers verbatim).
// Central-angle fixtures are radius-independent and anchor the formula itself.
func TestPoint_DistanceHaversine(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Point
		expected float64
		delta    float64
	}{
		{
			// https://github.com/Turfjs/turf/blob/master/packages/turf-distance/test.ts
			// central angle 0.015245501024842149 rad (radius-free gold value)
			name:     "turf gold pair Philadelphia area",
			a:        Point{Lon: -75.343, Lat: 39.984},
			b:        Point{Lon: -75.534, Lat: 39.123},
			expected: 0.015245501024842149 * 6371000,
			delta:    1e-6,
		},
		{
			// https://github.com/mapado/haversine tests/geo_ressources.py
			// central angle 0.061562818679421795 rad
			name:     "mapado Lyon to Paris",
			a:        Point{Lon: 4.8422, Lat: 45.7597},
			b:        Point{Lon: 2.3508, Lat: 48.8567},
			expected: 0.061562818679421795 * 6371000,
			delta:    1e-6,
		},
		{
			// https://github.com/paulmach/orb geo/distance_test.go (normalized
			// from orb's equatorial R=6378137)
			name:     "orb Sheffield to Cambridge",
			a:        Point{Lon: -1.8444, Lat: 53.1506},
			b:        Point{Lon: 0.1406, Lat: 52.2047},
			expected: 170199.139351,
			delta:    1e-5,
		},
		{
			// https://www.movable-type.co.uk/scripts/latlong.html (R=6371000)
			name:     "veness Cambridge to Paris",
			a:        Point{Lon: 0.119, Lat: 52.205},
			b:        Point{Lon: 2.351, Lat: 48.857},
			expected: 404279.163989,
			delta:    0.5,
		},
		{
			// Ed Williams aviation formulary cross-check via movable-type
			name:     "LAX to JFK",
			a:        Point{Lon: -118.4, Lat: 33.95},
			b:        Point{Lon: -73.783333, Lat: 40.633333},
			expected: 3972857.776253,
			delta:    0.5,
		},
		{
			// golang/geo earth/earth_test.go TestLengthFromLatLngs
			// (normalized from s2's R=6371010)
			name:     "s2 southern hemisphere long haul",
			a:        Point{Lon: 25, Lat: -37},
			b:        Point{Lon: -155, Lat: -66},
			expected: 8562009.351631,
			delta:    1e-5,
		},
		{
			name:     "s2 equator crossing antimeridian",
			a:        Point{Lon: 165, Lat: 0},
			b:        Point{Lon: -80, Lat: 0},
			expected: 12787416.564124,
			delta:    1e-5,
		},
		{
			// near-antipodal: haversine term a→1, the numerically nastiest
			// region — atan2 form must stay finite
			name:     "s2 near-antipodal",
			a:        Point{Lon: -127, Lat: 47},
			b:        Point{Lon: 53, Lat: -47},
			expected: 20015086.661762,
			delta:    1e-5,
		},
		{
			// longitudes outside [-180,180] on both sides: only Δlon enters
			// the formula through sin/cos, so no normalization is needed
			name:     "s2 out-of-range longitudes",
			a:        Point{Lon: -180.227156, Lat: 51.961951},
			b:        Point{Lon: 181.126878, Lat: 51.782383},
			expected: 95078.207384,
			delta:    1e-5,
		},
		{
			name:     "antipodal on equator is pi R",
			a:        Point{Lon: 0, Lat: 0},
			b:        Point{Lon: 180, Lat: 0},
			expected: math.Pi * 6371000,
			delta:    1e-5,
		},
		{
			name:     "pole to pole is pi R",
			a:        Point{Lon: 0, Lat: 90},
			b:        Point{Lon: 0, Lat: -90},
			expected: math.Pi * 6371000,
			delta:    1e-5,
		},
		{
			name:     "quarter meridian",
			a:        Point{Lon: 0, Lat: 0},
			b:        Point{Lon: 0, Lat: 90},
			expected: math.Pi * 6371000 / 2,
			delta:    1e-5,
		},
		{
			name:     "one degree longitude at equator",
			a:        Point{Lon: 0, Lat: 0},
			b:        Point{Lon: 1, Lat: 0},
			expected: math.Pi * 6371000 / 180,
			delta:    1e-6,
		},
		{
			name:     "identical points",
			a:        Point{Lon: 10.5, Lat: 59.9},
			b:        Point{Lon: 10.5, Lat: 59.9},
			expected: 0,
			delta:    0,
		},
		{
			name:     "identical points at pole",
			a:        Point{Lon: 0, Lat: 90},
			b:        Point{Lon: 0, Lat: 90},
			expected: 0,
			delta:    0,
		},
		{
			// https://github.com/Turfjs/turf/issues/758 — both are the south
			// pole; cos(lat)=0 must kill the longitude term
			name:     "south pole across antimeridian",
			a:        Point{Lon: -180, Lat: -90},
			b:        Point{Lon: 180, Lat: -90},
			expected: 0,
			delta:    1e-6,
		},
		{
			name:     "small step 1e-5 degrees latitude",
			a:        Point{Lon: 10, Lat: 60},
			b:        Point{Lon: 10, Lat: 60.00001},
			expected: 1.111949,
			delta:    1e-4,
		},
		{
			// centimeter scale: the spherical law of cosines returns exactly 0
			// here (acos argument rounds to 1) — the reason haversine exists
			name:     "centimeter scale precision",
			a:        Point{Lon: 10, Lat: 50.0},
			b:        Point{Lon: 10, Lat: 50.0000001},
			expected: 0.0111195,
			delta:    1e-6,
		},
		{
			// depth is deliberately ignored: 2D haversine (fiber runs in the
			// ground; altitude noise would only inflate the estimate)
			name:     "depth ignored",
			a:        Point{Lon: 10, Lat: 60, Depth: 0},
			b:        Point{Lon: 10, Lat: 60, Depth: 500},
			expected: 0,
			delta:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.DistanceHaversine(tt.b)
			assert.False(t, math.IsNaN(got))
			assert.InDelta(t, tt.expected, got, tt.delta)
			assert.Equal(t, got, tt.b.DistanceHaversine(tt.a), "must be symmetric")
		})
	}
}

// The turf gold pair with lon/lat swapped gives a wildly different result —
// guards against GeoJSON coordinate-order mistakes at call sites.
func TestPoint_DistanceHaversine_LonLatOrder(t *testing.T) {
	correct := Point{Lon: -75.343, Lat: 39.984}.DistanceHaversine(Point{Lon: -75.534, Lat: 39.123})
	swapped := Point{Lon: 39.984, Lat: -75.343}.DistanceHaversine(Point{Lon: 39.123, Lat: -75.534})
	assert.InDelta(t, 97129.087029, correct, 1e-5)
	assert.Greater(t, math.Abs(swapped-correct), 1000.0)
}

func TestPoint_DistanceHaversine_AntimeridianSymmetry(t *testing.T) {
	// https://github.com/paulmach/orb geo/distance_test.go: 1° of longitude
	// across the antimeridian must equal 1° across Greenwich
	atGreenwich := Point{Lon: 0.5, Lat: 30}.DistanceHaversine(Point{Lon: -0.5, Lat: 30})
	atAntimeridian := Point{Lon: 179.5, Lat: 30}.DistanceHaversine(Point{Lon: -179.5, Lat: 30})
	assert.InDelta(t, atGreenwich, atAntimeridian, 1e-6)
	assert.InDelta(t, 96297.325678, atGreenwich, 1e-5)
}

func TestLineString_LengthHaversine(t *testing.T) {
	a := Point{Lon: 10.0, Lat: 59.0}
	b := Point{Lon: 10.1, Lat: 59.05}
	c := Point{Lon: 10.2, Lat: 59.02}

	tests := []struct {
		name     string
		ls       LineString
		expected float64
		delta    float64
	}{
		{
			name:     "empty",
			ls:       LineString{},
			expected: 0,
			delta:    0,
		},
		{
			name:     "single point",
			ls:       LineString{a},
			expected: 0,
			delta:    0,
		},
		{
			name:     "two points equals pairwise distance",
			ls:       LineString{a, b},
			expected: a.DistanceHaversine(b),
			delta:    0,
		},
		{
			name:     "multi segment equals sum of segments",
			ls:       LineString{a, b, c},
			expected: a.DistanceHaversine(b) + b.DistanceHaversine(c),
			delta:    1e-9,
		},
		{
			name:     "repeated vertex adds nothing",
			ls:       LineString{a, b, b, c},
			expected: a.DistanceHaversine(b) + b.DistanceHaversine(c),
			delta:    1e-9,
		},
		{
			name: "depth ignored",
			ls: LineString{
				{Lon: a.Lon, Lat: a.Lat, Depth: 100},
				{Lon: b.Lon, Lat: b.Lat, Depth: -100},
			},
			expected: a.DistanceHaversine(b),
			delta:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ls.LengthHaversine()
			assert.False(t, math.IsNaN(got))
			assert.InDelta(t, tt.expected, got, tt.delta)
		})
	}
}

func TestNullLineString_LengthHaversine(t *testing.T) {
	ls := LineString{{Lon: 10, Lat: 59}, {Lon: 10, Lat: 59.001}}
	assert.InDelta(t, 111.194927, ls.LengthHaversine(), 1e-4)
}
