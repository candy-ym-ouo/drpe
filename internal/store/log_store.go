package store

import (
	"database/sql"
	"drpe/internal/model"
	"time"
)

type Logs struct{ DB *sql.DB }

func (s Logs) Add(l *model.OperationLog) error {
	_, e := s.DB.Exec(`INSERT INTO operation_logs(execution_id,action,target_keys,summary_json,checksum,created_at) VALUES(?,?,?,?,?,?)`, l.ExecutionID, l.Action, l.TargetKeys, l.Summary, l.Checksum, time.Now().UTC())
	return e
}
func (s Logs) List() ([]model.OperationLog, error) {
	rows, e := s.DB.Query(`SELECT id,execution_id,action,target_keys,summary_json,checksum FROM operation_logs ORDER BY id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.OperationLog
	for rows.Next() {
		var l model.OperationLog
		if e = rows.Scan(&l.ID, &l.ExecutionID, &l.Action, &l.TargetKeys, &l.Summary, &l.Checksum); e != nil {
			return nil, e
		}
		out = append(out, l)
	}
	return out, rows.Err()
}
func (s Logs) ByExecution(id int64) ([]model.OperationLog, error) {
	all, e := s.List()
	if e != nil {
		return nil, e
	}
	out := make([]model.OperationLog, 0)
	for _, l := range all {
		if l.ExecutionID == id {
			out = append(out, l)
		}
	}
	return out, nil
}
func (s Logs) Verify(secret string) (int, int, error) {
	ls, e := s.List()
	if e != nil {
		return 0, 0, e
	}
	bad := 0
	for _, l := range ls {
		if !l.Verify(secret) {
			bad++
		}
	}
	return len(ls), bad, nil
}
func (s Logs) Count() (int, error) {
	var n int
	e := s.DB.QueryRow(`SELECT COUNT(*) FROM operation_logs`).Scan(&n)
	return n, e
}
func (s Logs) Recent(limit int) ([]model.OperationLog, error) {
	all, e := s.List()
	if e != nil {
		return nil, e
	}
	if limit < 0 {
		limit = 0
	}
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}
func (s Logs) ForAction(action string) ([]model.OperationLog, error) {
	all, e := s.List()
	if e != nil {
		return nil, e
	}
	out := make([]model.OperationLog, 0)
	for _, l := range all {
		if l.Action == action {
			out = append(out, l)
		}
	}
	return out, nil
}
func (s Logs) ChecksumFor(l model.OperationLog, secret string) string {
	return model.Checksum(l.ExecutionID, l.Action, l.TargetKeys, l.Summary, secret)
}
