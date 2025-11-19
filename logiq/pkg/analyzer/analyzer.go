package analyzer

import (
	"fmt"
	"logiq/pkg/config"
	"logiq/pkg/scanner"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Anomaly struct {
	Pattern   string
	Count     int
	Severity  string
	Timestamp string
	Spike     bool
	Files     map[string][]int // file -> line numbers
	Lines     []string         // sample lines
}

func AnalyzeLogs(cfg *config.Config, logs []scanner.LogLine) []Anomaly {
	patternCounts := make(map[string]int)
	anomalyFiles := make(map[string]map[string][]int) // pattern -> file -> lines
	anomalyLines := make(map[string][]string)         // pattern -> sample lines
	anomalies := []Anomaly{}

	if len(logs) == 0 {
		return anomalies
	}

	// Precompile alert rules from config
	for _, rule := range cfg.AlertRules {
		re, err := regexp.Compile(rule.Keyword)
		if err != nil {
			fmt.Printf("Skipping invalid regex: %s\n", rule.Keyword)
			continue
		}

		// Initialize maps
		if _, exists := anomalyFiles[rule.Keyword]; !exists {
			anomalyFiles[rule.Keyword] = make(map[string][]int)
		}
		if _, exists := anomalyLines[rule.Keyword]; !exists {
			anomalyLines[rule.Keyword] = []string{}
		}

		// Scan logs for matches
		for _, line := range logs {
			if re.MatchString(line.Content) {
				patternCounts[rule.Keyword]++

				// Track file and line numbers
				anomalyFiles[rule.Keyword][line.FilePath] = append(anomalyFiles[rule.Keyword][line.FilePath], int(line.LineNum))

				anomalyLines[rule.Keyword] = append(anomalyLines[rule.Keyword], fmt.Sprintf("[%d] %s", line.LineNum, line.Content))
			}
		}
	}

	// Convert to Anomaly objects
	totalLines := float64(len(logs))
	threshold := cfg.Analysis.ThresholdSpike

	for pattern, count := range patternCounts {
		severity := getSeverity(cfg.AlertRules, pattern)
		ratio := float64(count) / totalLines
		spike := ratio > threshold

		anomalies = append(anomalies, Anomaly{
			Pattern:   pattern,
			Count:     count,
			Severity:  severity,
			Timestamp: time.Now().Format(time.RFC3339),
			Spike:     spike,
			Files:     anomalyFiles[pattern],
			Lines:     anomalyLines[pattern],
		})
	}

	// Sort by frequency (highest first)
	sort.Slice(anomalies, func(i, j int) bool {
		return anomalies[i].Count > anomalies[j].Count
	})

	return anomalies
}

func getSeverity(rules []config.AlertRule, keyword string) string {
	for _, r := range rules {
		if strings.EqualFold(r.Keyword, keyword) {
			return r.Severity
		}
	}
	return "info"
}
