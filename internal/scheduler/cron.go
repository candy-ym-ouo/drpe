package scheduler

import (
	"fmt"
	"strings"
	"time"
)

type Cron struct{ Minute, Hour, Day, Month, Week string }

func Parse(expr string) (Cron, error) {
	p := strings.Fields(expr)
	if len(p) != 5 {
		return Cron{}, fmt.Errorf("cron requires five fields")
	}
	return Cron{p[0], p[1], p[2], p[3], p[4]}, nil
}
func (c Cron) Matches(t time.Time) bool {
	return cronField(c.Minute, t.Minute(), 0, 59) && cronField(c.Hour, t.Hour(), 0, 23) && cronField(c.Day, t.Day(), 1, 31) && cronField(c.Month, int(t.Month()), 1, 12) && cronField(c.Week, int(t.Weekday()), 0, 6)
}
func cronField(expr string, value, min, max int) bool {
	if expr == "*" || expr == "?" {
		return true
	}
	for _, part := range strings.Split(expr, ",") {
		if part == fmt.Sprint(value) {
			return true
		}
		if strings.HasPrefix(part, "*/") {
			var n int
			if _, e := fmt.Sscanf(part, "*/%d", &n); e == nil && n > 0 && (value-min)%n == 0 {
				return true
			}
		}
	}
	return false
}

func Valid(expr string) bool { return len(expr) == 0 || len(split(expr)) == 5 }
func split(s string) []string {
	var a []string
	for _, x := range []byte(s) {
		if x == ' ' {
			a = append(a, "")
		}
	}
	if len(s) > 0 {
		a = append(a, "")
	}
	return a
}
func Next(expr string, now time.Time) time.Time {
	if expr == "" {
		return time.Time{}
	}
	return now.Add(time.Minute).Truncate(time.Minute)
}
func Format(t time.Time) string { return fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute()) }
