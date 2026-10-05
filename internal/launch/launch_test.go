package launch

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateAccepts(t *testing.T) {
	for in, want := range map[string]string{
		"https://example.com/a?b=1#c":     "https://example.com/a?b=1#c",
		"HTTP://Example.com/":             "http://Example.com/",
		"  https://example.com/x  ":       "https://example.com/x",
		`https://example.com/?q="quoted"`: "https://example.com/?q=%22quoted%22",
		"https://example.com/?q=a b":      "https://example.com/?q=a%20b",
		"https://example.com/a b":         "https://example.com/a%20b",
	} {
		got, err := Validate(in)
		if err != nil {
			t.Errorf("%q: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	for _, in := range []string{
		"--gpu-launcher=calc.exe",
		"-new-window",
		"https:///path",
		"file:///etc/passwd",
		"javascript:alert(1)",
		"mailto:someone@example.com",
		"example.com",
		"",
		"https://exa\x00mple.com/",
	} {
		if got, err := Validate(in); !errors.Is(err, ErrNotWebURL) {
			t.Errorf("%q: got %q, %v; want ErrNotWebURL", in, got, err)
		}
	}
}

func TestValidatedURLIsSingleSafeArgument(t *testing.T) {
	got, err := Validate("https://example.com/?a=$(id)&b=;ls&c=`x`&d=\"e\" f")
	if err != nil {
		t.Fatal(err)
	}
	if strings.ContainsAny(got, " \t\"") || strings.HasPrefix(got, "-") {
		t.Fatalf("unsafe result %q", got)
	}
}
