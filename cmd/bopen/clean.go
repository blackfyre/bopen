package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/blackfyre/bopen/internal/clean"
	"github.com/blackfyre/bopen/internal/launch"
)

// runClean implements `bopen clean [--explain] [URL...]`: it prints each
// link with the default suggestions applied, reading links from stdin
// when none are given. It returns the exit status.
func runClean(args []string, rules []clean.Rule, stdin io.Reader, stdout, stderr io.Writer) int {
	explain := false
	var inputs []string
	for _, a := range args {
		if a == "--explain" {
			explain = true
			continue
		}
		inputs = append(inputs, a)
	}
	status := 0
	process := func(in string) {
		valid, err := launch.Validate(in)
		if err != nil {
			fmt.Fprintln(stdout, in)
			fmt.Fprintf(stderr, "bopen clean: %q: %v\n", in, err)
			status = 1
			return
		}
		a := clean.Analyse(valid, rules)
		accepted := a.Defaults()
		fmt.Fprintln(stdout, a.Clean(accepted))
		if !explain {
			return
		}
		for i, s := range a.Suggestions {
			if !accepted[i] || !a.Available(i, accepted) {
				continue
			}
			fmt.Fprintf(stderr, "removed %s (%s, %s): %s\n", s.Text, s.Kind, s.Source, s.Reason)
		}
	}
	if len(inputs) > 0 {
		for _, in := range inputs {
			process(in)
		}
		return status
	}
	sc := bufio.NewScanner(stdin)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			process(line)
		}
	}
	if err := sc.Err(); err != nil {
		fmt.Fprintln(stderr, "bopen clean:", err)
		return 1
	}
	return status
}
