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

func TestValidateKeepsLinkByteForByte(t *testing.T) {
	for _, in := range []string{
		"https://example.com/café/ü?q=ö#",
		"https://example.com/a%2Fb/c?x=%E2%82%AC&y=a+b",
		"https://example.com/it's(here)!*;p=1?#frag%41",
		"https://example.com/a%zz/100%",
		"http://user@Example.COM:8080/",
	} {
		got, err := Validate(in)
		if err != nil {
			t.Errorf("%q: %v", in, err)
			continue
		}
		if got != in {
			t.Errorf("%q changed to %q", in, got)
		}
	}
}

func TestValidateMalformedEscapeWithCleaning(t *testing.T) {
	got, err := Validate("https://example.com/a%zz?utm_source=x")
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com/a%zz?utm_source=x" {
		t.Fatalf("got %q", got)
	}
}
