package store

import (
	"database/sql"
	"drpe/internal/model"
)

type Rules struct{ DB *sql.DB }

func (s Rules) Create(r *model.Rule) error {
	if err := r.Validate(); err != nil {
		return err
	}
	q, e := s.DB.Exec(`INSERT INTO rules(policy_id,name,priority,enabled,condition_json,action,action_params_json,version) VALUES(?,?,?,?,?,?,?,?)`, r.PolicyID, r.Name, r.Priority, r.Enabled, string(r.Condition), r.Action, string(r.Params), r.Version)
	if e == nil {
		r.ID, _ = q.LastInsertId()
	}
	return e
}

func (s Rules) CountEnabled(pid int64) (int, error) {
	var n int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM rules WHERE policy_id=? AND enabled=1`, pid).Scan(&n)
	return n, err
}
func (s Rules) Toggle(id int64, enabled bool) error {
	_, err := s.DB.Exec(`UPDATE rules SET enabled=? WHERE id=?`, enabled, id)
	return err
}
func (s Rules) Delete(id int64) error {
	_, e := s.DB.Exec(`DELETE FROM rules WHERE id=?`, id)
	return e
}
func (s Rules) Get(id int64) (*model.Rule, error) {
	r := &model.Rule{}
	var en int
	var c, p string
	e := s.DB.QueryRow(`SELECT id,policy_id,name,priority,enabled,condition_json,action,action_params_json,version FROM rules WHERE id=?`, id).Scan(&r.ID, &r.PolicyID, &r.Name, &r.Priority, &en, &c, &r.Action, &p, &r.Version)
	r.Enabled = en != 0
	r.Condition = []byte(c)
	r.Params = []byte(p)
	return r, e
}
func (s Rules) Reorder(pid int64, priorities map[int64]int) error {
	tx, e := s.DB.Begin()
	if e != nil {
		return e
	}
	for id, p := range priorities {
		if _, e = tx.Exec(`UPDATE rules SET priority=? WHERE id=? AND policy_id=?`, p, id, pid); e != nil {
			tx.Rollback()
			return e
		}
	}
	return tx.Commit()
}
func (s Rules) Versions(pid int64) []int {
	rs, _ := s.List(pid)
	out := make([]int, 0, len(rs))
	for _, r := range rs {
		out = append(out, r.Version)
	}
	return out
}
func (s Rules) List(pid int64) ([]model.Rule, error) {
	rows, e := s.DB.Query(`SELECT id,policy_id,name,priority,enabled,condition_json,action,action_params_json,version FROM rules WHERE policy_id=? ORDER BY priority DESC,id`, pid)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.Rule
	for rows.Next() {
		var r model.Rule
		var en int
		var c, p string
		if e = rows.Scan(&r.ID, &r.PolicyID, &r.Name, &r.Priority, &en, &c, &r.Action, &p, &r.Version); e != nil {
			return nil, e
		}
		r.Enabled = en != 0
		r.Condition = []byte(c)
		r.Params = []byte(p)
		out = append(out, r)
	}
	return out, rows.Err()
}
