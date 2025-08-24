package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// displayAnalysisSummary shows a summary of the analyzed GeoJSON features
func displayAnalysisSummary(analysis map[string]string) {
	if len(analysis) == 0 {
		fmt.Println("No features found in file")
		return
	}

	// Count geometry types for summary
	geometryTypes := make(map[string]int)
	for _, desc := range analysis {
		// Extract geometry type from description (format: "id: label GeometryType = X coords")
		parts := strings.Fields(desc)
		for _, part := range parts {
			if strings.Contains(part, "Point") || strings.Contains(part, "LineString") || 
			   strings.Contains(part, "Polygon") || strings.Contains(part, "Multi") {
				geometryTypes[part]++
				break
			}
		}
	}

	// Display summary
	fmt.Printf("Found %d features:", len(analysis))
	var types []string
	for geomType, count := range geometryTypes {
		types = append(types, fmt.Sprintf("%d %s", count, geomType))
	}
	
	if len(types) > 0 {
		fmt.Printf(" %s\n", strings.Join(types, ", "))
	} else {
		fmt.Println()
	}
	fmt.Println()
}

// promptForObjectTypes interactively prompts the user for SOSI object types
func promptForObjectTypes(analysis map[string]string) (map[string]string, error) {
	objectTypes := make(map[string]string)

	// Sort feature IDs for consistent ordering
	var featureIDs []string
	for id := range analysis {
		featureIDs = append(featureIDs, id)
	}
	sort.Strings(featureIDs)

	// Prompt for each feature
	for _, featureID := range featureIDs {
		featureInfo := analysis[featureID]
		
		for {
			// Display feature information and prompt for object type
			objectType, err := readLineWithPrompt(featureInfo)
			if err != nil {
				return nil, fmt.Errorf("error reading input: %w", err)
			}

			objectType = strings.TrimSpace(objectType)
			if objectType != "" {
				objectTypes[featureID] = objectType
				break
			} else {
				fmt.Println("Please enter a valid object type.")
			}
		}
	}

	return objectTypes, nil
}

// readLineWithPrompt reads a line from stdin with a prompt message
func readLineWithPrompt(message string) (string, error) {
	fmt.Printf("%s\nObject type: ", message)
	
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		// EOF reached
		return "", fmt.Errorf("input terminated")
	}
	
	return scanner.Text(), nil
}