package main

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want mode
	}{
		{nil, modeUsage},
		{[]string{""}, modeUsage},
		{[]string{"--help"}, modeUsage},
		{[]string{"a", "b"}, modeUsage},
		{[]string{"register"}, modeRegister},
		{[]string{"unregister"}, modeUnregister},
		{[]string{"settings"}, modeSettings},
		{[]string{"https://example.com/"}, modeOpen},
		{[]string{"file:///etc/passwd"}, modeOpen},
	} {
		if got := parseArgs(tc.args); got != tc.want {
			t.Errorf("%q: got %v, want %v", tc.args, got, tc.want)
		}
	}
}

func TestUsageShowsVersion(t *testing.T) {
	old := version
	defer func() { version = old }()
	version = "0.1.0"
	if got := usageText(); !strings.HasPrefix(got, "bopen 0.1.0\n") {
		t.Fatalf("usage = %q", got)
	}
}

func TestResolveVersion(t *testing.T) {
	info := func(v string) func() (*debug.BuildInfo, bool) {
		return func() (*debug.BuildInfo, bool) {
			return &debug.BuildInfo{Main: debug.Module{Path: "github.com/blackfyre/bopen", Version: v}}, true
		}
	}
	none := func() (*debug.BuildInfo, bool) { return nil, false }
	for _, tc := range []struct {
		linked string
		info   func() (*debug.BuildInfo, bool)
		want   string
	}{
		{"0.3.0", info("v9.9.9"), "0.3.0"},
		{"dev", info("v0.3.0"), "0.3.0"},
		{"dev", info("v0.4.0-0.20261005120000-abcdef123456"), "0.4.0-0.20261005120000-abcdef123456"},
		{"dev", info("(devel)"), "dev"},
		{"dev", info(""), "dev"},
		{"dev", none, "dev"},
	} {
		if got := resolveVersion(tc.linked, tc.info); got != tc.want {
			t.Errorf("linked %q: got %q, want %q", tc.linked, got, tc.want)
		}
	}
}
