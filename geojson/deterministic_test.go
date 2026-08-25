package geojson

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Feature.Properties is a map, and encoding/json/v2 emits map entries in Go
// map iteration order rather than sorted. Without json.Deterministic(true) in
// Feature.MarshalJSON the same Feature serialises to different bytes on every
// run, which breaks stable diffs, caching and golden tests for a format
// converter. Enough keys here that random ordering is essentially certain to
// show up if the option is ever dropped.
func TestFeature_MarshalJSON_Deterministic(t *testing.T) {
	f := NewFeature(Point{Lon: 10.5, Lat: 59.9})
	for _, k := range []string{
		"OBJTYPE", "KVALITET", "OMRÅDEID", "DATAFANGSTDATO", "OPPDATERINGSDATO",
		"HØYDE", "navn", "srid", "sosi_id", "active",
	} {
		f.Properties[k] = k + "-value"
	}

	want, err := f.MarshalJSON()
	require.NoError(t, err)

	for range 50 {
		got, err := f.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, string(want), string(got),
			"Feature.MarshalJSON must be byte-stable across calls")
	}
}

// The same guarantee has to survive through a FeatureCollection, since that is
// what the SOSI conversion entry points actually emit.
func TestFeatureCollection_MarshalJSON_Deterministic(t *testing.T) {
	fc := NewFeatureCollection()
	for i := range 5 {
		f := NewFeature(Point{Lon: float64(i), Lat: float64(i)})
		for _, k := range []string{"OBJTYPE", "KVALITET", "navn", "srid", "sosi_id"} {
			f.Properties[k] = k
		}
		fc.Append(f)
	}

	want, err := fc.MarshalJSON()
	require.NoError(t, err)

	for range 50 {
		got, err := fc.MarshalJSON()
		require.NoError(t, err)
		require.Equal(t, string(want), string(got),
			"FeatureCollection.MarshalJSON must be byte-stable across calls")
	}
}

// Deterministic means sorted by key, so the property order is predictable and
// not merely repeatable.
func TestFeature_MarshalJSON_SortsPropertyKeys(t *testing.T) {
	f := NewFeature(Point{Lon: 1, Lat: 2})
	f.Properties["zebra"] = 1
	f.Properties["alpha"] = 2
	f.Properties["mike"] = 3

	data, err := f.MarshalJSON()
	require.NoError(t, err)

	assert.Contains(t, string(data), `"properties":{"alpha":2,"mike":3,"zebra":1}`)
}
