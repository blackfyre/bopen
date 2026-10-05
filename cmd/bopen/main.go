// Command bopen inspects a link, suggests removing tracking parts, and opens
// it in the browser the user chooses.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/blackfyre/bopen/internal/app"
	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/launch"
	"github.com/blackfyre/bopen/internal/prefs"
	"github.com/blackfyre/bopen/internal/register"
	"github.com/blackfyre/bopen/internal/ui"
)

const usage = `usage:
  bopen <url>       inspect a link and open it in a browser
  bopen register    make bopen the default handler for web links
  bopen unregister  undo 'bopen register'
`

type mode int

const (
	modeUsage mode = iota
	modeOpen
	modeRegister
	modeUnregister
)

func parseArgs(args []string) mode {
	if len(args) != 1 || args[0] == "" {
		return modeUsage
	}
	switch args[0] {
	case "register":
		return modeRegister
	case "unregister":
		return modeUnregister
	case "-h", "-help", "--help", "help":
		return modeUsage
	}
	return modeOpen
}

func main() {
	switch parseArgs(os.Args[1:]) {
	case modeUsage:
		attachConsole()
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	case modeRegister, modeUnregister:
		attachConsole()
		os.Exit(runRegistration(os.Args[1], os.Stdout, os.Stderr))
	case modeOpen:
		runOpen(os.Args[1])
	}
}

func runRegistration(cmd string, stdout, stderr io.Writer) int {
	r, err := register.System()
	if err == nil {
		var msg string
		if cmd == "register" {
			msg, err = r.Register()
		} else {
			msg, err = r.Unregister()
		}
		if err == nil {
			fmt.Fprintln(stdout, msg)
			return 0
		}
	}
	fmt.Fprintln(stderr, "bopen:", err)
	return 1
}

func runOpen(input string) {
	s, m := prepare(input)
	if !app.NeedWindow(s) {
		// Silent path: the link has nothing to suggest, so it is opened unchanged.
		if m.OpenSelected() {
			return
		}
	}
	ui.Run(m)
}

// prepare gathers preferences, browsers and the analysis for input.
func prepare(input string) (app.Situation, *ui.Model) {
	var s app.Situation
	m := &ui.Model{Input: input}

	s.Config = prefs.DefaultConfig()
	var st prefs.State
	dir, err := prefs.Dir()
	if err != nil {
		s.Problems = append(s.Problems, err)
	} else {
		s.Config, s.Problems = prefs.LoadConfig(dir)
		st = prefs.LoadState(dir)
	}

	s.Browsers = discovery.Discover()
	m.Browsers = s.Browsers
	m.Selected = app.Preselect(s.Browsers, st)

	rules, err := clean.Builtin()
	if err != nil {
		s.Problems = append(s.Problems, fmt.Errorf("built-in rules: %w", err))
	}
	url, verr := launch.Validate(input)
	s.ValidationErr = verr
	if verr == nil {
		s.Analysis = clean.Analyse(url, rules)
		m.Analysis = s.Analysis
		m.Accepted = s.Analysis.Defaults()
	}

	for _, p := range s.Problems {
		m.Problems = append(m.Problems, "Configuration problem: "+p.Error()+" (defaults are used instead)")
	}
	switch {
	case verr != nil:
		m.Blocker = "This link cannot be opened: only http and https links are accepted."
	case len(s.Browsers) == 0:
		m.Blocker = "No web browsers were found on this system."
	}
	m.Open = func(b discovery.Browser, url string) error {
		if dir == "" {
			return launch.Start(b, url)
		}
		return app.Open(dir, st, b, url, launch.Start)
	}
	return s, m
}
