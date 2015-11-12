package config

import (
	"testing"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	if config == nil {
		t.Fatal("Config is not initialized")
	}
}

func TestLoadDefaultConfig(t *testing.T) {
	if err := LoadConfig("../../config/wigo.conf"); err != nil {
		t.Fatal(err)
	}
	if GetConfig() == nil {
		t.Fatal("Config is not initialized")
	}
}

func TestDumpConfig(t *testing.T) {
	config = NewConfig()
	Dump()
}
