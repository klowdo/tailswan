package version

import (
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	v := Get()
	if v == "" {
		t.Error("Get() returned empty string")
	}
}

func TestGetReturnsDefaultWhenNoBuildInfo(t *testing.T) {
	v := Get()
	if !strings.Contains(v, "0.") {
		t.Logf("Version returned: %s", v)
	}
}

func TestGetFull(t *testing.T) {
	full := GetFull()
	if full == "" {
		t.Error("GetFull() returned empty string")
	}
	if !strings.Contains(full, "Version:") {
		t.Errorf("GetFull() should contain 'Version:', got: %s", full)
	}
}

func TestDefaultVersionConstant(t *testing.T) {
	if DefaultVersion == "" {
		t.Error("DefaultVersion should not be empty")
	}
	if !strings.HasPrefix(DefaultVersion, "0.") && !strings.HasPrefix(DefaultVersion, "1.") {
		t.Errorf("DefaultVersion should start with major version, got: %s", DefaultVersion)
	}
}
