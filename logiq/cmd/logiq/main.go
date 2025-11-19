package main

import (
	"flag"
	"fmt"
	"logiq/pkg/analyzer"
	"logiq/pkg/config"
	"logiq/pkg/jsonConvertor"
	"logiq/pkg/parquetwriter"
	"logiq/pkg/reporter"
	"logiq/pkg/scanner"
	"os"
	"sync"
	"time"
)

// executeProcess handles one complete log scan → parquet → analysis → report cycle
func executeProcess(cfg *config.Config, generateReport bool) {
	fmt.Printf("\n⏳ Starting log scan at %v\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("📂 Scanning paths: %v\n", cfg.LogPaths)

	logLines, err := scanner.ScanLogs(cfg)
	if err != nil {
		fmt.Println("❌ Error scanning logs:", err)
		return
	}

	if len(logLines) == 0 {
		fmt.Println("⚠️ No new logs found, skipping this cycle.")
		return
	}

	err = jsonConvertor.SaveAsJSONArray(logLines)
	if err != nil {
		fmt.Println("❌ Error saving logs as JSON:", err)
		return
	}
	fmt.Println("✅ Logs saved as JSON")

	err = parquetwriter.ConvertJSONToParquet(*cfg)
	if err != nil {
		fmt.Println("❌ Error converting JSON to Parquet:", err)
		return
	}

	anomalies := analyzer.AnalyzeLogs(cfg, logLines)
	fmt.Printf("🧠 Detected %d anomalies\n", len(anomalies))

	if generateReport {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("📄 Generating reports...")
			if err := reporter.GenerateReport(anomalies, cfg.Output); err != nil {
				fmt.Println("❌ Error generating report:", err)
			} else {
				fmt.Println("✅ Report generated successfully")
			}
		}()
		wg.Wait()
	}

	fmt.Printf("✅ Completed cycle at %v\n", time.Now().Format("15:04:05"))
}

func main() {
	configPath := flag.String("config", "configs/logiq.yaml", "Path to the config file")
	scan := flag.Bool("scan", false, "Run scan immediately")
	report := flag.Bool("report", false, "Generate report after scan")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Println("❌ Error loading config:", err)
		os.Exit(1)
	}

	// Extract interval from YAML (analysis.interval_seconds)
	intervalSeconds := cfg.Analysis.IntervalSeconds
	if intervalSeconds <= 0 {
		intervalSeconds = 60 // fallback default
	}

	if *scan {
		fmt.Println("🚀 Manual scan triggered")
		executeProcess(cfg, *report)
		return
	}

	fmt.Printf("🔁 Running every %d Seconds based on config\n", intervalSeconds)
	ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		executeProcess(cfg, true)
		fmt.Printf("⏳ Waiting %d seconds for next run...\n", intervalSeconds)
		<-ticker.C
	}
}
