package geojson

import (
	"testing"

	"github.com/go-json-experiment/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPoint_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		point    Point
		expected string
	}{
		{
			name:     "2D point",
			point:    Point{Lon: 10.5, Lat: 59.9},
			expected: `{"type":"Point","coordinates":[10.5,59.9]}`,
		},
		{
			name:     "3D point with depth",
			point:    Point{Lon: 10.5, Lat: 59.9, Depth: 2.5},
			expected: `{"type":"Point","coordinates":[10.5,59.9,2.5]}`,
		},
		{
			name:     "zero depth omitted",
			point:    Point{Lon: 10.5, Lat: 59.9, Depth: 0},
			expected: `{"type":"Point","coordinates":[10.5,59.9]}`,
		},
		{
			name:     "negative depth",
			point:    Point{Lon: 10.5, Lat: 59.9, Depth: -5.0},
			expected: `{"type":"Point","coordinates":[10.5,59.9,-5]}`,
		},
		{
			name:     "high precision coordinates",
			point:    Point{Lon: 10.654321987654, Lat: 59.123456789012, Depth: 1.234567},
			expected: `{"type":"Point","coordinates":[10.654321987654,59.123456789012,1.234567]}`,
		},
		{
			name:     "zero coordinates",
			point:    Point{Lon: 0, Lat: 0},
			expected: `{"type":"Point","coordinates":[0,0]}`,
		},
		{
			name:     "very small depth included",
			point:    Point{Lon: 10.5, Lat: 59.9, Depth: 0.001},
			expected: `{"type":"Point","coordinates":[10.5,59.9,0.001]}`,
		},
		{
			name:     "large depth value",
			point:    Point{Lon: 10.5, Lat: 59.9, Depth: 999.99},
			expected: `{"type":"Point","coordinates":[10.5,59.9,999.99]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.point)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestPoint_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Point
	}{
		{
			name:     "2D point",
			json:     `{"type":"Point","coordinates":[10.5,59.9]}`,
			expected: Point{Lon: 10.5, Lat: 59.9, Depth: 0},
		},
		{
			name:     "3D point with depth",
			json:     `{"type":"Point","coordinates":[10.5,59.9,2.5]}`,
			expected: Point{Lon: 10.5, Lat: 59.9, Depth: 2.5},
		},
		{
			name:     "negative depth",
			json:     `{"type":"Point","coordinates":[10.5,59.9,-5.0]}`,
			expected: Point{Lon: 10.5, Lat: 59.9, Depth: -5.0},
		},
		{
			name:     "high precision coordinates",
			json:     `{"type":"Point","coordinates":[10.654321987654,59.123456789012,1.234567]}`,
			expected: Point{Lon: 10.654321987654, Lat: 59.123456789012, Depth: 1.234567},
		},
		{
			name:     "zero coordinates",
			json:     `{"type":"Point","coordinates":[0,0]}`,
			expected: Point{Lon: 0, Lat: 0, Depth: 0},
		},
		{
			name:     "4D point ignores extra coordinates",
			json:     `{"type":"Point","coordinates":[10.5,59.9,2.5,99.9]}`,
			expected: Point{Lon: 10.5, Lat: 59.9, Depth: 2.5},
		},
		{
			name:     "very small depth",
			json:     `{"type":"Point","coordinates":[10.5,59.9,0.001]}`,
			expected: Point{Lon: 10.5, Lat: 59.9, Depth: 0.001},
		},
		{
			name:     "large depth value",
			json:     `{"type":"Point","coordinates":[10.5,59.9,999.99]}`,
			expected: Point{Lon: 10.5, Lat: 59.9, Depth: 999.99},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Point

			err := json.Unmarshal([]byte(tt.json), &p)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, p)
		})
	}
}

func TestPoint_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		errContains string
	}{
		{
			name:        "wrong type",
			json:        `{"type":"Polygon","coordinates":[10.5,59.9]}`,
			errContains: "expected Point",
		},
		{
			name:        "too few coordinates",
			json:        `{"type":"Point","coordinates":[10.5]}`,
			errContains: "at least 2 coordinates",
		},
		{
			name:        "empty coordinates",
			json:        `{"type":"Point","coordinates":[]}`,
			errContains: "at least 2 coordinates",
		},
		{
			name:        "invalid json",
			json:        `{"type":"Point","coordinates":invalid}`,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Point

			err := json.Unmarshal([]byte(tt.json), &p)
			require.Error(t, err)

			if tt.errContains != "" {
				assert.Contains(t, err.Error(), tt.errContains)
			}
		})
	}
}

func TestPoint_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		point Point
	}{
		{
			name:  "2D point",
			point: Point{Lon: 10.5, Lat: 59.9},
		},
		{
			name:  "3D point with depth",
			point: Point{Lon: 10.654321, Lat: 59.123456, Depth: 3.75},
		},
		{
			name:  "negative depth",
			point: Point{Lon: 10.5, Lat: 59.9, Depth: -2.5},
		},
		{
			name:  "high precision",
			point: Point{Lon: 10.654321987654, Lat: 59.123456789012, Depth: 1.234567890123},
		},
		{
			name:  "zero coordinates",
			point: Point{Lon: 0, Lat: 0, Depth: 0},
		},
		{
			name:  "very small depth",
			point: Point{Lon: 10.5, Lat: 59.9, Depth: 0.0001},
		},
		{
			name:  "large values",
			point: Point{Lon: 180.0, Lat: 90.0, Depth: 10000.0},
		},
		{
			name:  "negative lon lat",
			point: Point{Lon: -122.4194, Lat: -37.7749, Depth: 5.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.point)
			require.NoError(t, err)

			var parsed Point

			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)

			assert.Equal(t, tt.point, parsed)
		})
	}
}

func TestNullPoint_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected NullPoint
		wantErr  bool
	}{
		{
			name:  "scan 3D JSON string",
			input: `{"type":"Point","coordinates":[10.5,59.9,2.5]}`,
			expected: NullPoint{
				Point: Point{Lon: 10.5, Lat: 59.9, Depth: 2.5},
				Valid: true,
			},
		},
		{
			name:  "scan 2D JSON string",
			input: `{"type":"Point","coordinates":[10.5,59.9]}`,
			expected: NullPoint{
				Point: Point{Lon: 10.5, Lat: 59.9, Depth: 0},
				Valid: true,
			},
		},
		{
			name:  "scan 3D JSON bytes",
			input: []byte(`{"type":"Point","coordinates":[10.5,59.9,2.5]}`),
			expected: NullPoint{
				Point: Point{Lon: 10.5, Lat: 59.9, Depth: 2.5},
				Valid: true,
			},
		},
		{
			name:     "scan nil",
			input:    nil,
			expected: NullPoint{Valid: false},
		},
		{
			name:     "scan empty string",
			input:    "",
			expected: NullPoint{Valid: false},
		},
		{
			name:     "scan empty bytes",
			input:    []byte{},
			expected: NullPoint{Valid: false},
		},
		{
			name:  "high precision depth",
			input: `{"type":"Point","coordinates":[10.654321,59.123456,3.141592653589793]}`,
			expected: NullPoint{
				Point: Point{Lon: 10.654321, Lat: 59.123456, Depth: 3.141592653589793},
				Valid: true,
			},
		},
		{
			name:    "invalid JSON",
			input:   `{"type":"Point","coordinates":invalid}`,
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   12345,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var np NullPoint

			err := np.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Valid, np.Valid)

			if tt.expected.Valid {
				assert.Equal(t, tt.expected.Point, np.Point)
			}
		})
	}
}

func TestNullPoint_Value(t *testing.T) {
	tests := []struct {
		name     string
		np       NullPoint
		expected any
	}{
		{
			name: "valid 3D point",
			np: NullPoint{
				Point: Point{Lon: 10.5, Lat: 59.9, Depth: 2.5},
				Valid: true,
			},
			expected: `{"type":"Point","coordinates":[10.5,59.9,2.5]}`,
		},
		{
			name: "valid 2D point",
			np: NullPoint{
				Point: Point{Lon: 10.5, Lat: 59.9},
				Valid: true,
			},
			expected: `{"type":"Point","coordinates":[10.5,59.9]}`,
		},
		{
			name:     "null value",
			np:       NullPoint{Valid: false},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.np.Value()
			require.NoError(t, err)

			if tt.expected == nil {
				assert.Nil(t, val)
			} else {
				assert.JSONEq(t, tt.expected.(string), val.(string))
			}
		})
	}
}

func TestNullPoint_DatabaseRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		np   NullPoint
	}{
		{
			name: "3D point with depth",
			np: NullPoint{
				Point: Point{Lon: 10.5, Lat: 59.9, Depth: 2.5},
				Valid: true,
			},
		},
		{
			name: "2D point without depth",
			np: NullPoint{
				Point: Point{Lon: 10.5, Lat: 59.9},
				Valid: true,
			},
		},
		{
			name: "high precision coordinates",
			np: NullPoint{
				Point: Point{Lon: 10.654321987654, Lat: 59.123456789012, Depth: 1.234567890123},
				Valid: true,
			},
		},
		{
			name: "negative depth",
			np: NullPoint{
				Point: Point{Lon: 10.5, Lat: 59.9, Depth: -5.0},
				Valid: true,
			},
		},
		{
			name: "null point",
			np:   NullPoint{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.np.Value()
			require.NoError(t, err)

			var np2 NullPoint

			err = np2.Scan(val)
			require.NoError(t, err)

			assert.True(t, tt.np.Equal(np2), "round trip should preserve value")
		})
	}
}

func TestNullPoint_Equal(t *testing.T) {
	tests := []struct {
		name     string
		a        NullPoint
		b        NullPoint
		expected bool
	}{
		{
			name:     "both invalid",
			a:        NullPoint{Valid: false},
			b:        NullPoint{Valid: false},
			expected: true,
		},
		{
			name:     "first valid second invalid",
			a:        NullPoint{Point: Point{Lon: 10, Lat: 59}, Valid: true},
			b:        NullPoint{Valid: false},
			expected: false,
		},
		{
			name:     "first invalid second valid",
			a:        NullPoint{Valid: false},
			b:        NullPoint{Point: Point{Lon: 10, Lat: 59}, Valid: true},
			expected: false,
		},
		{
			name:     "equal 3D points",
			a:        NullPoint{Point: Point{Lon: 10, Lat: 59, Depth: 2.5}, Valid: true},
			b:        NullPoint{Point: Point{Lon: 10, Lat: 59, Depth: 2.5}, Valid: true},
			expected: true,
		},
		{
			name:     "equal 2D points",
			a:        NullPoint{Point: Point{Lon: 10, Lat: 59}, Valid: true},
			b:        NullPoint{Point: Point{Lon: 10, Lat: 59}, Valid: true},
			expected: true,
		},
		{
			name:     "different depth",
			a:        NullPoint{Point: Point{Lon: 10, Lat: 59, Depth: 2.5}, Valid: true},
			b:        NullPoint{Point: Point{Lon: 10, Lat: 59, Depth: 3.0}, Valid: true},
			expected: false,
		},
		{
			name:     "different lon",
			a:        NullPoint{Point: Point{Lon: 10, Lat: 59, Depth: 2.5}, Valid: true},
			b:        NullPoint{Point: Point{Lon: 11, Lat: 59, Depth: 2.5}, Valid: true},
			expected: false,
		},
		{
			name:     "different lat",
			a:        NullPoint{Point: Point{Lon: 10, Lat: 59, Depth: 2.5}, Valid: true},
			b:        NullPoint{Point: Point{Lon: 10, Lat: 60, Depth: 2.5}, Valid: true},
			expected: false,
		},
		{
			name:     "zero depth equals zero depth",
			a:        NullPoint{Point: Point{Lon: 10, Lat: 59, Depth: 0}, Valid: true},
			b:        NullPoint{Point: Point{Lon: 10, Lat: 59, Depth: 0}, Valid: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Equal(tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLineString_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		ls       LineString
		expected string
	}{
		{
			name:     "2D linestring",
			ls:       LineString{{Lon: 10.1, Lat: 59.1}, {Lon: 10.2, Lat: 59.2}},
			expected: `{"type":"LineString","coordinates":[[10.1,59.1],[10.2,59.2]]}`,
		},
		{
			name: "3D linestring",
			ls: LineString{
				{Lon: 10.1, Lat: 59.1, Depth: 1.5},
				{Lon: 10.2, Lat: 59.2, Depth: 2.5},
			},
			expected: `{"type":"LineString","coordinates":[[10.1,59.1,1.5],[10.2,59.2,2.5]]}`,
		},
		{
			name:     "mixed 2D and 3D",
			ls:       LineString{{Lon: 10.1, Lat: 59.1}, {Lon: 10.2, Lat: 59.2, Depth: 2.5}},
			expected: `{"type":"LineString","coordinates":[[10.1,59.1],[10.2,59.2,2.5]]}`,
		},
		{
			name:     "empty linestring",
			ls:       LineString{},
			expected: `{"type":"LineString","coordinates":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.ls)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestLineString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected LineString
	}{
		{
			name:     "2D linestring",
			json:     `{"type":"LineString","coordinates":[[10.1,59.1],[10.2,59.2]]}`,
			expected: LineString{{Lon: 10.1, Lat: 59.1}, {Lon: 10.2, Lat: 59.2}},
		},
		{
			name: "3D linestring",
			json: `{"type":"LineString","coordinates":[[10.1,59.1,1.5],[10.2,59.2,2.5]]}`,
			expected: LineString{
				{Lon: 10.1, Lat: 59.1, Depth: 1.5},
				{Lon: 10.2, Lat: 59.2, Depth: 2.5},
			},
		},
		{
			name:     "empty linestring",
			json:     `{"type":"LineString","coordinates":[]}`,
			expected: LineString{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ls LineString

			err := json.Unmarshal([]byte(tt.json), &ls)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, ls)
		})
	}
}

func TestLineString_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		errContains string
	}{
		{
			name:        "wrong type",
			json:        `{"type":"Point","coordinates":[[10.1,59.1]]}`,
			errContains: "expected LineString",
		},
		{
			name:        "insufficient coordinates in point",
			json:        `{"type":"LineString","coordinates":[[10.1]]}`,
			errContains: "at least 2 coordinates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ls LineString

			err := json.Unmarshal([]byte(tt.json), &ls)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestLineString_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		ls   LineString
	}{
		{
			name: "2D linestring",
			ls:   LineString{{Lon: 10.1, Lat: 59.1}, {Lon: 10.2, Lat: 59.2}},
		},
		{
			name: "3D linestring",
			ls: LineString{
				{Lon: 10.1, Lat: 59.1, Depth: 1.5},
				{Lon: 10.2, Lat: 59.2, Depth: 2.5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.ls)
			require.NoError(t, err)

			var parsed LineString

			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)
			assert.Equal(t, tt.ls, parsed)
		})
	}
}

func TestNullLineString_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected NullLineString
		wantErr  bool
	}{
		{
			name:  "scan JSON string",
			input: `{"type":"LineString","coordinates":[[10.1,59.1],[10.2,59.2]]}`,
			expected: NullLineString{
				LineString: LineString{{Lon: 10.1, Lat: 59.1}, {Lon: 10.2, Lat: 59.2}},
				Valid:      true,
			},
		},
		{
			name:  "scan JSON bytes",
			input: []byte(`{"type":"LineString","coordinates":[[10.1,59.1,1.5],[10.2,59.2,2.5]]}`),
			expected: NullLineString{
				LineString: LineString{
					{Lon: 10.1, Lat: 59.1, Depth: 1.5},
					{Lon: 10.2, Lat: 59.2, Depth: 2.5},
				},
				Valid: true,
			},
		},
		{
			name:     "scan nil",
			input:    nil,
			expected: NullLineString{Valid: false},
		},
		{
			name:     "scan empty string",
			input:    "",
			expected: NullLineString{Valid: false},
		},
		{
			name:    "unsupported type",
			input:   12345,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nls NullLineString

			err := nls.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Valid, nls.Valid)

			if tt.expected.Valid {
				assert.Equal(t, tt.expected.LineString, nls.LineString)
			}
		})
	}
}

func TestNullLineString_Value(t *testing.T) {
	tests := []struct {
		name     string
		nls      NullLineString
		expected any
	}{
		{
			name: "valid linestring",
			nls: NullLineString{
				LineString: LineString{{Lon: 10.1, Lat: 59.1}, {Lon: 10.2, Lat: 59.2}},
				Valid:      true,
			},
			expected: `{"type":"LineString","coordinates":[[10.1,59.1],[10.2,59.2]]}`,
		},
		{
			name:     "null value",
			nls:      NullLineString{Valid: false},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.nls.Value()
			require.NoError(t, err)

			if tt.expected == nil {
				assert.Nil(t, val)
			} else {
				assert.JSONEq(t, tt.expected.(string), val.(string))
			}
		})
	}
}

func TestNullLineString_DatabaseRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		nls  NullLineString
	}{
		{
			name: "2D linestring",
			nls: NullLineString{
				LineString: LineString{{Lon: 10.1, Lat: 59.1}, {Lon: 10.2, Lat: 59.2}},
				Valid:      true,
			},
		},
		{
			name: "3D linestring",
			nls: NullLineString{
				LineString: LineString{
					{Lon: 10.1, Lat: 59.1, Depth: 1.5},
					{Lon: 10.2, Lat: 59.2, Depth: 2.5},
				},
				Valid: true,
			},
		},
		{
			name: "null linestring",
			nls:  NullLineString{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.nls.Value()
			require.NoError(t, err)

			var nls2 NullLineString

			err = nls2.Scan(val)
			require.NoError(t, err)

			assert.True(t, tt.nls.Equal(nls2), "round trip should preserve value")
		})
	}
}

func TestNullLineString_Equal(t *testing.T) {
	tests := []struct {
		name     string
		a        NullLineString
		b        NullLineString
		expected bool
	}{
		{
			name:     "both invalid",
			a:        NullLineString{Valid: false},
			b:        NullLineString{Valid: false},
			expected: true,
		},
		{
			name: "first valid second invalid",
			a: NullLineString{
				LineString: LineString{{Lon: 10, Lat: 59}},
				Valid:      true,
			},
			b:        NullLineString{Valid: false},
			expected: false,
		},
		{
			name: "equal linestrings",
			a: NullLineString{
				LineString: LineString{{Lon: 10, Lat: 59}, {Lon: 11, Lat: 60}},
				Valid:      true,
			},
			b: NullLineString{
				LineString: LineString{{Lon: 10, Lat: 59}, {Lon: 11, Lat: 60}},
				Valid:      true,
			},
			expected: true,
		},
		{
			name: "different lengths",
			a: NullLineString{
				LineString: LineString{{Lon: 10, Lat: 59}},
				Valid:      true,
			},
			b: NullLineString{
				LineString: LineString{{Lon: 10, Lat: 59}, {Lon: 11, Lat: 60}},
				Valid:      true,
			},
			expected: false,
		},
		{
			name: "different coordinates",
			a: NullLineString{
				LineString: LineString{{Lon: 10, Lat: 59}, {Lon: 11, Lat: 60}},
				Valid:      true,
			},
			b: NullLineString{
				LineString: LineString{{Lon: 10, Lat: 59}, {Lon: 12, Lat: 60}},
				Valid:      true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Equal(tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Ring tests

func TestRing_IsClosed(t *testing.T) {
	tests := []struct {
		name     string
		ring     Ring
		expected bool
	}{
		{
			name:     "closed ring",
			ring:     Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}},
			expected: true,
		},
		{
			name:     "open ring",
			ring:     Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}},
			expected: false,
		},
		{
			name:     "closed ring with depth",
			ring:     Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 5}, {Lon: 0, Lat: 0, Depth: 5}},
			expected: true,
		},
		{
			name:     "open ring different depth",
			ring:     Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 5}, {Lon: 0, Lat: 0, Depth: 10}},
			expected: false,
		},
		{
			name:     "empty ring",
			ring:     Ring{},
			expected: false,
		},
		{
			name:     "single point",
			ring:     Ring{{Lon: 0, Lat: 0}},
			expected: false,
		},
		{
			name:     "two same points",
			ring:     Ring{{Lon: 0, Lat: 0}, {Lon: 0, Lat: 0}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ring.IsClosed()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRing_Close(t *testing.T) {
	tests := []struct {
		name     string
		ring     Ring
		expected Ring
	}{
		{
			name:     "already closed",
			ring:     Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 0, Lat: 0}},
			expected: Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 0, Lat: 0}},
		},
		{
			name:     "needs closing",
			ring:     Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}},
			expected: Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}},
		},
		{
			name:     "empty ring",
			ring:     Ring{},
			expected: Ring{},
		},
		{
			name:     "with depth",
			ring:     Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 10}},
			expected: Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 10}, {Lon: 0, Lat: 0, Depth: 5}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ring.Close()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRing_Equal(t *testing.T) {
	tests := []struct {
		name     string
		a        Ring
		b        Ring
		expected bool
	}{
		{
			name:     "equal rings",
			a:        Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
			b:        Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
			expected: true,
		},
		{
			name:     "different lengths",
			a:        Ring{{Lon: 0, Lat: 0}},
			b:        Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
			expected: false,
		},
		{
			name:     "different coordinates",
			a:        Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
			b:        Ring{{Lon: 0, Lat: 0}, {Lon: 2, Lat: 2}},
			expected: false,
		},
		{
			name:     "different depth",
			a:        Ring{{Lon: 0, Lat: 0, Depth: 5}},
			b:        Ring{{Lon: 0, Lat: 0, Depth: 10}},
			expected: false,
		},
		{
			name:     "both empty",
			a:        Ring{},
			b:        Ring{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Equal(tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Polygon tests

func TestPolygon_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		polygon  Polygon
		expected string
	}{
		{
			name: "simple polygon",
			polygon: Polygon{
				Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}},
			},
			expected: `{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]]]}`,
		},
		{
			name: "polygon with hole",
			polygon: Polygon{
				Ring{{Lon: 0, Lat: 0}, {Lon: 10, Lat: 0}, {Lon: 10, Lat: 10}, {Lon: 0, Lat: 0}},
				Ring{{Lon: 2, Lat: 2}, {Lon: 8, Lat: 2}, {Lon: 8, Lat: 8}, {Lon: 2, Lat: 2}},
			},
			expected: `{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,0]],[[2,2],[8,2],[8,8],[2,2]]]}`,
		},
		{
			name: "polygon with depth",
			polygon: Polygon{
				Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 5}, {Lon: 0, Lat: 0, Depth: 5}},
			},
			expected: `{"type":"Polygon","coordinates":[[[0,0,5],[1,0,5],[0,0,5]]]}`,
		},
		{
			name: "auto-closes open ring",
			polygon: Polygon{
				Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}},
			},
			expected: `{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]]]}`,
		},
		{
			name:     "empty polygon",
			polygon:  Polygon{},
			expected: `{"type":"Polygon","coordinates":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.polygon)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestPolygon_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected Polygon
	}{
		{
			name: "simple polygon",
			json: `{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]]]}`,
			expected: Polygon{
				Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}},
			},
		},
		{
			name: "polygon with hole",
			json: `{"type":"Polygon","coordinates":[[[0,0],[10,0],[10,10],[0,0]],[[2,2],[8,2],[8,8],[2,2]]]}`,
			expected: Polygon{
				Ring{{Lon: 0, Lat: 0}, {Lon: 10, Lat: 0}, {Lon: 10, Lat: 10}, {Lon: 0, Lat: 0}},
				Ring{{Lon: 2, Lat: 2}, {Lon: 8, Lat: 2}, {Lon: 8, Lat: 8}, {Lon: 2, Lat: 2}},
			},
		},
		{
			name: "polygon with depth",
			json: `{"type":"Polygon","coordinates":[[[0,0,5],[1,0,5],[0,0,5]]]}`,
			expected: Polygon{
				Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 5}, {Lon: 0, Lat: 0, Depth: 5}},
			},
		},
		{
			name:     "empty polygon",
			json:     `{"type":"Polygon","coordinates":[]}`,
			expected: Polygon{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Polygon

			err := json.Unmarshal([]byte(tt.json), &p)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, p)
		})
	}
}

func TestPolygon_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		errContains string
	}{
		{
			name:        "wrong type",
			json:        `{"type":"Point","coordinates":[[[0,0]]]}`,
			errContains: "expected Polygon",
		},
		{
			name:        "insufficient coordinates in point",
			json:        `{"type":"Polygon","coordinates":[[[0]]]}`,
			errContains: "at least 2 coordinates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Polygon

			err := json.Unmarshal([]byte(tt.json), &p)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestPolygon_RoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		polygon Polygon
	}{
		{
			name: "simple polygon",
			polygon: Polygon{
				Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}},
			},
		},
		{
			name: "polygon with hole",
			polygon: Polygon{
				Ring{{Lon: 0, Lat: 0}, {Lon: 10, Lat: 0}, {Lon: 10, Lat: 10}, {Lon: 0, Lat: 0}},
				Ring{{Lon: 2, Lat: 2}, {Lon: 8, Lat: 2}, {Lon: 8, Lat: 8}, {Lon: 2, Lat: 2}},
			},
		},
		{
			name: "polygon with depth",
			polygon: Polygon{
				Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 10}, {Lon: 0, Lat: 0, Depth: 5}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.polygon)
			require.NoError(t, err)

			var parsed Polygon

			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)
			assert.Equal(t, tt.polygon, parsed)
		})
	}
}

// NullPolygon tests

func TestNullPolygon_Scan(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected NullPolygon
		wantErr  bool
	}{
		{
			name:  "scan JSON string",
			input: `{"type":"Polygon","coordinates":[[[0,0],[1,0],[1,1],[0,0]]]}`,
			expected: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}}},
				Valid:   true,
			},
		},
		{
			name:  "scan JSON bytes",
			input: []byte(`{"type":"Polygon","coordinates":[[[0,0,5],[1,0,5],[0,0,5]]]}`),
			expected: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 5}, {Lon: 0, Lat: 0, Depth: 5}}},
				Valid:   true,
			},
		},
		{
			name:     "scan nil",
			input:    nil,
			expected: NullPolygon{Valid: false},
		},
		{
			name:     "scan empty string",
			input:    "",
			expected: NullPolygon{Valid: false},
		},
		{
			name:    "unsupported type",
			input:   12345,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var np NullPolygon

			err := np.Scan(tt.input)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected.Valid, np.Valid)

			if tt.expected.Valid {
				assert.Equal(t, tt.expected.Polygon, np.Polygon)
			}
		})
	}
}

func TestNullPolygon_Value(t *testing.T) {
	tests := []struct {
		name     string
		np       NullPolygon
		expected any
	}{
		{
			name: "valid polygon",
			np: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 0, Lat: 0}}},
				Valid:   true,
			},
			expected: `{"type":"Polygon","coordinates":[[[0,0],[1,0],[0,0]]]}`,
		},
		{
			name:     "null value",
			np:       NullPolygon{Valid: false},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.np.Value()
			require.NoError(t, err)

			if tt.expected == nil {
				assert.Nil(t, val)
			} else {
				assert.JSONEq(t, tt.expected.(string), val.(string))
			}
		})
	}
}

func TestNullPolygon_DatabaseRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		np   NullPolygon
	}{
		{
			name: "simple polygon",
			np: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}}},
				Valid:   true,
			},
		},
		{
			name: "polygon with depth",
			np: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 10}, {Lon: 0, Lat: 0, Depth: 5}}},
				Valid:   true,
			},
		},
		{
			name: "null polygon",
			np:   NullPolygon{Valid: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.np.Value()
			require.NoError(t, err)

			var np2 NullPolygon

			err = np2.Scan(val)
			require.NoError(t, err)

			assert.True(t, tt.np.Equal(np2), "round trip should preserve value")
		})
	}
}

func TestNullPolygon_Equal(t *testing.T) {
	tests := []struct {
		name     string
		a        NullPolygon
		b        NullPolygon
		expected bool
	}{
		{
			name:     "both invalid",
			a:        NullPolygon{Valid: false},
			b:        NullPolygon{Valid: false},
			expected: true,
		},
		{
			name: "first valid second invalid",
			a: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}}},
				Valid:   true,
			},
			b:        NullPolygon{Valid: false},
			expected: false,
		},
		{
			name: "equal polygons",
			a: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}}},
				Valid:   true,
			},
			b: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}}},
				Valid:   true,
			},
			expected: true,
		},
		{
			name: "different ring count",
			a: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}}},
				Valid:   true,
			},
			b: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}}, Ring{{Lon: 1, Lat: 1}}},
				Valid:   true,
			},
			expected: false,
		},
		{
			name: "different coordinates",
			a: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 0, Lat: 0}}},
				Valid:   true,
			},
			b: NullPolygon{
				Polygon: Polygon{Ring{{Lon: 1, Lat: 1}}},
				Valid:   true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.a.Equal(tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// MultiPoint tests

func TestMultiPoint_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		mp       MultiPoint
		expected string
	}{
		{
			name:     "2D multipoint",
			mp:       MultiPoint{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
			expected: `{"type":"MultiPoint","coordinates":[[0,0],[1,1]]}`,
		},
		{
			name:     "3D multipoint",
			mp:       MultiPoint{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 1, Depth: 10}},
			expected: `{"type":"MultiPoint","coordinates":[[0,0,5],[1,1,10]]}`,
		},
		{
			name:     "mixed 2D and 3D",
			mp:       MultiPoint{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1, Depth: 5}},
			expected: `{"type":"MultiPoint","coordinates":[[0,0],[1,1,5]]}`,
		},
		{
			name:     "empty multipoint",
			mp:       MultiPoint{},
			expected: `{"type":"MultiPoint","coordinates":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.mp)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestMultiPoint_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected MultiPoint
	}{
		{
			name:     "2D multipoint",
			json:     `{"type":"MultiPoint","coordinates":[[0,0],[1,1]]}`,
			expected: MultiPoint{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
		},
		{
			name:     "3D multipoint",
			json:     `{"type":"MultiPoint","coordinates":[[0,0,5],[1,1,10]]}`,
			expected: MultiPoint{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 1, Depth: 10}},
		},
		{
			name:     "empty multipoint",
			json:     `{"type":"MultiPoint","coordinates":[]}`,
			expected: MultiPoint{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mp MultiPoint

			err := json.Unmarshal([]byte(tt.json), &mp)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, mp)
		})
	}
}

func TestMultiPoint_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		errContains string
	}{
		{
			name:        "wrong type",
			json:        `{"type":"Point","coordinates":[[0,0]]}`,
			errContains: "expected MultiPoint",
		},
		{
			name:        "insufficient coordinates",
			json:        `{"type":"MultiPoint","coordinates":[[0]]}`,
			errContains: "at least 2 coordinates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mp MultiPoint

			err := json.Unmarshal([]byte(tt.json), &mp)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestMultiPoint_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		mp   MultiPoint
	}{
		{
			name: "2D multipoint",
			mp:   MultiPoint{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
		},
		{
			name: "3D multipoint",
			mp:   MultiPoint{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 1, Depth: 10}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.mp)
			require.NoError(t, err)

			var parsed MultiPoint

			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)
			assert.Equal(t, tt.mp, parsed)
		})
	}
}

// MultiLineString tests

func TestMultiLineString_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		mls      MultiLineString
		expected string
	}{
		{
			name: "2D multilinestring",
			mls: MultiLineString{
				LineString{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
				LineString{{Lon: 2, Lat: 2}, {Lon: 3, Lat: 3}},
			},
			expected: `{"type":"MultiLineString","coordinates":[[[0,0],[1,1]],[[2,2],[3,3]]]}`,
		},
		{
			name: "3D multilinestring",
			mls: MultiLineString{
				LineString{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 1, Depth: 10}},
			},
			expected: `{"type":"MultiLineString","coordinates":[[[0,0,5],[1,1,10]]]}`,
		},
		{
			name:     "empty multilinestring",
			mls:      MultiLineString{},
			expected: `{"type":"MultiLineString","coordinates":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.mls)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestMultiLineString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected MultiLineString
	}{
		{
			name: "2D multilinestring",
			json: `{"type":"MultiLineString","coordinates":[[[0,0],[1,1]],[[2,2],[3,3]]]}`,
			expected: MultiLineString{
				LineString{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
				LineString{{Lon: 2, Lat: 2}, {Lon: 3, Lat: 3}},
			},
		},
		{
			name: "3D multilinestring",
			json: `{"type":"MultiLineString","coordinates":[[[0,0,5],[1,1,10]]]}`,
			expected: MultiLineString{
				LineString{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 1, Depth: 10}},
			},
		},
		{
			name:     "empty multilinestring",
			json:     `{"type":"MultiLineString","coordinates":[]}`,
			expected: MultiLineString{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mls MultiLineString

			err := json.Unmarshal([]byte(tt.json), &mls)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, mls)
		})
	}
}

func TestMultiLineString_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		errContains string
	}{
		{
			name:        "wrong type",
			json:        `{"type":"LineString","coordinates":[[[0,0]]]}`,
			errContains: "expected MultiLineString",
		},
		{
			name:        "insufficient coordinates",
			json:        `{"type":"MultiLineString","coordinates":[[[0]]]}`,
			errContains: "at least 2 coordinates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mls MultiLineString

			err := json.Unmarshal([]byte(tt.json), &mls)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestMultiLineString_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		mls  MultiLineString
	}{
		{
			name: "2D multilinestring",
			mls: MultiLineString{
				LineString{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 1}},
				LineString{{Lon: 2, Lat: 2}, {Lon: 3, Lat: 3}},
			},
		},
		{
			name: "3D multilinestring",
			mls: MultiLineString{
				LineString{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 1, Depth: 10}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.mls)
			require.NoError(t, err)

			var parsed MultiLineString

			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)
			assert.Equal(t, tt.mls, parsed)
		})
	}
}

// MultiPolygon tests

func TestMultiPolygon_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		mpg      MultiPolygon
		expected string
	}{
		{
			name: "simple multipolygon",
			mpg: MultiPolygon{
				Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}}},
			},
			expected: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[1,1],[0,0]]]]}`,
		},
		{
			name: "multiple polygons",
			mpg: MultiPolygon{
				Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 0, Lat: 0}}},
				Polygon{Ring{{Lon: 2, Lat: 2}, {Lon: 3, Lat: 2}, {Lon: 2, Lat: 2}}},
			},
			expected: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,0]]],[[[2,2],[3,2],[2,2]]]]}`,
		},
		{
			name: "with depth",
			mpg: MultiPolygon{
				Polygon{Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 5}, {Lon: 0, Lat: 0, Depth: 5}}},
			},
			expected: `{"type":"MultiPolygon","coordinates":[[[[0,0,5],[1,0,5],[0,0,5]]]]}`,
		},
		{
			name:     "empty multipolygon",
			mpg:      MultiPolygon{},
			expected: `{"type":"MultiPolygon","coordinates":[]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.mpg)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestMultiPolygon_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected MultiPolygon
	}{
		{
			name: "simple multipolygon",
			json: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[1,1],[0,0]]]]}`,
			expected: MultiPolygon{
				Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}}},
			},
		},
		{
			name: "multiple polygons",
			json: `{"type":"MultiPolygon","coordinates":[[[[0,0],[1,0],[0,0]]],[[[2,2],[3,2],[2,2]]]]}`,
			expected: MultiPolygon{
				Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 0, Lat: 0}}},
				Polygon{Ring{{Lon: 2, Lat: 2}, {Lon: 3, Lat: 2}, {Lon: 2, Lat: 2}}},
			},
		},
		{
			name: "with depth",
			json: `{"type":"MultiPolygon","coordinates":[[[[0,0,5],[1,0,5],[0,0,5]]]]}`,
			expected: MultiPolygon{
				Polygon{Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 5}, {Lon: 0, Lat: 0, Depth: 5}}},
			},
		},
		{
			name:     "empty multipolygon",
			json:     `{"type":"MultiPolygon","coordinates":[]}`,
			expected: MultiPolygon{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mpg MultiPolygon

			err := json.Unmarshal([]byte(tt.json), &mpg)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, mpg)
		})
	}
}

func TestMultiPolygon_UnmarshalJSON_Errors(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		errContains string
	}{
		{
			name:        "wrong type",
			json:        `{"type":"Polygon","coordinates":[[[[0,0]]]]}`,
			errContains: "expected MultiPolygon",
		},
		{
			name:        "insufficient coordinates",
			json:        `{"type":"MultiPolygon","coordinates":[[[[0]]]]}`,
			errContains: "at least 2 coordinates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mpg MultiPolygon

			err := json.Unmarshal([]byte(tt.json), &mpg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errContains)
		})
	}
}

func TestMultiPolygon_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		mpg  MultiPolygon
	}{
		{
			name: "simple multipolygon",
			mpg: MultiPolygon{
				Polygon{Ring{{Lon: 0, Lat: 0}, {Lon: 1, Lat: 0}, {Lon: 1, Lat: 1}, {Lon: 0, Lat: 0}}},
			},
		},
		{
			name: "with depth",
			mpg: MultiPolygon{
				Polygon{Ring{{Lon: 0, Lat: 0, Depth: 5}, {Lon: 1, Lat: 0, Depth: 10}, {Lon: 0, Lat: 0, Depth: 5}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.mpg)
			require.NoError(t, err)

			var parsed MultiPolygon

			err = json.Unmarshal(data, &parsed)
			require.NoError(t, err)
			assert.Equal(t, tt.mpg, parsed)
		})
	}
}
