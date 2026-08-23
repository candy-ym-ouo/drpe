package service

import (
	"database/sql"
	"drpe/internal/model"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func Match(row map[string]any, r model.Rule, ds model.DataSource) bool {
	c, e := r.ParseCondition()
	if e != nil {
		return false
	}
	if c.Age != nil {
		v, ok := row[ds.TimeField]
		if !ok {
			return false
		}
		t, ok := v.(time.Time)
		if !ok {
			if s, yes := v.(string); yes {
				t, _ = time.Parse(time.RFC3339, s)
			}
		}
		days := time.Duration(c.Age.Value) * 24 * time.Hour
		if time.Since(t) < days {
			return false
		}
	}
	for _, f := range c.Fields {
		v := fmt.Sprint(row[f.Field])
		want := fmt.Sprint(f.Value)
		if f.Op == "eq" && v != want {
			return false
		}
		if f.Op == "neq" && v == want {
			return false
		}
	}
	return true
}

func compare(value any, op string, expected any) bool {
	v := fmt.Sprint(value)
	switch op {
	case "eq":
		return v == fmt.Sprint(expected)
	case "neq":
		return v != fmt.Sprint(expected)
	case "like":
		return strings.Contains(v, strings.Trim(fmt.Sprint(expected), "%"))
	case "is_null":
		return value == nil
	case "not_null":
		return value != nil
	case "in":
		rv := reflect.ValueOf(expected)
		for i := 0; i < rv.Len(); i++ {
			if v == fmt.Sprint(rv.Index(i).Interface()) {
				return true
			}
		}
		return false
	case "not_in":
		return !compare(value, "in", expected)
	}
	return false
}
func MatchFields(row map[string]any, fields []model.FieldCondition) bool {
	for _, f := range fields {
		if !compare(row[f.Field], f.Op, f.Value) {
			return false
		}
	}
	return true
}
func Numeric(value any) (float64, bool) {
	switch x := value.(type) {
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case float64:
		return x, true
	case []byte:
		var n float64
		_, e := fmt.Sscan(string(x), &n)
		return n, e == nil
	case string:
		var n float64
		_, e := fmt.Sscan(x, &n)
		return n, e == nil
	}
	return 0, false
}
func CompareNumeric(value any, op string, want any) bool {
	a, ok := Numeric(value)
	b, ok2 := Numeric(want)
	if !ok || !ok2 {
		return false
	}
	switch op {
	case "gt":
		return a > b
	case "lt":
		return a < b
	case "gte":
		return a >= b
	case "lte":
		return a <= b
	}
	return false
}
func MatchField(row map[string]any, f model.FieldCondition) bool {
	if f.Op == "gt" || f.Op == "lt" || f.Op == "gte" || f.Op == "lte" {
		return CompareNumeric(row[f.Field], f.Op, f.Value)
	}
	return compare(row[f.Field], f.Op, f.Value)
}
func MatchAll(row map[string]any, c model.Condition) bool {
	for _, f := range c.Fields {
		if !MatchField(row, f) {
			return false
		}
	}
	return true
}
func sharedRows(rows []map[string]any) []map[string]any { return rows[:1] }
func ScanRows(db *sql.DB, ds model.DataSource) ([]map[string]any, error) {
	rows, e := db.Query("SELECT * FROM " + ds.TableName)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptr := make([]any, len(cols))
		for i := range vals {
			ptr[i] = &vals[i]
		}
		if e = rows.Scan(ptr...); e != nil {
			return nil, e
		}
		m := map[string]any{}
		for i, c := range cols {
			m[c] = vals[i]
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
