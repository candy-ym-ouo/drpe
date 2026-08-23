package model

import "time"

type Execution struct {
	ID         int64  `json:"id"`
	PolicyID   int64  `json:"policy_id"`
	Trigger    string `json:"trigger"`
	Status     string `json:"status"`
	DryRun     bool   `json:"dry_run"`
	Scanned    int    `json:"scanned"`
	Matched    int    `json:"matched"`
	Archived   int    `json:"archived"`
	Anonymized int    `json:"anonymized"`
	Purged     int    `json:"purged"`
	Failed     int    `json:"failed"`
	Error      string `json:"error_message,omitempty"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at,omitempty"`
}

const (
	Pending      = "pending"
	Running      = "running"
	Succeeded    = "succeeded"
	Partial      = "partial"
	FailedStatus = "failed"
)

func (e *Execution) Start() { e.Status = Running; e.StartedAt = time.Now().UTC().Format(time.RFC3339) }
func (e *Execution) Complete() {
	e.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	if e.Failed > 0 {
		e.Status = Partial
	} else {
		e.Status = Succeeded
	}
}
func (e Execution) IsTerminal() bool {
	return e.Status == Succeeded || e.Status == Partial || e.Status == FailedStatus
}
func (e Execution) Processed() int { return e.Archived + e.Anonymized + e.Purged }
func (e Execution) Summary() map[string]any {
	return map[string]any{"scanned": e.Scanned, "matched": e.Matched, "processed": e.Processed(), "archived": e.Archived, "anonymized": e.Anonymized, "purged": e.Purged, "failed": e.Failed, "status": e.Status}
}
func (e Execution) TriggerLabel() string {
	if e.Trigger == "schedule" {
		return "调度"
	}
	return "手动"
}
func (e Execution) DurationHint() string {
	if e.StartedAt == "" || e.FinishedAt == "" {
		return "运行中"
	}
	return e.StartedAt + " ~ " + e.FinishedAt
}
func (e *Execution) AddFailure() {
	e.Failed++
	if e.Status == Succeeded {
		e.Status = Partial
	}
}
func (e *Execution) AddAction(action string) {
	switch action {
	case "archive":
		e.Archived++
	case "anonymize":
		e.Anonymized++
	case "purge":
		e.Purged++
	}
}
func (e Execution) ActionCounts() map[string]int {
	return map[string]int{"archive": e.Archived, "anonymize": e.Anonymized, "purge": e.Purged}
}

func (e Execution) FailedRatio() float64 {
	if e.Matched == 0 {
		return 0
	}
	return float64(e.Failed) / float64(e.Matched)
}

func (e Execution) Successful() bool    { return e.Status == Succeeded && e.Failed == 0 }
func (e Execution) PartialResult() bool { return e.Status == Partial || e.Failed > 0 }
func (e Execution) HasError() bool      { return e.Error != "" || e.Failed > 0 }
func (e *Execution) SetError(err error) {
	if err != nil {
		e.Error = err.Error()
		e.Status = FailedStatus
	}
}
func (e Execution) TriggerIsManual() bool   { return e.Trigger == "manual" || e.Trigger == "" }
func (e Execution) TriggerIsSchedule() bool { return e.Trigger == "schedule" }
func (e Execution) DryRunLabel() string {
	if e.DryRun {
		return "预演"
	}
	return "正式"
}
