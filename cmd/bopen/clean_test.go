package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/blackfyre/bopen/internal/clean"
)

func builtinRules(t *testing.T) []clean.Rule {
	rules, err := clean.Builtin()
	if err != nil {
		t.Fatal(err)
	}
	return rules
}

func runCleanTest(t *testing.T, args []string, stdin string) (string, string, int) {
	var out, errOut bytes.Buffer
	status := runClean(args, builtinRules(t), strings.NewReader(stdin), &out, &errOut)
	return out.String(), errOut.String(), status
}

func TestCleanArgument(t *testing.T) {
	out, errOut, status := runCleanTest(t, []string{"https://example.com/a?utm_source=x&id=1"}, "")
	if out != "https://example.com/a?id=1\n" || errOut != "" || status != 0 {
		t.Fatalf("out %q err %q status %d", out, errOut, status)
	}
}

func TestCleanPipeKeepsOrderAndAffiliate(t *testing.T) {
	in := "https://example.com/?fbclid=x\n\n  https://www.amazon.de/dp/B000?tag=a-21&crid=1  \n"
	out, _, status := runCleanTest(t, nil, in)
	want := "https://example.com/\nhttps://www.amazon.de/dp/B000?tag=a-21&crid=1\n"
	if out != want || status != 0 {
		t.Fatalf("out %q status %d", out, status)
	}
}

func TestCleanInvalidLineEchoed(t *testing.T) {
	out, errOut, status := runCleanTest(t, nil, "https://example.com/?fbclid=x\nnot a link\n")
	if out != "https://example.com/\nnot a link\n" || !strings.Contains(errOut, `"not a link"`) || status != 1 {
		t.Fatalf("out %q err %q status %d", out, errOut, status)
	}
}

func TestCleanExplain(t *testing.T) {
	out, errOut, status := runCleanTest(t, []string{"--explain", "https://example.com/?fbclid=x"}, "")
	if out != "https://example.com/\n" || !strings.Contains(errOut, "fbclid=x") || !strings.Contains(errOut, "Facebook") || status != 0 {
		t.Fatalf("out %q err %q", out, errOut)
	}
}
