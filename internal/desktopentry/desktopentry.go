// Package desktopentry parses freedesktop.org Desktop Entry files and expands
// their Exec keys into argument vectors.
package desktopentry

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

// Entry holds the [Desktop Entry] keys bopen needs.
type Entry struct {
	// Path is the file the entry was read from; it expands the %k field code.
	Path      string
	Type      string
	Name      string
	Exec      string
	TryExec   string
	Icon      string
	MimeTypes []string
	Hidden    bool
	NoDisplay bool
}

// HasMimeType reports whether the entry lists mimeType.
func (e Entry) HasMimeType(mimeType string) bool {
	for _, m := range e.MimeTypes {
		if m == mimeType {
			return true
		}
	}
	return false
}

// ParseFile parses the desktop entry at path, localising Name for lang
// (a POSIX locale such as "en_GB.UTF-8").
func ParseFile(path, lang string) (Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return Entry{}, err
	}
	defer f.Close()
	e, err := Parse(f, lang)
	e.Path = path
	return e, err
}

// Parse parses a desktop entry, reading only the [Desktop Entry] group.
func Parse(r io.Reader, lang string) (Entry, error) {
	var e Entry
	names := map[string]string{}
	inGroup, seenGroup := false, false
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		if line[0] == '[' {
			if inGroup {
				break
			}
			inGroup = line == "[Desktop Entry]"
			seenGroup = seenGroup || inGroup
			continue
		}
		if !inGroup {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch {
		case key == "Type":
			e.Type = unescape(value)
		case key == "Name":
			names[""] = unescape(value)
		case strings.HasPrefix(key, "Name[") && strings.HasSuffix(key, "]"):
			names[key[len("Name["):len(key)-1]] = unescape(value)
		case key == "Exec":
			e.Exec = unescape(value)
		case key == "TryExec":
			e.TryExec = unescape(value)
		case key == "Icon":
			e.Icon = unescape(value)
		case key == "MimeType":
			e.MimeTypes = splitList(value)
		case key == "Hidden":
			e.Hidden = value == "true"
		case key == "NoDisplay":
			e.NoDisplay = value == "true"
		}
	}
	if err := sc.Err(); err != nil {
		return e, err
	}
	if !seenGroup {
		return e, errors.New("no [Desktop Entry] group")
	}
	e.Name = localised(names, lang)
	return e, nil
}

// localised picks the Name variant for lang following the Desktop Entry
// specification's matching order.
func localised(names map[string]string, lang string) string {
	// lang_COUNTRY.ENCODING@MODIFIER; the encoding is ignored.
	base, modifier, _ := strings.Cut(lang, "@")
	base, _, _ = strings.Cut(base, ".")
	language, country, _ := strings.Cut(base, "_")
	var candidates []string
	if country != "" && modifier != "" {
		candidates = append(candidates, language+"_"+country+"@"+modifier)
	}
	if country != "" {
		candidates = append(candidates, language+"_"+country)
	}
	if modifier != "" {
		candidates = append(candidates, language+"@"+modifier)
	}
	if language != "" && language != "C" && language != "POSIX" {
		candidates = append(candidates, language)
	}
	for _, c := range candidates {
		if v, ok := names[c]; ok {
			return v
		}
	}
	return names[""]
}

// unescape applies the string escapes of the specification: \s \n \t \r \\.
func unescape(s string) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' || i+1 == len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch s[i] {
		case 's':
			b.WriteByte(' ')
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '\\':
			b.WriteByte('\\')
		default:
			b.WriteByte('\\')
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// splitList splits a ';'-separated list value, honouring "\;" escapes.
func splitList(s string) []string {
	var items []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '\\' && i+1 < len(s) && s[i+1] == ';':
			cur.WriteByte(';')
			i++
		case s[i] == ';':
			items = append(items, unescape(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(s[i])
		}
	}
	if cur.Len() > 0 {
		items = append(items, unescape(cur.String()))
	}
	return items
}
