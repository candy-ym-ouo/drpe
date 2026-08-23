package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

func ActionName(action string) string {
	switch action {
	case "archive":
		return "归档"
	case "anonymize":
		return "匿名化"
	case "purge":
		return "清理"
	}
	return "未知"
}
func ValidateAction(action string, params map[string]any) error {
	if action != "archive" && action != "anonymize" && action != "purge" {
		return fmt.Errorf("unsupported action")
	}
	if action == "archive" && fmt.Sprint(params["destination"]) == "<nil>" {
		return fmt.Errorf("archive destination is required")
	}
	return nil
}
func Mask(value string, head, tail int) string {
	if len(value) <= head+tail {
		return strings.Repeat("*", len(value))
	}
	return value[:head] + strings.Repeat("*", len(value)-head-tail) + value[len(value)-tail:]
}

func ApplyAnonymize(v any, method string, p map[string]any) any {
	s := fmt.Sprint(v)
	switch method {
	case "redact":
		if x, ok := p["replacement"].(string); ok {
			return x
		}
		return "***"
	case "null":
		return nil
	case "hash":
		h := sha256.Sum256([]byte(s + fmt.Sprint(p["salt"])))
		return hex.EncodeToString(h[:])
	case "mask":
		head, tail := 2, 2
		if x, ok := p["keep_head"].(float64); ok {
			head = int(x)
		}
		if x, ok := p["keep_tail"].(float64); ok {
			tail = int(x)
		}
		if len(s) <= head+tail {
			return strings.Repeat("*", len(s))
		}
		return s[:head] + strings.Repeat("*", len(s)-head-tail) + s[len(s)-tail:]
	}
	return s
}
func ArchiveLine(path string, row map[string]any) error {
	f, e := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if e != nil {
		return e
	}
	defer f.Close()
	for k, v := range row {
		if _, e = f.WriteString(fmt.Sprintf("%s=%v ", k, v)); e != nil {
			return e
		}
	}
	_, e = f.WriteString("\n")
	return e
}
func EnsureArchiveDir(path string) error { return os.MkdirAll(path, 0755) }
func ArchiveRows(path string, rows []map[string]any) (int, error) {
	if e := EnsureArchiveDir(path); e != nil {
		return 0, e
	}
	n := 0
	for _, row := range rows {
		if e := ArchiveLine(path+"/archive.jsonl", row); e != nil {
			return n, e
		}
		n++
	}
	return n, nil
}
