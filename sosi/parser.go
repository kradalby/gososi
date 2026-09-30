package sosi

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Parser handles SOSI file parsing following the JavaScript implementation pattern
type Parser struct {
	// Norwegian coordinate system mappings from JavaScript util.js
	coordSystems map[int]CoordinateSystem
}

// NewParser creates a new SOSI parser with Norwegian coordinate system mappings
func NewParser() *Parser {
	return &Parser{
		coordSystems: getNorwegianCoordinateSystems(),
	}
}

// Parse parses SOSI data from a reader and returns a SOSIDocument
func (p *Parser) Parse(reader io.Reader) (*SOSIDocument, error) {
	lines, err := readLines(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read SOSI data: %w", err)
	}

	// Clean and filter lines (JavaScript splitOnNewline equivalent)
	cleanedLines := p.splitOnNewline(lines)

	// Parse the hierarchical tree structure (JavaScript parseTree equivalent)
	tree, err := p.parseTree(cleanedLines, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SOSI tree structure: %w", err)
	}

	// Create SOSI document from parsed tree
	doc, err := p.createDocument(tree)
	if err != nil {
		return nil, fmt.Errorf("failed to create SOSI document: %w", err)
	}

	return doc, nil
}

// ParseFile parses a SOSI file by filename
func (p *Parser) ParseFile(filename string) (*SOSIDocument, error) {
	file, err := openFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open SOSI file %s: %w", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// In production, you might want to use a proper logger here
			// For now, we silently handle the close error to avoid masking the main error
			_ = closeErr
		}
	}()

	return p.Parse(file)
}

// splitOnNewline cleans SOSI lines by removing comments and trimming whitespace
// Equivalent to JavaScript splitOnNewline function (parser.js:36)
func (p *Parser) splitOnNewline(lines []string) []string {
	var cleaned []string

	for _, line := range lines {
		// Skip lines that start with ! (full comment lines)
		if strings.HasPrefix(strings.TrimSpace(line), "!") {
			continue
		}

		// Remove inline comments (everything after !)
		if idx := strings.Index(line, "!"); idx != -1 {
			line = line[:idx]
		}

		// Trim whitespace
		line = strings.TrimSpace(line)

		// Only add non-empty lines
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return cleaned
}

// parseTree recursively parses the hierarchical SOSI structure
// Equivalent to JavaScript parseTree function (util.js:70)
func (p *Parser) parseTree(lines []string, parentLevel int) (map[string][]string, error) {
	result := make(map[string][]string)
	var currentKey string

	for _, line := range lines {
		line = p.cleanupLine(line)

		if p.isParent(line, parentLevel) {
			key, err := p.getKey(line, parentLevel)
			if err != nil {
				return nil, fmt.Errorf("failed to extract key from line '%s': %w", line, err)
			}
			currentKey = key

			// Get values from the same line
			values := p.getValues(line)
			if values != "" {
				if result[currentKey] == nil {
					result[currentKey] = []string{}
				}
				result[currentKey] = append(result[currentKey], values)
			}
		} else if currentKey != "" && line != "" {
			// Add line to current key's values
			if result[currentKey] == nil {
				result[currentKey] = []string{}
			}
			result[currentKey] = append(result[currentKey], line)
		}
	}

	return result, nil
}

// cleanupLine removes comments and trailing whitespace
// Equivalent to JavaScript cleanupLine function (util.js:21)
func (p *Parser) cleanupLine(line string) string {
	// Remove comments
	if idx := strings.Index(line, "!"); idx != -1 {
		line = line[:idx]
	}

	// Remove trailing whitespace
	return strings.TrimRightFunc(line, func(r rune) bool {
		return r == ' ' || r == '\t'
	})
}

// isParent checks if a line is at the specified parent level
// Equivalent to JavaScript isParent function (util.js:62)
func (p *Parser) isParent(line string, parentLevel int) bool {
	// Must have the correct number of leading dots
	if p.countStartingDots(line) != parentLevel {
		return false
	}

	// For parentLevel 0 (top level), must start with a dot (like .PUNKT, .KURVE, .HODE)
	// This excludes coordinate lines that have 0 dots but don't start with '.'
	if parentLevel == 0 && !strings.HasPrefix(line, ".") {
		return false
	}

	return true
}

// countStartingDots counts leading dots to determine hierarchy level
// Equivalent to JavaScript countStartingDots function (util.js:51)
func (p *Parser) countStartingDots(line string) int {
	count := 0
	for _, char := range line {
		if char == '.' {
			count++
		} else {
			break
		}
	}
	return count
}

// getKey extracts the key from a SOSI line at the specified parent level
// Equivalent to JavaScript getKey function (util.js:28)
func (p *Parser) getKey(line string, parentLevel int) (string, error) {
	// Remove leading dots based on parent level
	numDots := p.getNumDots(parentLevel)
	if !strings.HasPrefix(line, numDots) {
		return "", fmt.Errorf("line does not have expected %d leading dots: %s", parentLevel, line)
	}

	line = strings.TrimPrefix(line, numDots)

	// Extract key (everything before : or first space)
	key := p.getKeyFromLine(line)

	return p.cleanupLine(key), nil
}

// getKeyFromLine extracts the key portion from a line
// Equivalent to JavaScript getKeyFromLine function (util.js:14)
func (p *Parser) getKeyFromLine(line string) string {
	if before, _, ok := strings.Cut(line, ":"); ok {
		return strings.TrimSpace(before)
	}

	// Split on first space and return first part
	parts := strings.Fields(line)
	if len(parts) > 0 {
		return parts[0]
	}

	return line
}

// getValues extracts values from a SOSI line (everything after the key)
// Equivalent to JavaScript getValues function (util.js:6)
func (p *Parser) getValues(line string) string {
	parts := strings.Fields(line)
	if len(parts) <= 1 {
		return ""
	}

	// Join all parts after the first one
	return strings.TrimSpace(strings.Join(parts[1:], " "))
}

// getNumDots creates a string with the specified number of dots
// Equivalent to JavaScript getNumDots function (util.js:10)
func (p *Parser) getNumDots(num int) string {
	return strings.Repeat(".", num)
}

// createDocument creates a SOSIDocument from the parsed tree structure
func (p *Parser) createDocument(tree map[string][]string) (*SOSIDocument, error) {
	doc := &SOSIDocument{}

	// Parse header (HODE section)
	if hodeData, exists := tree["HODE"]; exists || tree["HODE 0"] != nil {
		if tree["HODE 0"] != nil {
			hodeData = tree["HODE 0"]
		}

		var err error
		doc.Header, err = p.parseHeader(hodeData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SOSI header: %w", err)
		}
	} else {
		return nil, fmt.Errorf("missing required HODE (header) section")
	}

	// Parse features (everything except HODE, DEF, OBJDEF, SLUTT)
	features, err := p.parseFeatures(tree, doc.Header)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SOSI features: %w", err)
	}
	doc.Features = features

	// Calculate bounding box from features
	doc.Bounds = p.calculateBounds(features)

	return doc, nil
}

// parseHeader parses the SOSI header (HODE) section
func (p *Parser) parseHeader(hodeData []string) (SOSIHeader, error) {
	header := SOSIHeader{}

	// Parse header using level 2 parsing (equivalent to parseFromLevel2)
	headerMap, err := p.parseFromLevel2(hodeData)
	if err != nil {
		return header, fmt.Errorf("failed to parse header data: %w", err)
	}

	// Extract header fields
	header.CharacterSet = p.getString(headerMap, "TEGNSETT", "UTF-8")
	header.Producer = p.getString(headerMap, "PRODUSENT", "")
	header.Version = p.getString(headerMap, "SOSI-VERSJON", "")
	header.Level = p.getInt(headerMap, "SOSI-NIVÅ", 0)
	header.Owner = p.getString(headerMap, "EIER", "")
	header.ObjectCatalog = p.getString(headerMap, "OBJEKTKATALOG", "")
	header.VerificationDate = p.getString(headerMap, "VERIFISERINGSDATO", "")

	// Parse KVALITET (Quality) in header
	if kvalitetData, ok := headerMap["KVALITET"]; ok {
		if kvalitetStr, ok := kvalitetData.(string); ok {
			// Handle KVALITET as string (needs parsing)
			header.Quality = p.parseKvalitet(kvalitetStr)
		} else if kvalitetMap, ok := kvalitetData.(map[string]any); ok {
			// Handle KVALITET as pre-parsed map from tree parsing
			header.Quality = kvalitetMap
		}
	}

	// Parse TRANSPAR section
	if transpar, ok := headerMap["TRANSPAR"].(map[string]any); ok {
		header.CoordSystem = p.getInt(transpar, "KOORDSYS", 84)
		header.Unit = p.getFloat(transpar, "ENHET", 1.0)
		header.HeightUnit = p.getFloat(transpar, "ENHET-H", header.Unit)
		header.DepthUnit = p.getFloat(transpar, "ENHET-D", header.Unit)

		// Parse ORIGO-NØ
		if origoStr := p.getString(transpar, "ORIGO-NØ", "0 0"); origoStr != "" {
			header.Origo = p.parseOrigo(origoStr)
		}
	}

	// Parse OMRÅDE section
	if omrade, ok := headerMap["OMRÅDE"].(map[string]any); ok {
		header.Area = p.parseBoundingBox(omrade)
	}

	return header, nil
}

// parseFromLevel2 parses level-2 hierarchical SOSI data
// Equivalent to JavaScript parseFromLevel2 function (util.js:165)
func (p *Parser) parseFromLevel2(lines []string) (map[string]any, error) {
	tree, err := p.parseTree(lines, 2)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)

	for key, values := range tree {
		if len(values) == 0 {
			continue
		}

		// Check if this has sub-level data (starts with dots)
		hasSubLevel := false
		for _, value := range values {
			if strings.HasPrefix(value, ".") {
				hasSubLevel = true
				break
			}
		}

		if hasSubLevel {
			// Parse sub-dictionary
			subResult, err := p.parseSubdict(values)
			if err != nil {
				return nil, fmt.Errorf("failed to parse subdictionary for key %s: %w", key, err)
			}
			result[key] = subResult
		} else if len(values) == 1 {
			// Single value - convert to appropriate type
			result[key] = p.convertDataType(key, values[0])
		} else {
			// Multiple values - keep as array
			converted := make([]any, len(values))
			for i, value := range values {
				converted[i] = p.convertDataType(key, value)
			}
			result[key] = converted
		}
	}

	return result, nil
}

// parseSubdict parses sub-dictionary structures at level 3
// Equivalent to JavaScript parseSubdict function (util.js:150)
func (p *Parser) parseSubdict(lines []string) (map[string]any, error) {
	tree, err := p.parseTree(lines, 3)
	if err != nil {
		return nil, err
	}

	result := make(map[string]any)
	for key, values := range tree {
		if len(values) > 0 {
			result[key] = p.convertDataType(key, values[0])
		}
	}

	return result, nil
}

// Helper functions for data extraction
func (p *Parser) getString(data map[string]any, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case string:
			// Remove quotes if present
			return strings.Trim(v, `"'`)
		case int:
			return fmt.Sprintf("%d", v)
		case float64:
			return fmt.Sprintf("%.1f", v)
		}
	}
	return defaultValue
}

func (p *Parser) getInt(data map[string]any, key string, defaultValue int) int {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

func (p *Parser) getFloat(data map[string]any, key string, defaultValue float64) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		case int:
			return float64(v)
		}
	}
	return defaultValue
}

// convertDataType converts a string value to its appropriate Go type
// Enhanced implementation with SOSI data type support
func (p *Parser) convertDataType(key, value string) any {
	// Special handling for KVALITET attributes
	if key == "KVALITET" {
		return p.parseKvalitet(value)
	}

	// Special handling for REGISTRERINGSVERSJON attributes
	if key == "REGISTRERINGSVERSJON" {
		return p.parseRegistreringsversjon(value)
	}

	// Special handling for date fields - keep as strings for now
	// Future enhancement: convert to time.Time objects
	dateFields := map[string]bool{
		"OPPDATERINGSDATO":  true,
		"DATAFANGSTDATO":    true,
		"AJOURFØRTDATO":     true,
		"DATO":              true,
		"KOPIDATO":          true,
		"VERIFISERINGSDATO": true,
	}

	if dateFields[key] {
		return value // Keep dates as strings for now
	}

	// Try to convert to number
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}

	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}

	// Remove quotes if present
	value = strings.Trim(value, `"'`)

	return value
}

// parseKvalitet parses KVALITET attribute values into structured data
// KVALITET format: "målemetode [nøyaktighet] [måleskala] [*] [*]"
// Example: "82" -> {målemetode: 82}, "40 58" -> {målemetode: 40, nøyaktighet: 58}
func (p *Parser) parseKvalitet(value string) map[string]any {
	fields := strings.Fields(value)
	kvalitet := make(map[string]any)

	if len(fields) >= 1 {
		// Parse measurement method (målemetode) - always present
		if målemetode, err := strconv.Atoi(fields[0]); err == nil {
			kvalitet["målemetode"] = målemetode
		}
	}

	if len(fields) >= 2 {
		// Parse accuracy (nøyaktighet) - optional second field
		if nøyaktighet, err := strconv.Atoi(fields[1]); err == nil {
			kvalitet["nøyaktighet"] = nøyaktighet
		}
	}

	if len(fields) >= 3 {
		// Parse measurement scale (måleskala) - optional third field
		if måleskala, err := strconv.Atoi(fields[2]); err == nil {
			kvalitet["måleskala"] = måleskala
		}
	}

	// Additional KVALITET fields can be added here as needed
	// fields[3] and beyond may contain additional quality indicators

	return kvalitet
}

// parseRegistreringsversjon parses REGISTRERINGSVERSJON attribute values
// REGISTRERINGSVERSJON format: "system" "version"
// Example: "FKB" "3.4 eller eldre" -> {system: "FKB", versjon: "3.4 eller eldre"}
func (p *Parser) parseRegistreringsversjon(value string) map[string]any {
	regVers := make(map[string]any)

	// Parse quoted strings from the value
	var parts []string
	current := ""
	inQuote := false
	escapeNext := false

	for _, char := range value {
		if escapeNext {
			current += string(char)
			escapeNext = false
		} else if char == '\\' {
			escapeNext = true
		} else if char == '"' {
			if inQuote {
				// End of quoted string
				parts = append(parts, current)
				current = ""
				inQuote = false
			} else {
				// Start of quoted string
				inQuote = true
			}
		} else if inQuote {
			current += string(char)
		}
		// Skip whitespace outside quotes
	}

	// Add any remaining content
	if current != "" {
		parts = append(parts, current)
	}

	// Map parts to structured data
	if len(parts) >= 1 {
		regVers["system"] = parts[0]
	}

	if len(parts) >= 2 {
		regVers["versjon"] = parts[1]
	}

	return regVers
}

// parseOrigo parses the ORIGO-NØ coordinate string
func (p *Parser) parseOrigo(origoStr string) Coordinate {
	parts := strings.Fields(origoStr)
	if len(parts) >= 2 {
		y, _ := strconv.ParseFloat(parts[0], 64) // N (latitude)
		x, _ := strconv.ParseFloat(parts[1], 64) // Ø (longitude)
		return Coordinate{X: x, Y: y, Z: 0}
	}
	return Coordinate{}
}

// parseBoundingBox parses the OMRÅDE section into a bounding box
func (p *Parser) parseBoundingBox(omrade map[string]any) BoundingBox {
	bbox := BoundingBox{}

	if minStr := p.getString(omrade, "MIN-NØ", ""); minStr != "" {
		parts := strings.Fields(minStr)
		if len(parts) >= 2 {
			bbox.MinLat, _ = strconv.ParseFloat(parts[0], 64)
			bbox.MinLon, _ = strconv.ParseFloat(parts[1], 64)
		}
	}

	if maxStr := p.getString(omrade, "MAX-NØ", ""); maxStr != "" {
		parts := strings.Fields(maxStr)
		if len(parts) >= 2 {
			bbox.MaxLat, _ = strconv.ParseFloat(parts[0], 64)
			bbox.MaxLon, _ = strconv.ParseFloat(parts[1], 64)
		}
	}

	return bbox
}

// parseFeatures parses all features from the SOSI tree
func (p *Parser) parseFeatures(tree map[string][]string, header SOSIHeader) ([]SOSIFeature, error) {
	var features []SOSIFeature

	// Skip non-feature sections
	skipKeys := map[string]bool{
		"HODE":   true,
		"HODE 0": true,
		"DEF":    true,
		"OBJDEF": true,
		"SLUTT":  true,
	}

	// Get keys and sort by feature ID for deterministic parsing order
	type keyWithID struct {
		key string
		id  int
	}
	var keyData []keyWithID

	for key := range tree {
		if !skipKeys[key] {
			// Extract ID from key (e.g., "PUNKT 1" -> 1)
			var id int
			parts := strings.Fields(key)
			if len(parts) >= 2 {
				if parsedID, err := strconv.Atoi(parts[1]); err == nil {
					id = parsedID
				}
			}
			keyData = append(keyData, keyWithID{key: key, id: id})
		}
	}

	// Sort by ID to ensure consistent parsing order
	sort.Slice(keyData, func(i, j int) bool {
		return keyData[i].id < keyData[j].id
	})

	// Parse features in deterministic order
	for _, keyItem := range keyData {
		key := keyItem.key
		lines := tree[key]

		// Parse feature key (e.g., "PUNKT 123" -> type="PUNKT", id=123)
		feature, err := p.parseFeature(key, lines, header)
		if err != nil {
			return nil, fmt.Errorf("failed to parse feature %s: %w", key, err)
		}

		features = append(features, feature)
	}

	return features, nil
}

// parseFeature parses a single SOSI feature
func (p *Parser) parseFeature(key string, lines []string, header SOSIHeader) (SOSIFeature, error) {
	// Parse feature header (e.g., "PUNKT 123")
	keyParts := strings.Fields(strings.ReplaceAll(key, ":", ""))
	if len(keyParts) < 2 {
		return SOSIFeature{}, fmt.Errorf("invalid feature key format: %s", key)
	}

	geometryType := keyParts[0]
	id, err := strconv.Atoi(keyParts[1])
	if err != nil {
		return SOSIFeature{}, fmt.Errorf("invalid feature ID in key %s: %w", key, err)
	}

	// Skip first line as it's already processed in the key
	if len(lines) <= 1 {
		return SOSIFeature{}, fmt.Errorf("feature has no data lines beyond header")
	}
	featureLines := lines[1:]

	// Parse feature data using level 2 parsing
	featureData, err := p.parseFromLevel2(featureLines)
	if err != nil {
		return SOSIFeature{}, fmt.Errorf("failed to parse feature data: %w", err)
	}

	var objectType string
	if objType, ok := featureData["OBJTYPE"]; ok {
		objectType = fmt.Sprintf("%v", objType)
	}

	geom, err := p.parseGeometry(geometryType, featureData, header)
	if err != nil {
		return SOSIFeature{}, fmt.Errorf("failed to parse geometry: %w", err)
	}

	return SOSIFeature{
		ID:          id,
		Type:        geometryType,
		ObjectType:  objectType,
		Coordinates: geom.coords,
		Properties:  featureData,
		Refs:        geom.refs,
		OuterRing:   geom.outerRing,
		Holes:       geom.holes,
	}, nil
}

// geometry is the shape parsed from a feature's coordinate and REF keys.
type geometry struct {
	coords    []Coordinate
	refs      []int
	outerRing []int
	holes     [][]int
}

// parseGeometry parses geometry data for a feature
func (p *Parser) parseGeometry(geometryType string, data map[string]any, header SOSIHeader) (geometry, error) {
	var (
		coords []Coordinate
		err    error
	)
	switch geometryType {
	case "PUNKT", "TEKST":
		// TEKST is treated as a point with additional text attributes
		coords, err = p.parsePointGeometry(data, header)
	case "KURVE":
		coords, err = p.parseLineStringGeometry(data, header)
	case "FLATE":
		return p.parsePolygonGeometry(data, header)
	case "BUEP":
		coords, err = p.parseArcGeometry(data, header)
	default:
		return geometry{}, fmt.Errorf("unsupported geometry type: %s", geometryType)
	}
	if err != nil {
		return geometry{}, err
	}
	return geometry{coords: coords}, nil
}

// coordLines returns the coordinate lines stored under a NØ/NØH key: the
// level-2 parse keeps one line as a string and several as []any.
func coordLines(value any) []string {
	switch v := value.(type) {
	case string:
		return []string{v}
	case []any:
		lines := make([]string, 0, len(v))
		for _, line := range v {
			if s, ok := line.(string); ok {
				lines = append(lines, s)
			}
		}
		return lines
	}
	return nil
}

// parseCoordinates parses every line, labelling errors with what.
func (p *Parser) parseCoordinates(lines []string, header SOSIHeader, what string) ([]Coordinate, error) {
	coords := make([]Coordinate, 0, len(lines))
	for _, line := range lines {
		coord, err := p.parseCoordinate(line, header)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s coordinate: %w", what, err)
		}
		coords = append(coords, coord)
	}
	return coords, nil
}

// parsePointGeometry parses PUNKT (Point) geometry
func (p *Parser) parsePointGeometry(data map[string]any, header SOSIHeader) ([]Coordinate, error) {
	for _, key := range []string{"NØ", "NØH"} {
		if lines := coordLines(data[key]); len(lines) > 0 {
			return p.parseCoordinates(lines[:1], header, "point")
		}
	}
	return nil, nil
}

// parseLineStringGeometry parses KURVE (LineString) geometry
func (p *Parser) parseLineStringGeometry(data map[string]any, header SOSIHeader) ([]Coordinate, error) {
	// Collect coordinates from all available coordinate keys
	var coords []Coordinate
	for _, key := range []string{"NØ", "NØH"} {
		c, err := p.parseCoordinates(coordLines(data[key]), header, "linestring")
		if err != nil {
			return nil, err
		}
		coords = append(coords, c...)
	}
	return coords, nil
}

// parsePolygonGeometry parses FLATE (Polygon) geometry: REF lines, plus either
// a centre point or, for simple polygons, the ring itself under NØ or NØH.
func (p *Parser) parsePolygonGeometry(data map[string]any, header SOSIHeader) (geometry, error) {
	var geom geometry

	// REF can be a single string or array of strings (multi-line REF)
	if refData, ok := data["REF"]; ok {
		var allRefStrings []string

		switch refs := refData.(type) {
		case string:
			allRefStrings = []string{refs}
		case []any:
			for _, ref := range refs {
				if refStr, ok := ref.(string); ok {
					allRefStrings = append(allRefStrings, refStr)
				}
			}
		}

		for _, refStr := range allRefStrings {
			outerRing, holes, refs, err := p.parsePolygonReferencesWithHoles(refStr)
			if err != nil {
				return geometry{}, fmt.Errorf("failed to parse polygon references: %w", err)
			}
			geom.outerRing = append(geom.outerRing, outerRing...)
			geom.holes = append(geom.holes, holes...)
			// All references flattened, for backward compatibility
			geom.refs = append(geom.refs, refs...)
		}
	}

	// Only the first coordinate key present is used
	for _, key := range []string{"NØ", "NØH"} {
		if lines := coordLines(data[key]); len(lines) > 0 {
			coords, err := p.parseCoordinates(lines, header, "polygon")
			if err != nil {
				return geometry{}, err
			}
			geom.coords = coords
			break
		}
	}

	return geom, nil
}

// parseArcGeometry parses BUEP (Arc) geometry
func (p *Parser) parseArcGeometry(data map[string]any, header SOSIHeader) ([]Coordinate, error) {
	// BUEP requires 3 points to define the arc
	for _, key := range []string{"NØ", "NØH"} {
		switch coords := data[key].(type) {
		case []any:
			if len(coords) != 3 {
				return nil, fmt.Errorf("BUEP requires exactly 3 points, got %d", len(coords))
			}
			return p.parseCoordinates(coordLines(coords), header, "arc")
		case string:
			return nil, fmt.Errorf("BUEP requires 3 coordinate points, got single coordinate")
		}
	}

	return nil, fmt.Errorf("no coordinates found for BUEP geometry")
}

// parseCoordinate parses a coordinate string in SOSI format
// Also handles tie points (knutepunkter) marked with ...KP
func (p *Parser) parseCoordinate(coordStr string, header SOSIHeader) (Coordinate, error) {
	// Check for tie point marker (...KP)
	hasTiepoint := strings.Contains(coordStr, "...KP")
	var tiepointCode int

	if hasTiepoint {
		// Extract tiepoint code
		if kpIndex := strings.Index(coordStr, "...KP"); kpIndex != -1 {
			// Get the part after ...KP
			kpPart := strings.TrimSpace(coordStr[kpIndex+5:])
			kpFields := strings.Fields(kpPart)
			if len(kpFields) > 0 {
				if code, err := strconv.Atoi(kpFields[0]); err == nil {
					tiepointCode = code
				}
			}
			// Remove the KP part for coordinate parsing
			coordStr = strings.TrimSpace(coordStr[:kpIndex])
		}
	}

	parts := strings.Fields(coordStr)
	if len(parts) < 2 {
		return Coordinate{}, fmt.Errorf("coordinate requires at least N and Ø values: %s", coordStr)
	}

	// Parse N (North/Latitude) and Ø (East/Longitude)
	n, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return Coordinate{}, fmt.Errorf("invalid N coordinate: %s", parts[0])
	}

	o, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return Coordinate{}, fmt.Errorf("invalid Ø coordinate: %s", parts[1])
	}

	coord := Coordinate{
		Y:            n*header.Unit + header.Origo.Y, // N = Latitude (Y)
		X:            o*header.Unit + header.Origo.X, // Ø = Longitude (X)
		TiePointCode: tiepointCode,                   // Store tie point code
	}

	// Parse H (Height) if present
	if len(parts) > 2 {
		h, err := strconv.ParseFloat(parts[2], 64)
		if err == nil {
			coord.Z = h * header.HeightUnit
		}
	}

	return coord, nil
}

// parseReferences parses a reference string into feature IDs
// SOSI reference format: ":633" (positive), ":-633" (negative/reverse)
func (p *Parser) parseReferences(refStr string) ([]int, error) {
	var refs []int

	// Split on spaces and parse each reference
	parts := strings.FieldsSeq(strings.TrimSpace(refStr))
	for part := range parts {
		// Skip parentheses - they indicate hole references which are handled separately
		if strings.HasPrefix(part, "(") || strings.HasSuffix(part, ")") {
			continue
		}

		// Remove leading colon if present
		part = strings.TrimPrefix(part, ":")

		// Handle empty parts (can happen with line breaks in REF)
		if part == "" {
			continue
		}

		// Parse as integer (handles negative numbers automatically)
		ref, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid reference ID: %s", part)
		}

		refs = append(refs, ref)
	}

	return refs, nil
}

// parsePolygonReferencesWithHoles parses polygon references including holes
// Format: :100 :101 :102 :103 (:500) or :100 :101 :102 :103 (:200 :201 :202 :203)
func (p *Parser) parsePolygonReferencesWithHoles(refStr string) (outerRing []int, holes [][]int, allRefs []int, err error) {
	// Find hole references within parentheses
	holePattern := `\(([^)]+)\)`
	re := regexp.MustCompile(holePattern)
	holeMatches := re.FindAllStringSubmatch(refStr, -1)

	// Remove hole references from the string to get outer ring references
	outerRefStr := re.ReplaceAllString(refStr, "")

	// Parse outer ring references
	outerRing, err = p.parseReferences(outerRefStr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse outer ring references: %w", err)
	}

	// Parse hole references
	for _, match := range holeMatches {
		holeRefStr := match[1] // The content inside parentheses
		holeRefs, err := p.parseReferences(holeRefStr)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse hole references: %w", err)
		}
		holes = append(holes, holeRefs)
	}

	// Combine all references for backward compatibility
	allRefs = append(allRefs, outerRing...)
	for _, hole := range holes {
		allRefs = append(allRefs, hole...)
	}

	return outerRing, holes, allRefs, nil
}

// calculateBounds calculates the bounding box from all features
func (p *Parser) calculateBounds(features []SOSIFeature) BoundingBox {
	if len(features) == 0 {
		return BoundingBox{}
	}

	// Collect all coordinates from all features
	var allCoords []Coordinate
	for _, feature := range features {
		allCoords = append(allCoords, feature.Coordinates...)
	}

	// Use the existing CalculateBoundingBox function to avoid duplication
	return CalculateBoundingBox(allCoords)
}

// Helper functions for file operations
func readLines(reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)

	firstLine := true
	for scanner.Scan() {
		line := scanner.Text()

		// Strip BOM from first line if present
		if firstLine {
			firstLine = false
			// Remove UTF-8 BOM (EF BB BF) if present
			if len(line) > 0 && strings.HasPrefix(line, "\uFEFF") {
				line = strings.TrimPrefix(line, "\uFEFF")
			}
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading lines: %w", err)
	}

	return lines, nil
}

// openFile opens a file for reading
func openFile(filename string) (io.ReadCloser, error) {
	return os.Open(filename)
}
