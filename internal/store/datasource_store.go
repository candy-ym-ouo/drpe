package store

import (
	"database/sql"
	"drpe/internal/model"
)

type DataSources struct{ DB *sql.DB }

func (s DataSources) Create(d *model.DataSource) error {
	if err := d.Validate(); err != nil {
		return err
	}
	r, e := s.DB.Exec(`INSERT INTO datasources(name,type,connection,table_name,primary_key,time_field,time_format,where_prefix) VALUES(?,?,?,?,?,?,?,?)`, d.Name, d.Type, d.Connection, d.TableName, d.PrimaryKey, d.TimeField, d.TimeFormat, d.WherePrefix)
	if e == nil {
		d.ID, _ = r.LastInsertId()
	}
	return e
}

func (s DataSources) Check(id int64) (map[string]any, error) {
	d, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	if err = d.Validate(); err != nil {
		return nil, err
	}
	if err = Ping(s.DB); err != nil {
		return nil, err
	}
	cols, err := TableColumns(s.DB, d.TableName)
	if err != nil {
		return nil, err
	}
	return map[string]any{"name": d.Name, "table": d.TableName, "columns": cols, "primary_key_exists": HasColumn(cols, d.PrimaryKey), "time_field_exists": HasColumn(cols, d.TimeField)}, nil
}
func (s DataSources) List() ([]model.DataSource, error) {
	rows, e := s.DB.Query(`SELECT id,name,type,connection,table_name,primary_key,time_field,time_format,where_prefix FROM datasources ORDER BY id DESC`)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	var out []model.DataSource
	for rows.Next() {
		var d model.DataSource
		if e = rows.Scan(&d.ID, &d.Name, &d.Type, &d.Connection, &d.TableName, &d.PrimaryKey, &d.TimeField, &d.TimeFormat, &d.WherePrefix); e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
func (s DataSources) Get(id int64) (*model.DataSource, error) {
	d := &model.DataSource{}
	e := s.DB.QueryRow(`SELECT id,name,type,connection,table_name,primary_key,time_field,time_format,where_prefix FROM datasources WHERE id=?`, id).Scan(&d.ID, &d.Name, &d.Type, &d.Connection, &d.TableName, &d.PrimaryKey, &d.TimeField, &d.TimeFormat, &d.WherePrefix)
	return d, e
}
func (s DataSources) Update(d *model.DataSource) error {
	if e := d.Validate(); e != nil {
		return e
	}
	_, e := s.DB.Exec(`UPDATE datasources SET type=?,connection=?,table_name=?,primary_key=?,time_field=?,time_format=?,where_prefix=? WHERE id=?`, d.Type, d.Connection, d.TableName, d.PrimaryKey, d.TimeField, d.TimeFormat, d.WherePrefix, d.ID)
	return e
}
func (s DataSources) Delete(id int64) error {
	var n int
	if e := s.DB.QueryRow(`SELECT COUNT(*) FROM policies WHERE data_source_id=? AND deleted_at IS NULL`, id).Scan(&n); e != nil {
		return e
	}
	if n > 0 {
		return modelErrStore("data source is bound to a policy")
	}
	_, e := s.DB.Exec(`DELETE FROM datasources WHERE id=?`, id)
	return e
}
func (s DataSources) FindByName(name string) (*model.DataSource, error) {
	d := &model.DataSource{}
	e := s.DB.QueryRow(`SELECT id,name,type,connection,table_name,primary_key,time_field,time_format,where_prefix FROM datasources WHERE name=?`, name).Scan(&d.ID, &d.Name, &d.Type, &d.Connection, &d.TableName, &d.PrimaryKey, &d.TimeField, &d.TimeFormat, &d.WherePrefix)
	return d, e
}
