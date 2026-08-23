package model

import "strings"

type DataSource struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Connection  string `json:"connection"`
	TableName   string `json:"table_name"`
	PrimaryKey  string `json:"primary_key"`
	TimeField   string `json:"time_field"`
	TimeFormat  string `json:"time_format"`
	WherePrefix string `json:"where_prefix"`
}

func (d DataSource) Validate() error {
	if strings.TrimSpace(d.Name) == "" {
		return errRequired("name")
	}
	if d.Type == "" {
		return errRequired("type")
	}
	if d.Type != "sqlite" {
		return errValue("type", "only sqlite is supported")
	}
	if strings.TrimSpace(d.TableName) == "" {
		return errRequired("table_name")
	}
	if strings.TrimSpace(d.PrimaryKey) == "" {
		return errRequired("primary_key")
	}
	if strings.TrimSpace(d.TimeField) == "" {
		return errRequired("time_field")
	}
	for _, v := range []string{d.TableName, d.PrimaryKey, d.TimeField} {
		if !safeIdentifier(v) {
			return errValue("identifier", v)
		}
	}
	return nil
}
func (d DataSource) IsSQLite() bool { return d.Type == "sqlite" }
func (d DataSource) DisplayName() string {
	if d.Name != "" {
		return d.Name
	}
	return d.TableName
}
func (d DataSource) QualifiedTable() string { return d.TableName }
func (d DataSource) Fields() []string       { return []string{d.PrimaryKey, d.TimeField} }
func (d DataSource) ConnectionLabel() string {
	if d.Connection == "" {
		return ":memory:"
	}
	return d.Connection
}

func safeIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if !(r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || i > 0 && r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

type validationError struct{ field, message string }

func (e validationError) Error() string    { return e.field + ": " + e.message }
func errRequired(field string) error       { return validationError{field, "is required"} }
func errValue(field, message string) error { return validationError{field, message} }
