package model

import "errors"

const (
	Draft  = "draft"
	Active = "active"
	Paused = "paused"
)

type Policy struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	DataSourceID  int64   `json:"data_source_id"`
	Status        string  `json:"status"`
	Version       int     `json:"version"`
	ScheduleCron  string  `json:"schedule_cron"`
	Timezone      string  `json:"timezone"`
	CatchUp       bool    `json:"catch_up"`
	MaxRecords    int     `json:"max_records_per_run"`
	BatchSize     int     `json:"batch_size"`
	DryRunDefault bool    `json:"dry_run_default"`
	Owner         string  `json:"owner"`
	DeletedAt     *string `json:"deleted_at,omitempty"`
}

func (p *Policy) Transition(to string) error {
	if p.Status == Draft && to == Active || p.Status == Active && to == Paused || p.Status == Paused && to == Active {
		p.Status = to
		return nil
	}
	return errors.New("invalid policy state transition")
}
func (p *Policy) Validate() error {
	if p.Name == "" || p.DataSourceID == 0 {
		return errors.New("name and data_source_id are required")
	}
	if p.MaxRecords < 0 || p.BatchSize < 0 {
		return errors.New("limits must be non-negative")
	}
	return nil
}
func (p Policy) IsActive() bool { return p.Status == Active }
func (p Policy) IsPaused() bool { return p.Status == Paused }
func (p Policy) IsDraft() bool  { return p.Status == Draft }
func (p *Policy) ConfigureDefaults() {
	if p.Status == "" {
		p.Status = Draft
	}
	if p.Version == 0 {
		p.Version = 1
	}
	if p.MaxRecords == 0 {
		p.MaxRecords = 5000
	}
	if p.BatchSize == 0 {
		p.BatchSize = 200
	}
	if p.Timezone == "" {
		p.Timezone = "Asia/Shanghai"
	}
}
func (p Policy) LimitReached(count int) bool { return p.MaxRecords > 0 && count >= p.MaxRecords }
func (p Policy) BatchLimit() int {
	if p.BatchSize <= 0 {
		return 200
	}
	return p.BatchSize
}
func (p Policy) ScheduleEnabled() bool { return p.Status == Active && p.ScheduleCron != "" }
func (p Policy) CanDelete() bool       { return p.Status != Active }
func (p Policy) StatusLabel() string {
	switch p.Status {
	case Draft:
		return "草稿"
	case Active:
		return "运行中"
	case Paused:
		return "已暂停"
	}
	return "未知"
}
func (p Policy) Summary() map[string]any {
	return map[string]any{"id": p.ID, "name": p.Name, "status": p.Status, "version": p.Version, "scheduled": p.ScheduleEnabled(), "limit": p.MaxRecords}
}
