package content

import (
	"unicode/utf16"
	"unicode/utf8"
)

// PartialContent parses a prefix of the tool input JSON. It never fails. It
// reads the three known keys and closes an open string at the cut point. It
// drops an incomplete escape or an incomplete rune.
func PartialContent(prefix []byte) Content {
	var c Content
	i := skipSpace(prefix, 0)
	if i >= len(prefix) || prefix[i] != '{' {
		return c
	}
	i++
	for {
		i = skipSpace(prefix, i)
		if i < len(prefix) && prefix[i] == ',' {
			i = skipSpace(prefix, i+1)
		}
		if i >= len(prefix) || prefix[i] != '"' {
			return c
		}
		key, n, done := scanString(prefix, i+1)
		if !done {
			return c
		}
		i = skipSpace(prefix, n)
		if i >= len(prefix) || prefix[i] != ':' {
			return c
		}
		i = skipSpace(prefix, i+1)
		if i >= len(prefix) || prefix[i] != '"' {
			return c
		}
		val, n, done := scanString(prefix, i+1)
		switch key {
		case "sms":
			c.SMS = val
		case "email_subject":
			c.EmailSubject = val
		case "email_body":
			c.EmailBody = val
		}
		if !done {
			return c
		}
		i = n
	}
}

func skipSpace(b []byte, i int) int {
	for i < len(b) && (b[i] == ' ' || b[i] == '\n' || b[i] == '\t' || b[i] == '\r') {
		i++
	}
	return i
}

// scanString reads a JSON string body that starts at i, after the opening
// quote. It returns the decoded text, the index after the closing quote and
// whether the closing quote was found.
func scanString(b []byte, i int) (string, int, bool) {
	var out []byte
	for i < len(b) {
		ch := b[i]
		switch {
		case ch == '"':
			return string(out), i + 1, true
		case ch == '\\':
			if i+1 >= len(b) {
				return string(out), len(b), false
			}
			switch e := b[i+1]; e {
			case 'n':
				out, i = append(out, '\n'), i+2
			case 't':
				out, i = append(out, '\t'), i+2
			case 'r':
				out, i = append(out, '\r'), i+2
			case 'b':
				out, i = append(out, '\b'), i+2
			case 'f':
				out, i = append(out, '\f'), i+2
			case '"', '\\', '/':
				out, i = append(out, e), i+2
			case 'u':
				r, ok := hex4(b, i+2)
				if !ok {
					return string(out), len(b), false
				}
				i += 6
				if utf16.IsSurrogate(r) {
					if r >= 0xDC00 { // lone low surrogate
						out = utf8.AppendRune(out, utf8.RuneError)
						continue
					}
					if i+1 >= len(b) || (b[i] == '\\' && i+6 > len(b)) {
						return string(out), len(b), false // pair may still complete
					}
					if b[i] == '\\' && b[i+1] == 'u' {
						if r2, ok := hex4(b, i+2); ok && r2 >= 0xDC00 && r2 < 0xE000 {
							out = utf8.AppendRune(out, utf16.DecodeRune(r, r2))
							i += 6
							continue
						}
					}
					out = utf8.AppendRune(out, utf8.RuneError)
					continue
				}
				out = utf8.AppendRune(out, r)
			default:
				return string(out), len(b), false // invalid escape
			}
		default:
			r, size := utf8.DecodeRune(b[i:])
			if r == utf8.RuneError && size == 1 && !utf8.FullRune(b[i:]) {
				return string(out), len(b), false
			}
			out = append(out, b[i:i+size]...)
			i += size
		}
	}
	return string(out), len(b), false
}

func hex4(b []byte, i int) (rune, bool) {
	if i+4 > len(b) {
		return 0, false
	}
	var r rune
	for _, c := range b[i : i+4] {
		switch {
		case c >= '0' && c <= '9':
			r = r<<4 | rune(c-'0')
		case c >= 'a' && c <= 'f':
			r = r<<4 | rune(c-'a'+10)
		case c >= 'A' && c <= 'F':
			r = r<<4 | rune(c-'A'+10)
		default:
			return 0, false
		}
	}
	return r, true
}
