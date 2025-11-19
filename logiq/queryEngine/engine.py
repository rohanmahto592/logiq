import duckdb
import os

class ParquetQueryEngine:
    def __init__(self, base_dir: str = "mnt/data/logiq/parquet/"):
        self.base_dir = base_dir
        self.conn = duckdb.connect(database=':memory:')
        self.conn.execute("PRAGMA enable_object_cache;")
        self._register_parquet_dir()

    def _register_parquet_dir(self):
        parquet_path = os.path.join(self.base_dir, "**/*.parquet")
        self.conn.execute(f"""
            CREATE OR REPLACE VIEW logs AS
            SELECT * FROM read_parquet('{parquet_path}', hive_partitioning=true, filename=true);
        """)
        print(f"✅ Registered parquet directory as 'logs' view: {self.base_dir}")

    def query(self, sql: str):
        try:
            return self.conn.execute(sql).fetchdf()
        except Exception as e:
            print(f"❌ Query failed: {e}")
            return None

    def close(self):
        self.conn.close()
        print("🛑 Connection closed.")
