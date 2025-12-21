# geojson

A GeoJSON geometry package for Go with 3D coordinate support.

## Attribution

This package is derived from [github.com/paulmach/orb](https://github.com/paulmach/orb) by Paul Mach, licensed under the MIT License.

Key differences from the original:

- **Ergonomic Point type**: Named fields (`Lon`, `Lat`, `Depth`) instead of array indices
- **3D coordinates**: All geometry types support depth/altitude
- **Database integration**: `Null*` types with `sql.Scanner`/`driver.Valuer`
- **JSON-only**: Uses json/v2, no BSON/MongoDB dependencies

## License

MIT License - see the original [orb LICENSE](https://github.com/paulmach/orb/blob/master/LICENSE.md).
