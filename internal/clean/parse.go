package clean

import (
	"net/url"
	"strings"
)

// ParseTolerant parses s like url.Parse, but treats a '%' that does not
// start a valid escape as a literal character instead of failing, as
// browsers do.
func ParseTolerant(s string) (*url.URL, error) {
	return url.Parse(escapeStrayPercents(s))
}

func escapeStrayPercents(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			_, ok1 := unhexAt(s, i+1)
			_, ok2 := unhexAt(s, i+2)
			if !ok1 || !ok2 {
				b.WriteString("%25")
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func unhexAt(s string, i int) (byte, bool) {
	if i >= len(s) {
		return 0, false
	}
	return unhex(s[i])
}
