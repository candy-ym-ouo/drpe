package model

import (
	"crypto/sha256"
	"fmt"
)

type OperationLog struct {
	ID          int64  `json:"id"`
	ExecutionID int64  `json:"execution_id"`
	Action      string `json:"action,omitempty"`
	TargetKeys  string `json:"target_keys,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Checksum    string `json:"checksum,omitempty"`
}

func (l OperationLog) Verify(secret string) bool {
	return l.Checksum == Checksum(l.ExecutionID, l.Action, l.TargetKeys, l.Summary, secret)
}
func (l OperationLog) SafeSummary() string {
	if len(l.Summary) > 500 {
		return l.Summary[:500]
	}
	return l.Summary
}

func (l OperationLog) Fields() map[string]string {
	return map[string]string{
		"execution_id": fmt.Sprint(l.ExecutionID),
		"action":       l.Action,
		"target_keys":  l.TargetKeys,
		"summary":      l.SafeSummary(),
		"checksum":     l.Checksum,
	}
}

func (l *OperationLog) Normalize() {
	if l.Action == "" {
		l.Action = "unknown"
	}
	if l.TargetKeys == "" {
		l.TargetKeys = "-"
	}
	if l.Summary == "" {
		l.Summary = "{}"
	}
}

func (l OperationLog) HasTarget() bool   { return l.TargetKeys != "" && l.TargetKeys != "-" }
func (l OperationLog) IsArchive() bool   { return l.Action == "archive" }
func (l OperationLog) IsAnonymize() bool { return l.Action == "anonymize" }
func (l OperationLog) IsPurge() bool     { return l.Action == "purge" }

func Checksum(id int64, action, keys, summary, secret string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s|%s|%s", id, action, keys, summary, secret))))
}
