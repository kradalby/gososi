# gososi

A Go library for [SOSI](https://en.wikipedia.org/wiki/SOSI), the Norwegian
national standard for geographic data, with GeoJSON conversion and a
dependency-free coordinate-projection package.

## Packages

- **`sosi`** — read SOSI, write SOSI, and convert to/from GeoJSON.
- **`proj`** — pure-Go transverse-Mercator / UTM projection (Karney-Krüger n⁶),
  reusable on its own. Verified against PROJ 9.8.1 to ~0.05 mm.

## GeoJSON → SOSI

`sosi.GeoJSONToSOSI` produces Norwegian-government-compliant SOSI by default: the
data is reprojected into an auto-selected UTM zone (KOORDSYS 21–26), heights are
referenced to NN2000, coordinates are written at centimetre precision, and
GeoJSON properties become SOSI attributes.

```go
data, _ := os.ReadFile("pipes.geojson")

// Compliant defaults: auto UTM zone, NN2000, SOSI 5.0, cm precision, attributes.
out, err := sosi.GeoJSONToSOSI(data, map[string]string{"tube-56": "Roer"})
```

Behaviour is configured with typed options (in the style of `encoding/json/v2`):

```go
out, err := sosi.GeoJSONToSOSI(data, objectTypes,
    sosi.WithKoordSys(sosi.KoordSysUTM33), // force UTM zone 33 (code 23)
    sosi.WithVertDatum(sosi.VertDatumNN2000),
    sosi.WithSOSIVersion(sosi.SOSIVersion34),
    sosi.WithSOSILevel(sosi.SOSILevel4),
    sosi.WithProducer("Acme AS"),
)

// Legacy variety: raw WGS84 lon/lat, no projection, no attributes.
out, err = sosi.GeoJSONToSOSI(data, objectTypes,
    sosi.WithKoordSys(sosi.KoordSysWGS84Geographic),
    sosi.WithAccuracy(7, 3),
    sosi.WithVertDatum(sosi.VertDatumNone),
    sosi.WithAttributes(false),
)
```

The coordinate system is a typed enum, so invalid codes are hard to express:
`KoordSysUTM31`…`KoordSysUTM36` (SOSI codes 21–26) and `KoordSysWGS84Geographic`
(84). The value written to the `...KOORDSYS` directive is the SOSI *code*, not
the UTM zone number (e.g. `KoordSysUTM33` → `...KOORDSYS 23`).

## Projection

```go
crs, _ := proj.FromKOORDSYS(proj.KoordSysUTM32) // EPSG:32632
e, n := crs.Forward(10.210588982, 59.171243033) // → 569203.04, 6559748.10
lon, lat := crs.Inverse(e, n)
```

Legacy datums (NGO1948 codes 1–9, ED50 codes 31–36) need a Helmert transform and
return `proj.ErrUnsupportedDatum`; the API leaves room to add them.

## CLI

The `geojson2sosi` tool converts by file extension and prompts for each
feature's object type.

```sh
# Compliant conversion (auto-selects the UTM zone):
geojson2sosi pipes.geojson pipes.sos

# Pin the coordinate system and SOSI version:
geojson2sosi pipes.geojson pipes.sos --koordsys 22 --sosi-versjon 3.4

# Raw WGS84 lon/lat, no attributes:
geojson2sosi pipes.geojson out.sos --koordsys 84 --latlon-decimals 7 --attrs=false

# SOSI → GeoJSON (reverse):
geojson2sosi pipes.sos pipes.geojson
```

Flags: `--koordsys`, `--vert-datum`, `--sosi-versjon`, `--sosi-niva`,
`--producer`, `--attrs`, `--latlon-decimals`, `--alt-decimals`.

## Scope

Reprojection currently runs on the **forward** path (GeoJSON → SOSI). The reverse
path emits coordinates in whatever system the SOSI file declares without
reprojecting to WGS84 — a documented follow-up.
