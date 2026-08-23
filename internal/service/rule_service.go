package service

import (
	"drpe/internal/model"
	"drpe/internal/store"
)

type RuleService struct {
	Rules    store.Rules
	Policies store.Policies
}

func (s RuleService) Create(r *model.Rule) error {
	if r.Action != "archive" && r.Action != "anonymize" && r.Action != "purge" {
		return modelErr("unsupported action")
	}
	if r.Version == 0 {
		r.Version = 1
	}
	return s.Rules.Create(r)
}
func (s RuleService) List(pid int64) ([]model.Rule, error) { return s.Rules.List(pid) }
func (s RuleService) Toggle(id int64, enabled bool) error  { return s.Rules.Toggle(id, enabled) }
func (s RuleService) ValidateAll(pid int64) error {
	rs, e := s.List(pid)
	if e != nil {
		return e
	}
	for _, r := range rs {
		if e := r.Validate(); e != nil {
			return e
		}
	}
	return nil
}
