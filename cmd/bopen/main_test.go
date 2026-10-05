package main

import (
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
