package model

import "encoding/json"

import "fmt"

type Rule struct {
	ID        int64           `json:"id"`
	PolicyID  int64           `json:"policy_id"`
	Name      string          `json:"name"`
	Priority  int             `json:"priority"`
	Enabled   bool            `json:"enabled"`
	Condition json.RawMessage `json:"condition_json"`
	Action    string          `json:"action"`
	Params    json.RawMessage `json:"action_params"`
	Version   int             `json:"version"`
}
type AgeCondition struct {
	Op    string `json:"op"`
	Value int    `json:"value"`
	Unit  string `json:"unit"`
}
type FieldCondition struct {
	Field string `json:"field"`
	Op    string `json:"op"`
	Value any    `json:"value"`
}
type Condition struct {
	Age    *AgeCondition    `json:"age"`
	Fields []FieldCondition `json:"fields"`
}

func (r Rule) ParseCondition() (Condition, error) {
	var c Condition
	err := json.Unmarshal(r.Condition, &c)
	return c, err
}
func (r Rule) ParseParams() (map[string]any, error) {
	var p map[string]any
	err := json.Unmarshal(r.Params, &p)
	return p, err
}

func (r Rule) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("rule name is required")
	}
	if r.Priority < 0 {
		return fmt.Errorf("priority must be non-negative")
	}
	if !r.Enabled {
		return nil
	}
	if r.Action != "archive" && r.Action != "anonymize" && r.Action != "purge" {
		return fmt.Errorf("unsupported action %q", r.Action)
	}
	if len(r.Condition) == 0 || string(r.Condition) == "null" {
		return fmt.Errorf("condition is required")
	}
	_, err := r.ParseCondition()
	return err
}

func (c Condition) HasAge() bool { return c.Age != nil && c.Age.Value > 0 }
func (c Condition) FieldNames() []string {
	out := make([]string, 0, len(c.Fields))
	for _, f := range c.Fields {
		out = append(out, f.Field)
	}
	return out
}
func (r Rule) ActionLabel() string {
	switch r.Action {
	case "archive":
		return "归档"
	case "anonymize":
		return "匿名化"
	case "purge":
		return "清理"
	}
	return "未知"
}
func (r Rule) IsAction(action string) bool { return r.Action == action && r.Enabled }
func (r Rule) Snapshot(version int) Rule   { r.Version = version; return r }
func SortRules(rs []Rule) []Rule {
	out := append([]Rule(nil), rs...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Priority > out[i].Priority {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
func UniqueActions(rs []Rule) map[string]int {
	m := map[string]int{}
	for _, r := range rs {
		if r.Enabled {
			m[r.Action]++
		}
	}
	return m
}
func (r Rule) ConditionSummary() string {
	c, e := r.ParseCondition()
	if e != nil {
		return "invalid"
	}
	parts := make([]string, 0)
	if c.Age != nil {
		parts = append(parts, fmt.Sprintf("older_than:%d %s", c.Age.Value, c.Age.Unit))
	}
	for _, f := range c.Fields {
		parts = append(parts, f.Field+" "+f.Op)
	}
	return fmt.Sprint(parts)
}
func (r Rule) ParamsSummary() string {
	p, e := r.ParseParams()
	if e != nil {
		return "invalid"
	}
	return fmt.Sprint(len(p)) + " params"
}
func (r Rule) PriorityLabel() string { return fmt.Sprintf("P%d", r.Priority) }
func (r Rule) EnabledLabel() string {
	if r.Enabled {
		return "启用"
	}
	return "停用"
}
