package store

import (
	"database/sql"
	"drpe/internal/model"
)

type Policies struct{ DB *sql.DB }

func (s Policies) Create(p *model.Policy) error {
	r, e := s.DB.Exec(`INSERT INTO policies(name,description,data_source_id,status,version,schedule_cron,timezone,catch_up,max_records_per_run,batch_size,dry_run_default,owner) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, p.Name, p.Description, p.DataSourceID, p.Status, p.Version, p.ScheduleCron, p.Timezone, p.CatchUp, p.MaxRecords, p.BatchSize, p.DryRunDefault, p.Owner)
	if e == nil {
		p.ID, _ = r.LastInsertId()
	}
	return e
}
func (s Policies) Get(id int64) (*model.Policy, error) {
	p := &model.Policy{}
	var c, dr int
	e := s.DB.QueryRow(`SELECT id,name,description,data_source_id,status,version,schedule_cron,timezone,catch_up,max_records_per_run,batch_size,dry_run_default,owner,deleted_at FROM policies WHERE id=? AND deleted_at IS NULL`, id).Scan(&p.ID, &p.Name, &p.Description, &p.DataSourceID, &p.Status, &p.Version, &p.ScheduleCron, &p.Timezone, &c, &p.MaxRecords, &p.BatchSize, &dr, &p.Owner, &p.DeletedAt)
	p.CatchUp = c != 0
	p.DryRunDefault = dr != 0
	if e == sql.ErrNoRows {
		return nil, nil
	}
	return p, e
}
func (s Policies) List() ([]model.Policy, error) {
	rows, e := s.DB.Query(`SELECT id,name,description,data_source_id,status,version,schedule_cron,timezone,catch_up,max_records_per_run,batch_size,dry_run_default,owner,deleted_at FROM policies WHERE deleted_at IS NULL ORDER BY id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Policy
	for rows.Next() {
		var p model.Policy
		var c, d int
		if e = rows.Scan(&p.ID, &p.Name, &p.Description, &p.DataSourceID, &p.Status, &p.Version, &p.ScheduleCron, &p.Timezone, &c, &p.MaxRecords, &p.BatchSize, &d, &p.Owner, &p.DeletedAt); e != nil {
			return nil, e
		}
		p.CatchUp = c != 0
		p.DryRunDefault = d != 0
		out = append(out, p)
	}
	return out, rows.Err()
}
func (s Policies) Update(p *model.Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	_, e := s.DB.Exec(`UPDATE policies SET name=?,description=?,status=?,version=?,schedule_cron=?,timezone=?,catch_up=?,max_records_per_run=?,batch_size=?,dry_run_default=?,owner=? WHERE id=?`, p.Name, p.Description, p.Status, p.Version, p.ScheduleCron, p.Timezone, p.CatchUp, p.MaxRecords, p.BatchSize, p.DryRunDefault, p.Owner, p.ID)
	return e
}

func (s Policies) BumpVersion(id int64, expected int) error {
	r, err := s.DB.Exec(`UPDATE policies SET version=version+1 WHERE id=? AND version=? AND deleted_at IS NULL`, id, expected)
	if err != nil {
		return err
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return modelErrStore("version conflict")
	}
	return nil
}

type storeError string

func (e storeError) Error() string { return string(e) }
func modelErrStore(s string) error { return storeError(s) }
func (s Policies) Delete(id int64) error {
	_, e := s.DB.Exec(`UPDATE policies SET deleted_at=datetime('now') WHERE id=?`, id)
	return e
}
func (s Policies) Active() ([]model.Policy, error) {
	all, e := s.List()
	if e != nil {
		return nil, e
	}
	out := make([]model.Policy, 0)
	for _, p := range all {
		if p.IsActive() {
			out = append(out, p)
		}
	}
	return out, nil
}
func (s Policies) CountByStatus() (map[string]int, error) {
	rows, e := s.DB.Query(`SELECT status,COUNT(*) FROM policies WHERE deleted_at IS NULL GROUP BY status`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var status string
		var n int
		if e = rows.Scan(&status, &n); e != nil {
			return nil, e
		}
		out[status] = n
	}
	return out, rows.Err()
}
func (s Policies) Search(term string) ([]model.Policy, error) {
	rows, e := s.DB.Query(`SELECT id,name,description,data_source_id,status,version,schedule_cron,timezone,catch_up,max_records_per_run,batch_size,dry_run_default,owner,deleted_at FROM policies WHERE deleted_at IS NULL AND (name LIKE ? OR description LIKE ?) ORDER BY id DESC`, "%"+term+"%", "%"+term+"%")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Policy
	for rows.Next() {
		var p model.Policy
		var c, d int
		if e = rows.Scan(&p.ID, &p.Name, &p.Description, &p.DataSourceID, &p.Status, &p.Version, &p.ScheduleCron, &p.Timezone, &c, &p.MaxRecords, &p.BatchSize, &d, &p.Owner, &p.DeletedAt); e != nil {
			return nil, e
		}
		p.CatchUp = c != 0
		p.DryRunDefault = d != 0
		out = append(out, p)
	}
	return out, rows.Err()
}
