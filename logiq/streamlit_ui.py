import streamlit as st
from queryEngine.engine import ParquetQueryEngine
import pandas as pd
import re
import altair as alt

# ---------- Page Setup ----------
st.set_page_config(page_title="🪵 LogIQ Dashboard", layout="wide")
st.title("🪵 LogIQ Dashboard")

engine = ParquetQueryEngine()

st.subheader("📊 Logs Table Schema")
try:
    schema_df = engine.conn.execute("PRAGMA table_info('logs');").fetchdf()
    st.dataframe(schema_df, use_container_width=True)
except Exception as e:
    st.warning(f"⚠️ Could not fetch schema: {e}")


# ---------- Sidebar Filters ----------
st.sidebar.header("Filters")

# Date range filters
start_date = st.sidebar.date_input("Start Date", pd.to_datetime("2025-01-01"))
end_date = st.sidebar.date_input("End Date", pd.to_datetime("2025-12-31"))

# Severity classification patterns
SEVERITY_PATTERNS = {
    "CRITICAL": re.compile(r"\bCRITICAL\b", re.IGNORECASE),
    "ERROR": re.compile(r"\bERROR\b", re.IGNORECASE),
    "WARN": re.compile(r"\bWARN(ING)?\b", re.IGNORECASE),
    "INFO": re.compile(r"\bINFO\b", re.IGNORECASE),
    "DEBUG": re.compile(r"\bDEBUG\b", re.IGNORECASE),
}

def classify_severity(msg: str) -> str:
    for sev, pattern in SEVERITY_PATTERNS.items():
        if pattern.search(msg):
            return sev
    return "DEBUG"

# ---------- Quick Filters ----------
st.sidebar.markdown("### Quick Filters")
if "quick_filter" not in st.session_state:
    st.session_state.quick_filter = None

if st.sidebar.button("Last 100 Logs"):
    st.session_state.quick_filter = "last_100_logs"
if st.sidebar.button("Last 100 Errors"):
    st.session_state.quick_filter = "last_100_errors"


# ---------- Query Logic ----------
if st.session_state.quick_filter == "last_100_logs":
    base_query = "SELECT * FROM logs ORDER BY timestamp DESC LIMIT 100"
elif st.session_state.quick_filter == "last_100_errors":
    base_query = "SELECT * FROM logs where content like '%ERROR%' ORDER BY timestamp DESC LIMIT 100"
else:
    base_query = f"""
        SELECT * FROM logs 
        WHERE DATE(timestamp) BETWEEN '{start_date}' AND '{end_date}'
        ORDER BY timestamp DESC 
        LIMIT 1000
    """

# ---------- Query Input ----------
st.subheader("🧠 SQL Query")
user_query = st.text_area("Enter SQL query:", value=base_query, height=150)

# ---------- Run Query ----------
if st.button("▶️ Run Query"):
    with st.spinner("Fetching data..."):
        df = engine.query(user_query)

    if df is not None and not df.empty:
        # Normalize timestamps
        if "timestamp" in df.columns:
            df["timestamp"] = pd.to_datetime(df["timestamp"], errors="coerce")

        # Infer severity if missing
        if "severity" not in df.columns and "content" in df.columns:
            df["severity"] = df["content"].astype(str).apply(classify_severity)

        # Filter only ERROR logs if needed
        if st.session_state.quick_filter == "last_100_errors":
            df = df[df["severity"] == "ERROR"].sort_values("timestamp", ascending=False).head(100)
        # ---------- Summary Metrics ----------
        st.markdown("## 📈 Summary Metrics")
        total_logs = len(df)
        error_count = len(df[df["severity"] == "ERROR"]) if "severity" in df.columns else 0
        warn_count = len(df[df["severity"] == "WARN"]) if "severity" in df.columns else 0
        latest_time = df["timestamp"].max() if "timestamp" in df.columns else "N/A"

        col1, col2, col3, col4 = st.columns(4)
        col1.metric("Total Logs", total_logs)
        col2.metric("Errors", error_count)
        col3.metric("Warnings", warn_count)
        col4.metric("Last Log Time", str(latest_time))

        # ---------- Visualization ----------
        st.markdown("## 📊 Log Insights")
        col1, col2 = st.columns(2)

        with col1:
            st.subheader("Severity Counts")
            st.bar_chart(df["severity"].value_counts())

        with col2:
            st.subheader("Logs Over Time")
            if "timestamp" in df.columns:
                df_time = df.set_index("timestamp").resample("D").size()
                st.line_chart(df_time)

        # Heatmap (file vs severity)
        if {"file_path", "severity"} <= set(df.columns):
            st.subheader("File vs Severity Heatmap")
            heatmap = (
                alt.Chart(df)
                .mark_rect()
                .encode(
                    x=alt.X("file_path:N", sort="-y", title="File Path"),
                    y=alt.Y("severity:N", title="Severity"),
                    color=alt.Color("count():Q", scale=alt.Scale(scheme="reds")),
                    tooltip=["file_path", "severity", "count()"]
                )
                .properties(width=600, height=300)
            )
            st.altair_chart(heatmap, use_container_width=True)

        # ---------- Table with Color-Coded Severity ----------
        def highlight_severity(val):
            colors = {
                "ERROR": "background-color: #FFCCCC",
                "WARN": "background-color: #FFF5CC",
                "INFO": "background-color: #E8F4FF",
                "DEBUG": "background-color: #F0F0F0",
                "CRITICAL": "background-color: #FF9999"
            }
            return colors.get(val, "")

        st.markdown("## 📋 Logs Table")
        st.dataframe(df.style.applymap(highlight_severity, subset=["severity"]), use_container_width=True)

        # ---------- Download Results ----------
        csv = df.to_csv(index=False).encode("utf-8")
        st.download_button(
            label="📥 Download CSV",
            data=csv,
            file_name="logiq_results.csv",
            mime="text/csv"
        )
    else:
        st.warning("No results found.")
