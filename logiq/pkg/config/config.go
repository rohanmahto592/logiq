package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type AlertRule struct {
	Keyword  string `yaml:"keyword"`
	Severity string `yaml:"severity"`
}

type OutputConfig struct {
	Format     string `yaml:"format"`
	ReportFile string `yaml:"report_file"`
}

type AnalysisConfig struct {
	GroupBy         string  `yaml:"group_by"`
	ThresholdSpike  float64 `yaml:"threshold_spike"`
	IntervalSeconds int     `yaml:"interval_seconds"`
}
type OutputPathsConfig struct {
	ParquetDirPath string `yaml:"parquet_dir_path"`
	JSONDirPath    string `yaml:"json_dir_path"`
}

type Config struct {
	LogPaths          []string          `yaml:"log_paths"`
	IncludePatterns   []string          `yaml:"include_patterns"`
	ExcludePatterns   []string          `yaml:"exclude_patterns"`
	TimestampPatterns []string          `yaml:"timestamp_patterns"`
	AlertRules        []AlertRule       `yaml:"alert_rules"`
	Output            OutputConfig      `yaml:"output"`
	Analysis          AnalysisConfig    `yaml:"analysis"`
	OutputPaths       OutputPathsConfig `yaml:"output_paths"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &cfg, nil
}
