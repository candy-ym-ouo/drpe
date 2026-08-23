package demo

import (
	"database/sql"
	"drpe/internal/model"
	"drpe/internal/service"
	"drpe/internal/store"
	"fmt"
	"time"
)

func Run(db *sql.DB) error {
	_, e := db.Exec(`CREATE TABLE IF NOT EXISTS orders(id INTEGER PRIMARY KEY, status TEXT, created_at TEXT, phone TEXT)`)
	if e != nil {
		return e
	}
	for i := 1; i <= 20; i++ {
		db.Exec(`INSERT OR IGNORE INTO orders VALUES(?,?,?,?)`, i, "closed", time.Now().AddDate(-2, 0, 0).Format(time.RFC3339), "13800138000")
	}
	ds := &model.DataSource{Name: "orders", Type: "sqlite", Connection: ":memory:", TableName: "orders", PrimaryKey: "id", TimeField: "created_at"}
	src := store.DataSources{DB: db}
	if existing, findErr := src.FindByName(ds.Name); findErr == nil {
		ds = existing
	} else if e = src.Create(ds); e != nil {
		return e
	}
	p := &model.Policy{Name: "订单数据保留", DataSourceID: ds.ID, Status: model.Draft}
	ps := service.PolicyService{Policies: store.Policies{DB: db}, Rules: store.Rules{DB: db}}
	if _, e = ps.Create(p); e != nil {
		return e
	}
	r := &model.Rule{PolicyID: p.ID, Name: "两年归档", Priority: 100, Enabled: true, Condition: []byte(`{"age":{"op":"older_than","value":365,"unit":"days"}}`), Action: "archive", Params: []byte(`{"destination":"./archive"}`)}
	if e = ps.Rules.Create(r); e != nil {
		return e
	}
	ex := &service.Executor{DB: db, Policies: ps.Policies, Sources: src, Rules: ps.Rules, Executions: store.Executions{DB: db}, Logs: store.Logs{DB: db}}
	x, e := ex.Run(p.ID, false)
	if e == nil {
		fmt.Printf("[DRPE DEMO] executed=%d archived=%d status=%s\n", x.Matched, x.Archived, x.Status)
	}
	return e
}
func SeedRows(db *sql.DB, n int) error {
	for i := 1; i <= n; i++ {
		if _, e := db.Exec(`INSERT OR IGNORE INTO orders VALUES(?,?,?,?)`, i, "closed", time.Now().AddDate(-2, 0, 0).Format(time.RFC3339), "13800138000"); e != nil {
			return e
		}
	}
	return nil
}
func Report(x *model.Execution) string {
	return fmt.Sprintf("status=%s scanned=%d matched=%d processed=%d failed=%d", x.Status, x.Scanned, x.Matched, x.Processed(), x.Failed)
}
