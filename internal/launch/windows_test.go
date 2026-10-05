package launch

import "testing"

const u = "https://example.com/?a=1&b=2"

func TestWindowsCommandLine(t *testing.T) {
	for _, tc := range []struct{ name, template, exe, line string }{
		{"chrome", `"C:\Program Files\Google\Chrome\Application\chrome.exe" --single-argument %1`,
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`"C:\Program Files\Google\Chrome\Application\chrome.exe" --single-argument ` + u},
		{"edge", `"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe" --single-argument %1`,
			`C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`,
			`"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe" --single-argument ` + u},
		{"firefox", `"C:\Program Files\Mozilla Firefox\firefox.exe" -osint -url "%1"`,
			`C:\Program Files\Mozilla Firefox\firefox.exe`,
			`"C:\Program Files\Mozilla Firefox\firefox.exe" -osint -url "` + u + `"`},
		{"brave", `"C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe" --single-argument %1`,
			`C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe`,
			`"C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe" --single-argument ` + u},
		{"opera", `"C:\Users\u\AppData\Local\Programs\Opera\launcher.exe" -noautoupdate -- "%1"`,
			`C:\Users\u\AppData\Local\Programs\Opera\launcher.exe`,
			`"C:\Users\u\AppData\Local\Programs\Opera\launcher.exe" -noautoupdate -- "` + u + `"`},
		{"no placeholder", `"C:\Browser\browser.exe"`, `C:\Browser\browser.exe`, `"C:\Browser\browser.exe" ` + u},
		{"unquoted exe", `C:\Browser\browser.exe %1 %*`, `C:\Browser\browser.exe`, `C:\Browser\browser.exe ` + u + ` `},
		{"long-name code", `"C:\B\b.exe" "%L"`, `C:\B\b.exe`, `"C:\B\b.exe" "` + u + `"`},
	} {
		exe, line, err := WindowsCommandLine(tc.template, u)
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if exe != tc.exe || line != tc.line {
			t.Errorf("%s:\n got %q %q\nwant %q %q", tc.name, exe, line, tc.exe, tc.line)
		}
	}
}

func TestWindowsEmbeddedQuoteCannotBreakOut(t *testing.T) {
	valid, err := Validate(`https://example.com/?q=a" --gpu-launcher="calc`)
	if err != nil {
		t.Fatal(err)
	}
	_, line, err := WindowsCommandLine(`"C:\B\b.exe" "%1"`, valid)
	if err != nil {
		t.Fatal(err)
	}
	want := `"C:\B\b.exe" "https://example.com/?q=a%22%20--gpu-launcher=%22calc"`
	if line != want {
		t.Fatalf("got %q, want %q", line, want)
	}
}

func TestWindowsTrailingBackslashInsideQuotes(t *testing.T) {
	_, line, err := WindowsCommandLine(`"C:\B\b.exe" "%1"`, `https://example.com/a\`)
	if err != nil {
		t.Fatal(err)
	}
	if want := `"C:\B\b.exe" "https://example.com/a\\"`; line != want {
		t.Fatalf("got %q, want %q", line, want)
	}
}

func TestWindowsCommandErrors(t *testing.T) {
	for _, tmpl := range []string{"", `"C:\unterminated.exe %1`} {
		if _, _, err := WindowsCommandLine(tmpl, u); err == nil {
			t.Errorf("%q: expected error", tmpl)
		}
	}
}

func TestWindowsCommandLineExtraArgs(t *testing.T) {
	_, line, err := WindowsCommandLine(`"C:\Chrome\chrome.exe" --single-argument %1`, u, "--incognito", "--profile-directory=Profile 1")
	if err != nil {
		t.Fatal(err)
	}
	if want := `"C:\Chrome\chrome.exe" --incognito "--profile-directory=Profile 1" --single-argument ` + u; line != want {
		t.Fatalf("got  %q\nwant %q", line, want)
	}
}

func TestQuoteArg(t *testing.T) {
	for in, want := range map[string]string{
		"plain":      "plain",
		"two words":  `"two words"`,
		"":           `""`,
		`say "hi"`:   `"say \"hi\""`,
		`dir\ name\`: `"dir\ name\\"`,
		`a\"b c`:     `"a\\\"b c"`,
	} {
		if got := quoteArg(in); got != want {
			t.Errorf("%q: got %s, want %s", in, got, want)
		}
	}
}
