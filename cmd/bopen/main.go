// Command bopen inspects a link, suggests removing tracking parts, and opens
// it in the browser the user chooses.
package main

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/blackfyre/bopen/internal/app"
	"github.com/blackfyre/bopen/internal/appearance"
	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/clearurls"
	"github.com/blackfyre/bopen/internal/discovery"
	"github.com/blackfyre/bopen/internal/launch"
	"github.com/blackfyre/bopen/internal/prefs"
	"github.com/blackfyre/bopen/internal/register"
	"github.com/blackfyre/bopen/internal/shortlinks"
	"github.com/blackfyre/bopen/internal/ui"
)

// version is set at release time with -ldflags "-X main.version=...".
var version = "dev"

// appVersion is the version bopen reports: the linked-in release version,
// else the module version recorded by `go install`, else "dev".
func appVersion() string {
	return resolveVersion(version, debug.ReadBuildInfo)
}

func resolveVersion(linked string, buildInfo func() (*debug.BuildInfo, bool)) string {
	if linked != "" && linked != "dev" {
		return linked
	}
	if info, ok := buildInfo(); ok && info != nil {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}
	return "dev"
}

func usageText() string {
	return "bopen " + appVersion() + "\n" + usage
}

const usage = `usage:
  bopen <url>       inspect a link and open it in a browser
  bopen register    make bopen the default handler for web links
  bopen unregister  undo 'bopen register'
  bopen settings    change bopen's preferences
  bopen clean [--explain] [url...]
                    print links without tracking parts (reads stdin
                    when no url is given)
`

type mode int

const (
	modeUsage mode = iota
	modeOpen
	modeRegister
	modeUnregister
	modeSettings
	modeClean
)

func parseArgs(args []string) mode {
	if len(args) >= 1 && args[0] == "clean" {
		return modeClean
	}
	if len(args) != 1 || args[0] == "" {
		return modeUsage
	}
	switch args[0] {
	case "register":
		return modeRegister
	case "unregister":
		return modeUnregister
	case "settings":
		return modeSettings
	case "-h", "-help", "--help", "help":
		return modeUsage
	}
	return modeOpen
}

func main() {
	switch parseArgs(os.Args[1:]) {
	case modeUsage:
		attachConsole()
		fmt.Fprint(os.Stderr, usageText())
		os.Exit(2)
	case modeRegister, modeUnregister:
		attachConsole()
		os.Exit(runRegistration(os.Args[1], os.Stdout, os.Stderr))
	case modeClean:
		attachConsole()
		env, problems := loadEnv()
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "bopen: configuration problem:", p)
		}
		os.Exit(runClean(os.Args[2:], env.Rules(), os.Stdin, os.Stdout, os.Stderr))
	case modeSettings:
		env, _ := loadEnv()
		ui.RunSettings(env)
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

// loadEnv gathers preferences, state, browsers, rules and the registrar.
// Problems with the configuration are returned rather than being fatal.
func loadEnv() (*ui.Env, []error) {
	env := &ui.Env{Config: prefs.DefaultConfig(), Appearance: appearance.Read()}
	var problems []error
	if dir, err := prefs.Dir(); err != nil {
		problems = append(problems, err)
	} else {
		env.ConfigDir = dir
		env.Config, problems = prefs.LoadConfig(dir)
		env.State = prefs.LoadState(dir)
	}
	env.AllBrowsers = discovery.Discover()
	rules, err := clean.Builtin()
	if err != nil {
		problems = append(problems, fmt.Errorf("built-in rules: %w", err))
	}
	env.Builtin = rules
	if r, err := register.System(); err == nil {
		env.Registrar = r
	}
	env.Expand = shortlinks.New(appVersion()).Expand
	if cache, err := clearurls.SystemCache(); err == nil {
		env.ClearURLs = &ui.ClearURLs{
			Cache:   cache,
			Fetcher: clearurls.NewFetcher(appVersion()),
			Update:  clearurls.Update,
			Meta:    cache.Meta(),
		}
		if env.Config.Rules.ClearURLs {
			env.ClearURLs.Rules = clearurls.Load(cache)
		}
	}
	return env, problems
}

// prepare builds the inspector model for input.
func prepare(input string) (app.Situation, *ui.Model) {
	env, problems := loadEnv()
	m := &ui.Model{Input: input, Env: env}
	for _, p := range problems {
		m.Problems = append(m.Problems, "Configuration problem: "+p.Error()+" (defaults are used instead)")
	}
	url, verr := launch.Validate(input)
	if verr != nil {
		m.Invalid = true
	} else {
		m.Analysis = clean.Analyse(url, env.Rules())
		m.Accepted = m.Analysis.Defaults()
	}
	m.Refresh()
	m.Open = func(b discovery.Browser, url string) error {
		if env.ConfigDir == "" {
			return launch.Start(b, url)
		}
		return app.Open(env.ConfigDir, env.State, b, url, launch.Start)
	}
	s := app.Situation{
		Config:        env.Config,
		Problems:      problems,
		ValidationErr: verr,
		Analysis:      m.Analysis,
		Browsers:      m.Browsers,
		Direct:        m.Site != nil && m.Site.Direct,
	}
	return s, m
}
