# 🪵 **LogIQ — Intelligent Log Scanner, Analyzer & Dashboard**

### *Fast. Automated. DuckDB Powered.*

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" />
  <img src="https://img.shields.io/badge/Python-3.9+-3776AB?logo=python&logoColor=white" />
  <img src="https://img.shields.io/badge/DuckDB-0.10+-FFF000?logo=duckdb&logoColor=black" />
  <img src="https://img.shields.io/badge/Streamlit-1.x-FF4B4B?logo=streamlit&logoColor=white" />
  <img src="https://img.shields.io/badge/Parquet-Optimized-0E7AFE?logo=apache" />
  <img src="https://img.shields.io/badge/License-MIT-green" />
</p>

---

# 📌 **Overview**

**LogIQ** is a high-performance log Intelligence platform that:

✔️ Scans logs from files or directories
✔️ Extracts timestamps, file paths, and metadata
✔️ Converts logs → JSON → Parquet
✔️ Loads them into **DuckDB** for instant analytics
✔️ Detects anomalies using rule-based analysis
✔️ Generates JSON + HTML reports
✔️ Includes a full **Streamlit dashboard** with SQL query editor and visualizations

It is built in **Go + Python**, optimized for **sub-second querying** of millions of log rows.

---

# 📂 Project Architecture

```
                 ┌───────────────────────────────┐
                 │          Log Sources           │
                 │  *.log  / app logs / system    │
                 └───────────────────────────────┘
                                │
                                ▼
             ┌─────────────────────────────────────────┐
             │        Go Log Scanner (scanner/)         │
             │  ✓ Pattern filters                       │
             │  ✓ Timestamp extraction                  │
             └─────────────────────────────────────────┘
                                │
                                ▼
        ┌─────────────────────────────┐     ┌───────────────────────────┐
        │      JSON Convertor         │     │       Parquet Writer       │
        │ jsonConvertor/              │     │ parquetwriter/             │
        └─────────────────────────────┘     └───────────────────────────┘
                                │
                                ▼
                   ┌──────────────────────┐
                   │ DuckDB Query Engine  │
                   │    queryengine/      │
                   └──────────────────────┘
                                │
                                ▼
             ┌───────────────────────────────────────┐
             │      Streamlit Dashboard (ui/)        │
             │  ✓ SQL Query editor                   │
             │  ✓ Charts / Heatmaps                  │
             │  ✓ CSV Export                         │
             └───────────────────────────────────────┘
```

---

# 📁 Folder Structure

```
logiq/
│── cmd/
│     └── main.go
│
│── pkg/
│     ├── scanner/              # Reads & filters logs
│     ├── jsonConvertor/        # Saves logs as JSON
│     ├── parquetwriter/        # Converts JSON → Parquet
│     ├── analyzer/             # Anomaly detection
│     ├── reporter/             # HTML/JSON report generation
│     ├── config/               # YAML config loader
│     └── ...
│
│── queryengine/
│     └── engine.py             # DuckDB Parquet engine
│
│── ui/
│     └── main.py               # Streamlit dashboard
│
│── configs/
│     └── logiq.yaml            # Main configuration file
│
│── logs/                       # Incoming log files
│── jsonLogs/                   # Temporary JSON storage
│── mnt/data/logiq/parquet/     # Parquet output
│── reports/                    # Generated reports
│
└── README.md
```

---

# ⚙️ Installation

### **1. Clone the repo**

```bash
git clone https://github.com/yourname/logiq.git
cd logiq
```

---

# 📦 Dependencies

### **Go**

```bash
go mod tidy
```

### **Python (Streamlit UI + DuckDB)**

```bash
pip install -r requirements.txt
```

Recommended packages:

```
duckdb
streamlit
pandas
altair
```

---

# 🚀 Running LogIQ

## ✅ **1. Manual Scan**

```bash
go run cmd/main.go --scan=true --report=true
```

## 🔁 **2. Scheduled Scan (every X seconds)**

Configured in `logiq.yaml`

```bash
go run cmd/main.go
```

---

# 🛠 Configuration (`configs/logiq.yaml`)

```yaml
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
```

---

# 📊 Streamlit Dashboard (UI)

Launch:

```bash
streamlit run ui/main.py
```

---

# 🖼 Dashboard Features

### ✔ SQL Query Editor

### ✔ Last 100 logs / last 100 errors

### ✔ Bar chart — Severity

### ✔ Line chart — Logs over Time

### ✔ Heatmap — File vs Severity

### ✔ Auto severity classification

### ✔ CSV export

### ✔ Auto schema detection

### ✔ Highlighted logs table (color-coded)

---

# 🧠 Sample SQL Queries

Fetch logs from Feb 2025:

```sql
SELECT * FROM logs 
WHERE DATE(timestamp) BETWEEN '2025-02-01' AND '2025-02-28';
```

Count daily errors:

```sql
SELECT DATE(timestamp) AS day, COUNT(*) 
FROM logs 
WHERE severity = 'ERROR'
GROUP BY day;
```

Find slow requests:

```sql
SELECT * FROM logs 
WHERE content LIKE '%timeout%' OR content LIKE '%slow%';
```

---

# 📄 Reports

Reports generated in:

```
reports/logiq-report.json
reports/logiq-report.html
```

Contains:

* Error distribution
* Spike detection
* Timeline graphs
* Keyword-level grouping

---

# 🧪 Parquet Query Performance

DuckDB can query:

* **1 million rows → < 200ms**
* **10 million rows → < 1s**

Tested using:

```sql
SELECT COUNT(*) FROM logs;
```

---

# 🧩 Example Go Runner (Main Loop)

```go
ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)
for {
    executeProcess(cfg, true)
    <-ticker.C
}
```

---

# 🔮 Future Enhancements

* Machine Learning anomaly detection
* Distributed log collectors
* Slack / Email alerting
* Kubernetes operator integration
* Kafka ingestion
* Real-time dashboards

---

# 🤝 Contributing

PRs are welcome!
Please follow Go formatting and PEP-8 for Python.

---
