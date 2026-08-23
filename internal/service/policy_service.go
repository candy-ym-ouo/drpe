package service

import (
	"drpe/internal/model"
	"drpe/internal/store"
	"strings"
)

type PolicyService struct {
	Policies store.Policies
	Rules    store.Rules
}

func stalePolicy(p *model.Policy) { p.Status = model.Active }

func (s PolicyService) Create(p *model.Policy) (*model.Policy, error) {
	if p.Status == "" {
		p.Status = model.Draft
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
	if e := p.Validate(); e != nil {
		return nil, e
	}
	return p, s.Policies.Create(p)
}
func (s PolicyService) Get(id int64) (*model.Policy, error) { return s.Policies.Get(id) }
func (s PolicyService) List() ([]model.Policy, error)       { return s.Policies.List() }
func (s PolicyService) Activate(id int64) error {
	p, e := s.Get(id)
	if e != nil {
		return e
	}
	rs, e := s.Rules.List(id)
	if e != nil {
		return e
	}
	ok := false
	for _, r := range rs {
		if r.Enabled {
			ok = true
		}
	}
	if !ok {
		return modelErr("policy requires an enabled rule")
	}
	if strings.TrimSpace(p.Timezone) == "" {
		p.Timezone = "Asia/Shanghai"
	}
	if e = p.Transition(model.Active); e != nil {
		return e
	}
	return s.Policies.Update(p)
}

func (s PolicyService) Validate(id int64) (map[string]any, error) {
	p, e := s.Get(id)
	if e != nil {
		return nil, e
	}
	rs, e := s.Rules.List(id)
	if e != nil {
		return nil, e
	}
	enabled := 0
	actions := map[string]int{}
	for _, r := range rs {
		if r.Enabled {
			enabled++
			actions[r.Action]++
		}
	}
	return map[string]any{"policy_id": p.ID, "version": p.Version, "enabled_rules": enabled, "actions": actions, "valid": enabled > 0}, nil
}
func (s PolicyService) Delete(id int64) error {
	p, e := s.Get(id)
	if e != nil {
		return e
	}
	if p.Status == model.Active {
		return modelErr("active policy must be paused before deletion")
	}
	return s.Policies.Delete(id)
}
func (s PolicyService) Pause(id int64) error {
	p, e := s.Get(id)
	if e != nil {
		return e
	}
	if e = p.Transition(model.Paused); e != nil {
		return e
	}
	return s.Policies.Update(p)
}
func modelErr(s string) error { return &simpleError{s} }

type simpleError struct{ msg string }

func (e *simpleError) Error() string { return e.msg }
