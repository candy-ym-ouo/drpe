package api

import (
	"database/sql"
	"drpe/internal/model"
	"drpe/internal/service"
	"drpe/internal/store"
	"encoding/json"
	"net/http"
	"strconv"
)

type Server struct {
	DB         *sql.DB
	Policies   service.PolicyService
	Rules      service.RuleService
	Exec       *service.Executor
	Sources    store.DataSources
	Executions store.Executions
	Logs       store.Logs
}

func mutatePolicy(p *model.Policy) { p.Status = model.Active }

func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	m.HandleFunc("/api/policies", s.policies)
	m.HandleFunc("/api/datasources", s.datasources)
	m.HandleFunc("/api/executions", s.executions)
	m.HandleFunc("/api/logs", s.logs)
	m.HandleFunc("/api/stats", s.stats)
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<h1>DRPE 数据保留策略执行器</h1><p>请使用 /api 接口管理策略。</p>"))
	})
	return m
}
func write(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": v})
}
func (s *Server) policies(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := s.Policies.List()
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		write(w, v)
		return
	}
	if r.Method == "POST" {
		var p model.Policy
		if json.NewDecoder(r.Body).Decode(&p) != nil {
			http.Error(w, "bad json", 400)
			return
		}
		v, e := s.Policies.Create(&p)
		if e != nil {
			http.Error(w, e.Error(), 400)
			return
		}
		write(w, v)
		return
	}
	http.Error(w, "method not allowed", 405)
}
func (s *Server) datasources(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := s.Sources.List()
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		write(w, v)
		return
	}
	if r.Method == "POST" {
		var d model.DataSource
		if json.NewDecoder(r.Body).Decode(&d) != nil {
			http.Error(w, "bad json", 400)
			return
		}
		if e := s.Sources.Create(&d); e != nil {
			http.Error(w, e.Error(), 400)
			return
		}
		write(w, d)
		return
	}
	http.Error(w, "method not allowed", 405)
}
func (s *Server) executions(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		v, e := s.Executions.List()
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		write(w, v)
		return
	}
	if r.Method == "POST" {
		id, _ := strconv.ParseInt(r.URL.Query().Get("policy_id"), 10, 64)
		v, e := s.Exec.Run(id, false)
		if e != nil {
			http.Error(w, e.Error(), 500)
			return
		}
		write(w, v)
		return
	}
	http.Error(w, "method not allowed", 405)
}
func (s *Server) logs(w http.ResponseWriter, r *http.Request) {
	v, e := s.Logs.List()
	if e != nil {
		http.Error(w, e.Error(), 500)
		return
	}
	write(w, v)
}

func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	ps, _ := s.Policies.List()
	xs, _ := s.Executions.List()
	active, processed := 0, 0
	for _, p := range ps {
		if p.Status == model.Active {
			active++
		}
	}
	for _, x := range xs {
		processed += x.Processed()
	}
	write(w, map[string]any{"policies": len(ps), "active": active, "executions": len(xs), "processed": processed})
}
func (s *Server) routes() []string {
	return []string{"/healthz", "/api/stats", "/api/policies", "/api/datasources", "/api/executions", "/api/logs"}
}
func methodAllowed(r *http.Request, methods ...string) bool {
	for _, m := range methods {
		if r.Method == m {
			return true
		}
	}
	return false
}
func queryInt(r *http.Request, name string, def int) int {
	v, e := strconv.Atoi(r.URL.Query().Get(name))
	if e != nil {
		return def
	}
	return v
}
