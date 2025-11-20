📘 LogIQ — Intelligent Log Scanner, Analyzer & Parquet Pipeline

LogIQ is a lightweight, high-performance log processing engine built in Go + Python.
It automatically scans log directories, extracts events, converts logs into Parquet, detects anomalies, and generates JSON + HTML reports — all on a scheduled interval.

Designed for developers, SREs, and observability systems who want a simple, blazing-fast, local-first log intelligence tool.

🚀 Features
🔎 1. Smart Log Scanner (Go)

Scans multiple log files using patterns (*.log, nested folders)

Includes/excludes lines using regex-based filters

Extracts timestamps using multiple regex formats

Handles multi-format logs: JSON logs, flat logs, mixed logs

📦 2. JSON → Parquet Conversion (Go)

Logs saved as structured JSON arrays

Parquet conversion using columnar schema

Automatic folder partitioning:

parquet/year=YYYY/month=MM/day=DD/*.parquet

🧠 3. Intelligent Log Analyzer (Go)

Detects anomalies such as:

ERROR spikes

Critical events

Timeouts

Pattern-based anomaly rules (alert_rules)

📊 4. Reporting Engine (Go + Python)

Generates JSON report

Generates clean HTML report

Stores reports inside /reports/

⏱ 5. Scheduler

Runs automatically every interval_seconds from logiq.yaml.

🐍 6. Python Query Engine

Python module (queryengine/engine.py) allows:

DuckDB querying of parquet logs

Aggregations, filtering, dashboards

Used by Streamlit UI (if needed)

📁 Folder Structure
LOGIQ/
│
├── cmd/logiq/
│   └── main.go                   # Main runner + scheduler
│
├── configs/
│   └── logiq.yaml                # Central configuration file
│
├── logs/
│   ├── temp.log                  # Sample logs
│   └── writex.log
│
├── jsonLogs/
│   └── YYYY-MM-DD_logs.json      # Raw logs saved as JSON
│
├── mnt/data/logiq/parquet/
│   └── year=2025/month=11/day=20/
│       ├── logs_*.parquet
│
├── pkg/
│   ├── analyzer/                 # Pattern-based anomaly detection
│   │   └── analyzer.go
│   ├── config/                   # YAML loader + struct bindings
│   │   └── config.go
│   ├── jsonConvertor/            # Saves logs as JSON arrays
│   │   └── convertor.go
│   ├── parquetwriter/            # JSON → Parquet writer
│   │   └── parquetwriter.go
│   ├── reporter/                 # HTML + JSON report generator
│   │   └── reporter.go
│   └── scanner/                  # Log scanner
│       └── scanner.go
│
├── queryengine/                  # Python DuckDB engine
│   ├── __init__.py
│   └── engine.py
│
├── reports/
│   ├── logiq-report.json         # Output JSON report
│   └── logiq-report.json.html    # HTML report
│
└── ui/
    └── main.py                   # (Optional) Streamlit dashboard

⚙️ Configuration (logiq.yaml)
log_paths:
  - "./logs/*.log"

include_patterns:
  - "(?i)ERROR"
  - "(?i)WARN"
  - "(?i)CRITICAL"

exclude_patterns:
  - "(?i)DEBUG"

timestamp_patterns:
  - "\\d{4}-\\d{2}-\\d{2}[ T]\\d{2}:\\d{2}:\\d{2}"
  - "[A-Z][a-z]{2} [0-9]{1,2} [0-9:]{8}"
  - "\\d{2}/[A-Z][a-z]{3}/\\d{4}:\\d{2}:\\d{2}:\\d{2}"

alert_rules:
  - keyword: "(?i)ERROR"
    severity: "high"
  - keyword: "(?i)timeout"
    severity: "medium"

output:
  format: "json,html"
  report_file: "reports/logiq-report.json"

analysis:
  group_by: "keyword"
  threshold_spike: 50
  interval_seconds: 10

output_paths:
  parquet_dir_path: "mnt/data/logiq/parquet/"
  json_dir_path: "jsonLogs/"

▶️ Running LogIQ
Manual Run

Scan logs immediately:

go run cmd/logiq/main.go --scan --report

Scheduled Mode (default)

Runs every interval_seconds defined in config:

go run cmd/logiq/main.go

Custom config file
go run cmd/logiq/main.go --config myconfig.yaml

🧪 Example Query with DuckDB

Inside Python:

from queryengine.engine import ParquetQueryEngine

engine = ParquetQueryEngine()
df = engine.query("""
    SELECT *
    FROM logs
    WHERE level = 'ERROR'
    ORDER BY timestamp DESC
    LIMIT 100
""")

print(df)

📄 Sample Output Report
reports/logiq-report.json

Contains:

anomaly counts

event summaries

grouped statistics

reports/logiq-report.json.html

Clean HTML report viewable in browser.
