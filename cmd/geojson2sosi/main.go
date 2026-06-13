// Package main provides the geojson2sosi CLI tool for converting between GeoJSON and SOSI formats.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/creachadair/command"
	"github.com/kradalby/gososi/sosi"
)

func main() {
	var outputFile string

	root := &command.C{
		Name:  "geojson2sosi",
		Usage: "input_file [output_file]",
		Help:  "Convert between GeoJSON and SOSI formats bidirectionally",

		SetFlags: func(env *command.Env, fs *flag.FlagSet) {
			fs.StringVar(&outputFile, "output", "", "Output file path (optional)")
			fs.StringVar(&outputFile, "o", "", "Output file path (optional, shorthand)")
		},

		Run: func(env *command.Env) error {
			args := env.Args
			if len(args) < 1 {
				return command.UsageError{Env: env, Message: "missing required input file"}
			}

			inputFile := args[0]
			finalOutputFile := ""

			if len(args) >= 2 {
				finalOutputFile = args[1]
			} else if outputFile != "" {
				finalOutputFile = outputFile
			} else {
				finalOutputFile = generateOutputFilename(inputFile)
			}

			// Validate input file
			if err := validateInputFile(inputFile); err != nil {
				return fmt.Errorf("input validation failed: %w", err)
			}

			// Determine conversion direction based on file extension
			inputExt := strings.ToLower(filepath.Ext(inputFile))

			switch inputExt {
			case ".geojson", ".json":
				return convertGeoJSONtoSOSI(inputFile, finalOutputFile)
			case ".sos", ".sosi":
				return convertSOSItoGeoJSON(inputFile, finalOutputFile)
			default:
				return fmt.Errorf("unsupported input file format: %s. Supported formats: .geojson, .json, .sos, .sosi", inputExt)
			}
		},
	}

	env := &command.Env{Command: root}
	command.RunOrFail(env, os.Args[1:])
}

// convertGeoJSONtoSOSI handles GeoJSON to SOSI conversion
func convertGeoJSONtoSOSI(inputFile, outputFile string) error {
	fmt.Printf("Converting GeoJSON to SOSI: %s -> %s\n", inputFile, outputFile)

	// Read GeoJSON file
	geojsonData, err := readFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read GeoJSON file: %w", err)
	}

	// Analyze GeoJSON to get feature information
	analysis, err := sosi.AnalyzeGeoJSON(geojsonData)
	if err != nil {
		return fmt.Errorf("failed to analyze GeoJSON: %w", err)
	}

	// Display analysis
	displayAnalysisSummary(analysis)

	// Prompt for object types
	objectTypes, err := promptForObjectTypes(analysis)
	if err != nil {
		return fmt.Errorf("failed to get object types: %w", err)
	}

	// Convert to SOSI
	sosiData, err := sosi.GeoJSONtoSOSI(geojsonData, objectTypes)
	if err != nil {
		return fmt.Errorf("failed to convert to SOSI: %w", err)
	}

	// Write output
	if err := writeFile(sosiData, outputFile); err != nil {
		return fmt.Errorf("failed to write SOSI file: %w", err)
	}

	fmt.Printf("Successfully converted to SOSI (%d bytes)\n", len(sosiData))
	return nil
}

// convertSOSItoGeoJSON handles SOSI to GeoJSON conversion
func convertSOSItoGeoJSON(inputFile, outputFile string) error {
	fmt.Printf("Converting SOSI to GeoJSON: %s -> %s\n", inputFile, outputFile)

	// Read SOSI file
	sosiData, err := readFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read SOSI file: %w", err)
	}

	// Convert to GeoJSON
	geojsonData, err := sosi.SOSItoGeoJSON(sosiData)
	if err != nil {
		return fmt.Errorf("failed to convert to GeoJSON: %w", err)
	}

	// Write output
	if err := writeFile(geojsonData, outputFile); err != nil {
		return fmt.Errorf("failed to write GeoJSON file: %w", err)
	}

	fmt.Printf("Successfully converted to GeoJSON (%d bytes)\n", len(geojsonData))
	return nil
}

// generateOutputFilename creates output filename from input filename
func generateOutputFilename(inputFile string) string {
	ext := strings.ToLower(filepath.Ext(inputFile))
	base := strings.TrimSuffix(inputFile, ext)

	switch ext {
	case ".geojson", ".json":
		return base + ".sos"
	case ".sos", ".sosi":
		return base + ".geojson"
	default:
		return inputFile + ".converted"
	}
}

// validateInputFile checks if the input file exists and is readable
func validateInputFile(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("input file path cannot be empty")
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file '%s' does not exist", filePath)
		}
		return fmt.Errorf("cannot access file '%s': %w", filePath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a file", filePath)
	}

	if info.Size() == 0 {
		return fmt.Errorf("file '%s' is empty", filePath)
	}

	// Add reasonable file size limit to prevent DoS attacks (100MB)
	const maxFileSize = 100 * 1024 * 1024
	if info.Size() > maxFileSize {
		return fmt.Errorf("file '%s' is too large (%d bytes), maximum allowed is %d bytes", filePath, info.Size(), maxFileSize)
	}

	return nil
}

// readFile reads a file and returns its contents
func readFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}
	return data, nil
}

// writeFile writes data to a file
func writeFile(data []byte, filePath string) error {
	// Clean and validate the file path to prevent path traversal attacks
	cleanPath := filepath.Clean(filePath)
	if strings.Contains(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		return fmt.Errorf("invalid output path '%s': path traversal detected or absolute path not allowed", filePath)
	}

	// Ensure output directory exists
	dir := filepath.Dir(cleanPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create output directory '%s': %w", dir, err)
		}
	}

	err := os.WriteFile(cleanPath, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write file '%s': %w", cleanPath, err)
	}

	return nil
}
