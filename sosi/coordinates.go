package sosi

import (
	"fmt"
	"strconv"
	"strings"
)

// FormatCoordinateToSOSI converts a coordinate value to SOSI format
// This removes the decimal point as in JavaScript's toSOSI function
func FormatCoordinateToSOSI(value float64, precision int) string {
	formatted := fmt.Sprintf("%."+strconv.Itoa(precision)+"f", value)
	return strings.Replace(formatted, ".", "", 1)
}

// TransformCoordinate converts SOSI coordinate to SOSI format "lat lon alt"
// Note: Coordinate.X=longitude, Coordinate.Y=latitude, but SOSI expects latitude longitude
func TransformCoordinate(coord Coordinate, config SOSIConfig) string {
	// Format coordinates with proper precision and remove decimal points
	lat := FormatCoordinateToSOSI(coord.Y, config.LatLongAccuracy) // Y = Latitude
	lon := FormatCoordinateToSOSI(coord.X, config.LatLongAccuracy) // X = Longitude

	// Default altitude to 0 if not provided, format with altitude accuracy
	alt := coord.Z
	if alt == 0 {
		// Ensure we have proper zero padding for altitude
		alt = 0.0
	}
	altFormatted := FormatCoordinateToSOSI(alt, config.AltitudeAccuracy)

	return fmt.Sprintf("%s %s %s", lat, lon, altFormatted)
}

// ConvertCoordinate converts from public API coordinate format to internal coordinate format
func ConvertCoordinate(lat, lon, alt float64) Coordinate {
	return Coordinate{
		X: lon, // X = Longitude
		Y: lat, // Y = Latitude
		Z: alt, // Z = Altitude
	}
}

// TransformCoordinates converts a slice of coordinates to SOSI format strings
func TransformCoordinates(coords []Coordinate, config SOSIConfig) []string {
	result := make([]string, len(coords))
	for i, coord := range coords {
		result[i] = TransformCoordinate(coord, config)
	}
	return result
}

// CoordinateSystem represents a coordinate system with SRID and proj4 definition
type CoordinateSystem struct {
	SRID string // EPSG code (e.g., "EPSG:4326")
	Def  string // proj4 definition string
}

// getNorwegianCoordinateSystems returns the complete mapping of Norwegian coordinate systems
// Equivalent to JavaScript koordsysMap and geosysMap from util.js (lines 203-243)
func getNorwegianCoordinateSystems() map[int]CoordinateSystem {
	return map[int]CoordinateSystem{
		// NGO coordinate systems (Norwegian Geodetic Datum 1948)
		1: {
			SRID: "EPSG:27391",
			Def:  "+proj=tmerc +lat_0=58 +lon_0=-4.666666666666667 +k=1 +x_0=0 +y_0=0 +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +pm=oslo +units=m +no_defs",
		},
		2: {
			SRID: "EPSG:27392",
			Def:  "+proj=tmerc +lat_0=58 +lon_0=-2.333333333333333 +k=1 +x_0=0 +y_0=0 +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +pm=oslo +units=m +no_defs",
		},
		3: {
			SRID: "EPSG:27393",
			Def:  "+proj=tmerc +lat_0=58 +lon_0=0 +k=1 +x_0=0 +y_0=0 +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +pm=oslo +units=m +no_defs",
		},
		4: {
			SRID: "EPSG:27394",
			Def:  "+proj=tmerc +lat_0=58 +lon_0=2.5 +k=1 +x_0=0 +y_0=0 +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +pm=oslo +units=m +no_defs",
		},
		5: {
			SRID: "EPSG:27395",
			Def:  "+proj=tmerc +lat_0=58 +lon_0=6.166666666666667 +k=1 +x_0=0 +y_0=0 +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +pm=oslo +units=m +no_defs",
		},
		6: {
			SRID: "EPSG:27396",
			Def:  "+proj=tmerc +lat_0=58 +lon_0=10.16666666666667 +k=1 +x_0=0 +y_0=0 +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +pm=oslo +units=m +no_defs",
		},
		7: {
			SRID: "EPSG:27397",
			Def:  "+proj=tmerc +lat_0=58 +lon_0=14.16666666666667 +k=1 +x_0=0 +y_0=0 +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +pm=oslo +units=m +no_defs",
		},
		8: {
			SRID: "EPSG:27398",
			Def:  "+proj=tmerc +lat_0=58 +lon_0=18.33333333333333 +k=1 +x_0=0 +y_0=0 +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +pm=oslo +units=m +no_defs",
		},
		9: {
			SRID: "EPSG:4273",
			Def:  "+proj=longlat +a=6377492.018 +b=6356173.508712696 +towgs84=278.3,93,474.5,7.889,0.05,-6.61,6.21 +no_defs",
		},

		// UTM WGS84 zones
		21: {
			SRID: "EPSG:32631",
			Def:  "+proj=utm +zone=31 +ellps=WGS84 +datum=WGS84 +units=m +no_defs",
		},
		22: {
			SRID: "EPSG:32632",
			Def:  "+proj=utm +zone=32 +ellps=WGS84 +datum=WGS84 +units=m +no_defs",
		},
		23: {
			SRID: "EPSG:32633",
			Def:  "+proj=utm +zone=33 +ellps=WGS84 +datum=WGS84 +units=m +no_defs",
		},
		24: {
			SRID: "EPSG:32634",
			Def:  "+proj=utm +zone=34 +ellps=WGS84 +datum=WGS84 +units=m +no_defs",
		},
		25: {
			SRID: "EPSG:32635",
			Def:  "+proj=utm +zone=35 +ellps=WGS84 +datum=WGS84 +units=m +no_defs",
		},
		26: {
			SRID: "EPSG:32636",
			Def:  "+proj=utm +zone=35 +ellps=WGS84 +datum=WGS84 +units=m +no_defs",
		},

		// UTM ED50 zones
		31: {
			SRID: "EPSG:23031",
			Def:  "+proj=utm +zone=31 +ellps=intl +units=m +no_defs",
		},
		32: {
			SRID: "EPSG:23032",
			Def:  "+proj=utm +zone=32 +ellps=intl +units=m +no_defs",
		},
		33: {
			SRID: "EPSG:23033",
			Def:  "+proj=utm +zone=33 +ellps=intl +units=m +no_defs",
		},
		34: {
			SRID: "EPSG:23034",
			Def:  "+proj=utm +zone=34 +ellps=intl +units=m +no_defs",
		},
		35: {
			SRID: "EPSG:23035",
			Def:  "+proj=utm +zone=35 +ellps=intl +units=m +no_defs",
		},
		36: {
			SRID: "EPSG:23036",
			Def:  "+proj=utm +zone=36 +ellps=intl +units=m +no_defs",
		},

		// Geographic coordinate systems
		50: {
			SRID: "EPSG:4230",
			Def:  "+proj=longlat +ellps=intl +no_defs",
		},
		72: {
			SRID: "EPSG:4322",
			Def:  "+proj=longlat +ellps=WGS72 +no_defs",
		},
		84: {
			SRID: "EPSG:4326",
			Def:  "+proj=longlat +ellps=WGS84 +datum=WGS84 +no_defs",
		},
		87: {
			SRID: "EPSG:4231",
			Def:  "+proj=longlat +ellps=intl +no_defs",
		},

		// Additional geosys system (2 was already defined above as EPSG:27392)
		// This represents the geosysMap entry from JavaScript
	}
}

// GetSRIDFromCoordSystem returns the EPSG SRID for a given Norwegian coordinate system
// Returns empty string if the coordinate system is not found
func GetSRIDFromCoordSystem(coordSystem int) string {
	systems := getNorwegianCoordinateSystems()
	if system, exists := systems[coordSystem]; exists {
		return system.SRID
	}
	return ""
}

// GetCoordinateSystemFromSRID returns the coordinate system ID for a given EPSG SRID
// Returns 0 if the SRID is not found
func GetCoordinateSystemFromSRID(srid string) int {
	systems := getNorwegianCoordinateSystems()
	for coordSystem, system := range systems {
		if system.SRID == srid {
			return coordSystem
		}
	}
	return 0
}
