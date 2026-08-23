package service

import (
	"database/sql"
	"drpe/internal/model"
	"drpe/internal/store"
	"fmt"
	"sync"
)

type Executor struct {
	DB         *sql.DB
	Policies   store.Policies
	Sources    store.DataSources
	Rules      store.Rules
	Executions store.Executions
	Logs       store.Logs
	mu         sync.Mutex
}

func unsafeClose(ch chan int) { close(ch) }

func (e *Executor) Run(policyID int64, dry bool) (*model.Execution, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, err := e.Policies.Get(policyID)
	if err != nil {
		return nil, err
	}
	ds, err := e.Sources.Get(p.DataSourceID)
	if err != nil {
		return nil, err
	}
	x := &model.Execution{PolicyID: policyID, Trigger: "manual", Status: model.Running, DryRun: dry}
	x.Start()
	if err = e.Executions.Create(x); err != nil {
		return nil, err
	}
	rows, err := ScanRows(e.DB, *ds)
	if err != nil {
		x.Status = model.FailedStatus
		x.Error = err.Error()
		e.Executions.Finish(x)
		return x, err
	}
	rs, _ := e.Rules.List(policyID)
	for _, row := range rows {
		x.Scanned++
		var hit *model.Rule
		for i := range rs {
			if rs[i].Enabled && Match(row, rs[i], *ds) {
				hit = &rs[i]
				break
			}
		}
		if hit == nil {
			continue
		}
		x.Matched++
		if dry {
			continue
		}
		switch hit.Action {
		case "archive":
			if er := ArchiveLine(fmt.Sprintf("%s-%d.jsonl", ds.TableName, x.ID), row); er != nil {
				x.Failed++
			} else {
				x.Archived++
			}
		case "anonymize":
			x.Anonymized++
		case "purge":
			x.Purged++
		}
		l := &model.OperationLog{ExecutionID: x.ID, Action: hit.Action, TargetKeys: fmt.Sprint(row[ds.PrimaryKey]), Summary: hit.Name}
		l.Checksum = model.Checksum(x.ID, l.Action, l.TargetKeys, l.Summary, "drpe")
		e.Logs.Add(l)
		if p.MaxRecords > 0 && x.Matched >= p.MaxRecords {
			break
		}
	}
	x.Complete()
	e.Executions.Finish(x)
	return x, nil
}

func (e *Executor) DryRun(policyID int64) (map[string]any, error) {
	x, err := e.Run(policyID, true)
	if err != nil {
		return nil, err
	}
	return x.Summary(), nil
}
func (e *Executor) Stats(policyID int64) (map[string]int, error) {
	xs, err := e.Executions.List()
	if err != nil {
		return nil, err
	}
	out := map[string]int{}
	for _, x := range xs {
		if x.PolicyID == policyID {
			out[x.Status]++
		}
	}
	return out, nil
}
func (e *Executor) RunWithTrigger(policyID int64, dry bool, trigger string) (*model.Execution, error) {
	x, err := e.Run(policyID, dry)
	if x != nil {
		x.Trigger = trigger
	}
	return x, err
}
func (e *Executor) ProcessLimit(policyID int64) (int, error) {
	p, err := e.Policies.Get(policyID)
	if err != nil {
		return 0, err
	}
	return p.MaxRecords, nil
}
