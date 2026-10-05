package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// sandbox points every location bopen reads at a temporary home with a
// stub browser, the ClearURLs list enabled and cached, and no session bus.
func sandbox(tb testing.TB) {
	tb.Helper()
	home := tb.TempDir()
	for k, v := range map[string]string{
		"HOME":                     home,
		"XDG_CONFIG_HOME":          filepath.Join(home, "config"),
		"XDG_CACHE_HOME":           filepath.Join(home, "cache"),
		"XDG_DATA_HOME":            filepath.Join(home, "data"),
		"XDG_DATA_DIRS":            filepath.Join(home, "share"),
		"XDG_RUNTIME_DIR":          filepath.Join(home, "run"),
		"DBUS_SESSION_BUS_ADDRESS": "",
		"APPDATA":                  filepath.Join(home, "appdata"),
		"LOCALAPPDATA":             filepath.Join(home, "localappdata"),
	} {
		tb.Setenv(k, v)
	}
	write := func(path, content string) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			tb.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			tb.Fatal(err)
		}
	}
	for _, name := range []string{"one", "two", "three"} {
		write(filepath.Join(home, "share", "applications", name+".desktop"),
			"[Desktop Entry]\nType=Application\nName="+name+"\nExec=/bin/true %u\nMimeType=x-scheme-handler/https;\n")
	}
	config, _ := os.UserConfigDir()
	write(filepath.Join(config, "bopen", "config.toml"), "[rules]\nclearurls = true\n")
	list, err := os.ReadFile(filepath.Join("..", "..", "internal", "clearurls", "testdata", "rules.json"))
	if err != nil {
		tb.Fatal(err)
	}
	cache, _ := os.UserCacheDir()
	write(filepath.Join(cache, "bopen", "clearurls.json"), string(list))
}

const startupURL = "https://www.google.com/url?q=https%3A%2F%2Fwww.amazon.de%2Fdp%2FB000%2Fref%3Dsr_1_3%3Fcrid%3DX%26utm_source%3Dy&sa=D"

// startupBudget is far above the typical cost of the work done before the
// window appears; it catches order-of-magnitude regressions, not noise.
const startupBudget = 250 * time.Millisecond

func TestPrepareIsFast(t *testing.T) {
	sandbox(t)
	prepare(startupURL) // warm the file system cache
	start := time.Now()
	s, m := prepare(startupURL)
	elapsed := time.Since(start)
	t.Logf("prepare took %v (%d browsers, %d suggestions)", elapsed, len(m.Browsers), len(m.Analysis.Suggestions))
	if len(s.Problems) != 0 || m.Analysis == nil || len(m.Analysis.Suggestions) == 0 {
		t.Fatalf("unexpected preparation: problems %v", s.Problems)
	}
	if elapsed > startupBudget {
		t.Fatalf("prepare took %v, budget %v", elapsed, startupBudget)
	}
}

func BenchmarkPrepare(b *testing.B) {
	sandbox(b)
	for b.Loop() {
		prepare(startupURL)
	}
}
