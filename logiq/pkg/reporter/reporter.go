package reporter

import (
	"encoding/json"
	"fmt"
	"html/template"
	"logiq/pkg/analyzer"
	"logiq/pkg/config"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateReport creates text, JSON, and/or HTML reports depending on config
func GenerateReport(anomalies []analyzer.Anomaly, cfg config.OutputConfig) error {
	if len(anomalies) == 0 {
		fmt.Println("✅ No anomalies found.")
		return nil
	}

	// Ensure the report directory exists
	reportDir := filepath.Dir(cfg.ReportFile)
	if reportDir != "." {
		if err := os.MkdirAll(reportDir, 0755); err != nil {
			return fmt.Errorf("failed to create report directory: %w", err)
		}
	}

	// Support multiple formats (e.g., "json,html" or "all")
	formats := strings.Split(strings.ToLower(cfg.Format), ",")
	for _, f := range formats {
		switch strings.TrimSpace(f) {
		case "json":
			if err := generateJSONReport(cfg.ReportFile, anomalies); err != nil {
				return err
			}
		case "html":
			if err := generateHTMLReport(cfg.ReportFile, anomalies); err != nil {
				return err
			}
		case "all":
			if err := generateJSONReport(cfg.ReportFile, anomalies); err != nil {
				return err
			}
			if err := generateHTMLReport(cfg.ReportFile, anomalies); err != nil {
				return err
			}
		}
	}

	return nil
}

// ------------------ JSON Report ------------------

func generateJSONReport(baseFile string, anomalies []analyzer.Anomaly) error {
	jsonFile := baseFile
	if filepath.Ext(jsonFile) != ".json" {
		jsonFile = jsonFile + ".json"
	}

	jsonData, err := json.MarshalIndent(anomalies, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON report: %w", err)
	}

	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON report: %w", err)
	}

	fmt.Printf("📊 JSON report saved as: %s\n", jsonFile)
	return nil
}

// ------------------ HTML Report ------------------

func generateHTMLReport(baseFile string, anomalies []analyzer.Anomaly) error {
	htmlFile := baseFile
	if filepath.Ext(htmlFile) != ".html" {
		htmlFile = htmlFile + ".html"
	}

	const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>LogIQ Report</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 20px; }
    h1 { color: #333; }
    table { border-collapse: collapse; width: 100%; margin-top: 20px; }
    th, td { border: 1px solid #ddd; padding: 8px; }
    th { background-color: #f4f4f4; }
    tr:nth-child(even) { background-color: #f9f9f9; }
    .low { color: green; }
    .medium { color: orange; }
    .high { color: red; font-weight: bold; }
  </style>
</head>
<body>
  <h1>LogIQ Anomaly Report</h1>
  <p>Generated at {{ .Timestamp }}</p>
  <table>
    <tr>
      <th>Pattern</th>
      <th>Count</th>
      <th>Severity</th>
      <th>Spike</th>
      <th>Files</th>
      <th>Sample Lines</th>
      <th>Timestamp</th>
    </tr>
    {{ range .Anomalies }}
    <tr>
      <td>{{ .Pattern }}</td>
      <td>{{ .Count }}</td>
      <td class="{{ .Severity | ToLower }}">{{ .Severity }}</td>
      <td>{{ .Spike }}</td>
      <td>
        {{ range $file, $lines := .Files }}
          <strong>{{ $file }}</strong>: {{ $lines }}<br>
        {{ end }}
      </td>
      <td>{{ join .Lines " | " }}</td>
      <td>{{ .Timestamp }}</td>
    </tr>
    {{ end }}
  </table>
</body>
</html>
`

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
		"join":    strings.Join,
	}

	tmpl, err := template.New("report").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	data := struct {
		Timestamp string
		Anomalies []analyzer.Anomaly
	}{
		Timestamp: time.Now().Format(time.RFC1123),
		Anomalies: anomalies,
	}

	file, err := os.Create(htmlFile)
	if err != nil {
		return fmt.Errorf("failed to create HTML report: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	fmt.Printf("🌐 HTML report saved as: %s\n", htmlFile)
	return nil
}
