package store

import (
	"database/sql"
	"drpe/internal/model"
	"time"
)

type Executions struct{ DB *sql.DB }

func (s Executions) Create(e *model.Execution) error {
	e.StartedAt = time.Now().UTC().Format(time.RFC3339)
	r, x := s.DB.Exec(`INSERT INTO executions(policy_id,trigger,status,dry_run,started_at) VALUES(?,?,?,?,?)`, e.PolicyID, e.Trigger, e.Status, e.DryRun, e.StartedAt)
	if x == nil {
		e.ID, _ = r.LastInsertId()
	}
	return x
}
func (s Executions) Finish(e *model.Execution) error {
	e.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	_, x := s.DB.Exec(`UPDATE executions SET status=?,scanned=?,matched=?,archived=?,anonymized=?,purged=?,failed=?,error_message=?,finished_at=? WHERE id=?`, e.Status, e.Scanned, e.Matched, e.Archived, e.Anonymized, e.Purged, e.Failed, e.Error, e.FinishedAt, e.ID)
	return x
}
func (s Executions) List() ([]model.Execution, error) {
	rows, e := s.DB.Query(`SELECT id,policy_id,trigger,status,dry_run,scanned,matched,archived,anonymized,purged,failed,error_message,started_at,finished_at FROM executions ORDER BY id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Execution
	for rows.Next() {
		var x model.Execution
		var dr int
		if e = rows.Scan(&x.ID, &x.PolicyID, &x.Trigger, &x.Status, &dr, &x.Scanned, &x.Matched, &x.Archived, &x.Anonymized, &x.Purged, &x.Failed, &x.Error, &x.StartedAt, &x.FinishedAt); e != nil {
			return nil, e
		}
		x.DryRun = dr != 0
		out = append(out, x)
	}
	return out, rows.Err()
}

func (s Executions) Get(id int64) (*model.Execution, error) {
	x := &model.Execution{}
	var dry int
	err := s.DB.QueryRow(`SELECT id,policy_id,trigger,status,dry_run,scanned,matched,archived,anonymized,purged,failed,error_message,started_at,finished_at FROM executions WHERE id=?`, id).Scan(&x.ID, &x.PolicyID, &x.Trigger, &x.Status, &dry, &x.Scanned, &x.Matched, &x.Archived, &x.Anonymized, &x.Purged, &x.Failed, &x.Error, &x.StartedAt, &x.FinishedAt)
	x.DryRun = dry != 0
	return x, err
}
func (s Executions) ByPolicy(pid int64) ([]model.Execution, error) {
	all, e := s.List()
	if e != nil {
		return nil, e
	}
	out := make([]model.Execution, 0)
	for _, x := range all {
		if x.PolicyID == pid {
			out = append(out, x)
		}
	}
	return out, nil
}
func (s Executions) MarkInterrupted() error {
	_, e := s.DB.Exec(`UPDATE executions SET status='failed',error_message='interrupted' WHERE status='running'`)
	return e
}
func (s Executions) Count() (int, error) {
	var n int
	e := s.DB.QueryRow(`SELECT COUNT(*) FROM executions`).Scan(&n)
	return n, e
}
