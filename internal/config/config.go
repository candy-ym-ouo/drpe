package config

import "os"
import "strconv"

type Config struct {
	Addr, DBPath string
	Workers      int
}

func Load() Config {
	c := Config{Addr: ":8080", DBPath: "drpe.db", Workers: 2}
	if v := os.Getenv("DRPE_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("DRPE_DB"); v != "" {
		c.DBPath = v
	}
	return c
}
func (c Config) Address() string { return c.Addr }
func (c Config) Validate() error {
	if c.Addr == "" {
		return strconv.ErrSyntax
	}
	if c.Workers < 1 {
		return strconv.ErrRange
	}
	return nil
}
func envInt(name string, fallback int) int {
	v := os.Getenv(name)
	if v == "" {
		return fallback
	}
	n, e := strconv.Atoi(v)
	if e != nil {
		return fallback
	}
	return n
}
func (c *Config) ApplyEnv() { c.Workers = envInt("DRPE_SCHED_WORKERS", c.Workers) }

func (c Config) Values() map[string]string {
	return map[string]string{"addr": c.Addr, "db": c.DBPath, "workers": strconv.Itoa(c.Workers)}
}
func (c Config) IsMemoryDB() bool                { return c.DBPath == ":memory:" || c.DBPath == "" }
func (c Config) WithAddress(addr string) Config  { c.Addr = addr; return c }
func (c Config) WithDatabase(path string) Config { c.DBPath = path; return c }
func (c Config) WithWorkers(workers int) Config  { c.Workers = workers; return c }
func (c Config) EffectiveWorkers() int {
	if c.Workers < 1 {
		return 1
	}
	return c.Workers
}
func (c Config) IsProduction() bool { return !c.IsMemoryDB() && c.Addr != ":0" }
func (c Config) String() string {
	return c.Addr + " " + c.DBPath + " workers=" + strconv.Itoa(c.EffectiveWorkers())
}
func ParseWorkers(value string, fallback int) int {
	n, e := strconv.Atoi(value)
	if e != nil || n < 1 {
		return fallback
	}
	return n
}
func ParseBool(value string, fallback bool) bool {
	if value == "1" || value == "true" || value == "yes" {
		return true
	}
	if value == "0" || value == "false" || value == "no" {
		return false
	}
	return fallback
}
func Environment() map[string]string {
	out := map[string]string{}
	for _, k := range []string{"DRPE_ADDR", "DRPE_DB", "DRPE_SCHED_WORKERS", "DRPE_LOG_LEVEL"} {
		if v := os.Getenv(k); v != "" {
			out[k] = v
		}
	}
	return out
}
func DefaultAddress() string  { return ":8080" }
func DefaultDatabase() string { return "drpe.db" }
func DefaultWorkers() int     { return 2 }
func NormalizeAddress(addr string) string {
	if addr == "" {
		return DefaultAddress()
	}
	return addr
}
func NormalizeDatabase(path string) string {
	if path == "" {
		return DefaultDatabase()
	}
	return path
}
func NormalizeWorkers(n int) int {
	if n < 1 {
		return DefaultWorkers()
	}
	return n
}
func NewConfig(addr, db string, workers int) Config {
	return Config{Addr: NormalizeAddress(addr), DBPath: NormalizeDatabase(db), Workers: NormalizeWorkers(workers)}
}
