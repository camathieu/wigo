package utils

import (
	"testing"
)

func TestStatusCodeToString(t *testing.T) {
	var level string

	for code := -999 ; code < 100 ; code++ {
		level = StatusCodeToString(code)
		if level != "ERROR" {
			t.Fatalf("Invalid status string %s for status %d, expected %s", level, code, "ERROR")
		}
	}
	level = StatusCodeToString(100)
	if level != "OK" {
		t.Fatalf("Invalid status string %s for status %d, expected %s", level, 100, "OK")
	}
	for code := 101 ; code < 199 ; code++ {
		level = StatusCodeToString(code)
		if level != "INFO" {
			t.Fatalf("Invalid status string %s for status %d, expected %s", level, code, "ERROR")
		}
	}
	for code := 200 ; code < 299 ; code++ {
		level = StatusCodeToString(code)
		if level != "WARN" {
			t.Fatalf("Invalid status string %s for status %d, expected %s", level, code, "WARN")
		}
	}
	for code := 300 ; code < 499 ; code++ {
		level = StatusCodeToString(code)
		if level != "CRIT" {
			t.Fatalf("Invalid status string %s for status %d, expected %s", level, code, "CRITICAL")
		}
	}
	for code := 500 ; code < 999 ; code++ {
		level = StatusCodeToString(code)
		if level != "ERROR" {
			t.Fatalf("Invalid status string %s for status %d, expected %s", level, code, "ERROR")
		}
	}
}