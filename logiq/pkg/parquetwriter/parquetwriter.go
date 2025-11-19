package parquetwriter

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"logiq/pkg/config"
	"logiq/pkg/scanner"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/writer"
)

// ConvertJSONToParquet writes JSON logs into timestamped Parquet files per day.
func ConvertJSONToParquet(cfg config.Config) error {
	files, err := filepath.Glob(filepath.Join(cfg.OutputPaths.JSONDirPath, "*_logs.json"))
	if err != nil || len(files) == 0 {
		return fmt.Errorf("no JSON files found in %s", cfg.OutputPaths.JSONDirPath)
	}

	latestFile := files[len(files)-1]
	fmt.Println("📄 Converting:", latestFile)

	data, err := os.ReadFile(latestFile)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	var logs []scanner.LogLine
	if err := json.Unmarshal(data, &logs); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if len(logs) == 0 {
		log.Println("⚠️ No logs found to convert.")
		return nil
	}

	// Group logs by date
	grouped := make(map[string][]scanner.LogLine)
	for _, logLine := range logs {
		t, err := time.Parse("2006-01-02 15:04:05", logLine.TimeStamp)
		if err != nil {
			t = time.Now()
		}
		dateKey := t.Format("2006-01-02")
		grouped[dateKey] = append(grouped[dateKey], logLine)
	}

	for dateKey, group := range grouped {
		t, _ := time.Parse("2006-01-02", dateKey)
		outDir := filepath.Join(cfg.OutputPaths.ParquetDirPath,
			fmt.Sprintf("year=%d", t.Year()),
			fmt.Sprintf("month=%02d", t.Month()),
			fmt.Sprintf("day=%02d", t.Day()),
		)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Use current timestamp in filename
		timestamp := time.Now().Format("20060102_150405") // YYYYMMDD_HHMMSS
		parquetPath := filepath.Join(outDir, fmt.Sprintf("logs_%s.parquet", timestamp))

		fw, err := local.NewLocalFileWriter(parquetPath)
		if err != nil {
			return fmt.Errorf("failed to create parquet writer: %w", err)
		}

		pw, err := writer.NewParquetWriter(fw, new(scanner.LogLine), 4)
		if err != nil {
			return fmt.Errorf("failed to init parquet writer: %w", err)
		}
		pw.RowGroupSize = 128 * 1024 * 1024
		pw.CompressionType = parquet.CompressionCodec_SNAPPY

		for _, l := range group {
			if err := pw.Write(l); err != nil {
				log.Printf("❌ failed to write log: %v", err)
			}
		}

		if err := pw.WriteStop(); err != nil {
			return fmt.Errorf("failed to finalize parquet file: %w", err)
		}
		_ = fw.Close()

		log.Printf("✅ Created Parquet: %s (records: %d)", parquetPath, len(group))
	}

	return nil
}
