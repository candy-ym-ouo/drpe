package store

import "database/sql"

func Migrate(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY); CREATE TABLE IF NOT EXISTS datasources(id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT UNIQUE,type TEXT,connection TEXT,table_name TEXT,primary_key TEXT,time_field TEXT,time_format TEXT,where_prefix TEXT); CREATE TABLE IF NOT EXISTS policies(id INTEGER PRIMARY KEY AUTOINCREMENT,name TEXT,description TEXT,data_source_id INTEGER,status TEXT,version INTEGER,schedule_cron TEXT,timezone TEXT,catch_up INTEGER,max_records_per_run INTEGER,batch_size INTEGER,dry_run_default INTEGER,owner TEXT,deleted_at TEXT); CREATE TABLE IF NOT EXISTS rules(id INTEGER PRIMARY KEY AUTOINCREMENT,policy_id INTEGER,name TEXT,priority INTEGER,enabled INTEGER,condition_json TEXT,action TEXT,action_params_json TEXT,version INTEGER); CREATE TABLE IF NOT EXISTS executions(id INTEGER PRIMARY KEY AUTOINCREMENT,policy_id INTEGER,trigger TEXT,status TEXT,dry_run INTEGER,scanned INTEGER,matched INTEGER,archived INTEGER,anonymized INTEGER,purged INTEGER,failed INTEGER,error_message TEXT,started_at TEXT,finished_at TEXT); CREATE TABLE IF NOT EXISTS operation_logs(id INTEGER PRIMARY KEY AUTOINCREMENT,execution_id INTEGER,action TEXT,target_keys TEXT,summary_json TEXT,checksum TEXT,created_at TEXT); CREATE TABLE IF NOT EXISTS _drpe_processed(table_name TEXT,pk TEXT,execution_id INTEGER,action TEXT,created_at TEXT,PRIMARY KEY(table_name,pk));`)
	if err == nil {
		_, err = db.Exec(`INSERT OR IGNORE INTO schema_migrations(version) VALUES(1)`)
	}
	return err
}

func CurrentVersion(db *sql.DB) (int, error) {
	var v int
	err := db.QueryRow(`SELECT COALESCE(MAX(version),0) FROM schema_migrations`).Scan(&v)
	return v, err
}
func MigrationReady(db *sql.DB) bool { v, err := CurrentVersion(db); return err == nil && v >= 1 }
