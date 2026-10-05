package desktopentry

import (
	"errors"
	"strings"
)

type token struct {
	text   string
	quoted bool
}

// Args expands the entry's Exec key for opening url and returns the argument
// vector, program first. url always becomes exactly one argument. When Exec
// has no %u or %U field code, url is appended as the final argument.
func (e Entry) Args(url string) ([]string, error) {
	tokens, err := tokenise(e.Exec)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, errors.New("empty Exec")
	}
	var args []string
	hasURL := false
	for i, t := range tokens {
		if t.quoted || i == 0 {
			args = append(args, t.text)
			continue
		}
		if t.text == "%i" {
			if e.Icon != "" {
				args = append(args, "--icon", e.Icon)
			}
			continue
		}
		expanded, sawURL := e.expand(t.text, url)
		hasURL = hasURL || sawURL
		if expanded == "" && t.text != "" {
			continue
		}
		args = append(args, expanded)
	}
	if !hasURL {
		args = append(args, url)
	}
	return args, nil
}

// expand replaces the field codes in an unquoted argument.
func (e Entry) expand(arg, url string) (string, bool) {
	if !strings.Contains(arg, "%") {
		return arg, false
	}
	var b strings.Builder
	sawURL := false
	for i := 0; i < len(arg); i++ {
		if arg[i] != '%' || i+1 == len(arg) {
			b.WriteByte(arg[i])
			continue
		}
		i++
		switch arg[i] {
		case 'u', 'U':
			b.WriteString(url)
			sawURL = true
		case 'c':
			b.WriteString(e.Name)
		case 'k':
			b.WriteString(e.Path)
		case '%':
			b.WriteByte('%')
		default:
			// %f %F %d %D %n %N %v %m, %i inside an argument, and unknown
			// codes expand to nothing.
		}
	}
	return b.String(), sawURL
}

// tokenise splits an Exec value into arguments following the quoting rules
// of the Desktop Entry specification.
func tokenise(s string) ([]token, error) {
	var tokens []token
	var cur strings.Builder
	inToken, inQuote, quoted := false, false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQuote && c == '\\' && i+1 < len(s) && strings.IndexByte("\"`$\\", s[i+1]) >= 0:
			i++
			cur.WriteByte(s[i])
		case c == '"':
			inQuote = !inQuote
			inToken, quoted = true, true
		case !inQuote && (c == ' ' || c == '\t' || c == '\n'):
			if inToken {
				tokens = append(tokens, token{cur.String(), quoted})
				cur.Reset()
				inToken, quoted = false, false
			}
		default:
			cur.WriteByte(c)
			inToken = true
		}
	}
	if inQuote {
		return nil, errors.New("unterminated quote in Exec")
	}
	if inToken {
		tokens = append(tokens, token{cur.String(), quoted})
	}
	return tokens, nil
}

// QuoteArg quotes arg for use in an Exec value, including the string-level
// escaping needed when the value is written to a desktop file.
func QuoteArg(arg string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(arg); i++ {
		switch c := arg[i]; c {
		case '"', '`', '$':
			b.WriteString(`\\`)
			b.WriteByte(c)
		case '\\':
			b.WriteString(`\\\\`)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}
