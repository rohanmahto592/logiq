package jsonConvertor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"logiq/pkg/scanner"
)

var outputDir = "jsonLogs"

// SaveAsJSONArray clears the output directory and writes fresh JSON data for the current day.
func SaveAsJSONArray(logs []scanner.LogLine) error {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Delete all existing files inside the directory
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(outputDir, entry.Name()))
		}
	}

	// Create new file for today's logs
	filename := filepath.Join(outputDir, time.Now().Format("2006-01-02")+"_logs.json")
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// Stream encode JSON (efficient and memory-safe)
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(logs)
}
