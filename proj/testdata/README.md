# Projection reference vectors

`vectors.json` holds independent ground-truth samples used by
`../reference_test.go` to verify the transverse-Mercator implementation. Each
entry records a `(lon, lat)` input, the expected projected `(e, n)`, its SOSI
`koordsys` code, and the `source` it came from.

The current set was generated with **PROJ 9.8.1** (the reference implementation
of `etmerc`, accurate to nanometres), via `cs2cs`, over a grid of ±4° around
each central meridian at latitudes 55–70°N for the WGS84 UTM zones 31–36
(KOORDSYS 21–26). Our implementation agrees with these to ~0.05 mm — the limit
of the oracle's printed precision.

To regenerate (requires Nix):

```sh
# for each SOSI code / EPSG zone, feed "lon lat" lines to:
nix shell nixpkgs#proj -c cs2cs -d 4 \
  +proj=longlat +datum=WGS84 +to +init=epsg:32632   # EPSG:326xx per zone
```

Other independent implementations (e.g. `wroge/wgs84`) were tried as oracles but
use a lower-order series that drifts to decimetres far from the central
meridian, so PROJ is used as the authority.
