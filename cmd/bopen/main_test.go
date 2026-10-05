package main

import "testing"

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
		{[]string{"https://example.com/"}, modeOpen},
		{[]string{"file:///etc/passwd"}, modeOpen},
	} {
		if got := parseArgs(tc.args); got != tc.want {
			t.Errorf("%q: got %v, want %v", tc.args, got, tc.want)
		}
	}
}
