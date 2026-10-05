package launch

import (
	"errors"
	"strings"
)

// WindowsCommandLine builds the command line that opens url with a registered
// Windows command template such as `"C:\...\chrome.exe" --single-argument %1`.
// It returns the executable and the full command line.
//
// url must come from Validate: it then contains no whitespace or double
// quotes, so it is a single argument wherever it is inserted. A placeholder
// already wrapped in quotes stays quoted; an unquoted placeholder stays
// unquoted, because some browsers (Chrome's --single-argument) take the rest
// of the command line verbatim and would see the quotes as part of the URL.
// Without a placeholder the URL is appended.
//
// extra arguments (a private-window flag, profile arguments) are inserted
// directly after the executable, quoted when needed, so they precede an
// argument such as --single-argument that takes the rest of the line.
func WindowsCommandLine(template, url string, extra ...string) (string, string, error) {
	template = strings.TrimSpace(template)
	exe, rest, err := splitExecutable(template)
	if err != nil {
		return "", "", err
	}
	quotedURL := url
	// Inside quotes, backslashes before the closing quote must be doubled.
	if trimmed := strings.TrimRight(url, `\`); len(trimmed) < len(url) {
		quotedURL = trimmed + strings.Repeat(`\`, 2*(len(url)-len(trimmed)))
	}
	replaced := false
	var b strings.Builder
	for i := 0; i < len(rest); i++ {
		if rest[i] == '%' && i+1 < len(rest) {
			switch rest[i+1] {
			case '1', 'L', 'l':
				inQuotes := strings.Count(rest[:i], `"`)%2 == 1
				if inQuotes {
					b.WriteString(quotedURL)
				} else {
					b.WriteString(url)
				}
				replaced = true
				i++
				continue
			case '*':
				i++
				continue
			}
		}
		b.WriteByte(rest[i])
	}
	head := template[:len(template)-len(rest)]
	for _, a := range extra {
		head += " " + quoteArg(a)
	}
	line := head + b.String()
	if !replaced {
		line = strings.TrimRight(line, " ") + " " + url
	}
	return exe, line, nil
}

// quoteArg quotes a for a Windows command line when it contains spaces,
// tabs or quotes, following the CommandLineToArgvW rules.
func quoteArg(a string) string {
	if a != "" && !strings.ContainsAny(a, " \t\"") {
		return a
	}
	var b strings.Builder
	b.WriteByte('"')
	slashes := 0
	for i := 0; i < len(a); i++ {
		switch a[i] {
		case '\\':
			slashes++
		case '"':
			b.WriteString(strings.Repeat(`\`, slashes+1))
			slashes = 0
		default:
			slashes = 0
		}
		b.WriteByte(a[i])
	}
	b.WriteString(strings.Repeat(`\`, slashes))
	b.WriteByte('"')
	return b.String()
}

// splitExecutable separates the executable from the arguments of a command
// template. It returns the executable and the remainder of the template.
func splitExecutable(template string) (string, string, error) {
	if template == "" {
		return "", "", errors.New("empty command")
	}
	if template[0] == '"' {
		end := strings.IndexByte(template[1:], '"')
		if end < 0 {
			return "", "", errors.New("unterminated quote in command")
		}
		return template[1 : end+1], template[end+2:], nil
	}
	if sp := strings.IndexAny(template, " \t"); sp >= 0 {
		return template[:sp], template[sp:], nil
	}
	return template, "", nil
}
