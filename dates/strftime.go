package dates

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// strftimeFormat renders t using a C89/Python-style strftime format string.
// The date.autoformatter.* rcParams carry strftime formats, which Go's
// time.Format layout language cannot express in full (e.g. %j, %f), so this
// interprets the directives directly instead of converting to a layout.
//
// The supported set mirrors what matplotlib documents for its date
// formatters; a "-" flag (glibc style, e.g. %-d) suppresses zero padding.
// Unrecognized directives are emitted verbatim.
func strftimeFormat(t time.Time, format string) string {
	var b strings.Builder
	b.Grow(len(format) + 8)
	for i := 0; i < len(format); i++ {
		c := format[i]
		if c != '%' || i+1 >= len(format) {
			b.WriteByte(c)
			continue
		}
		i++
		noPad := false
		if format[i] == '-' && i+1 < len(format) {
			noPad = true
			i++
		}
		verb := format[i]
		pad := func(v, width int) {
			if noPad {
				b.WriteString(strconv.Itoa(v))
				return
			}
			fmt.Fprintf(&b, "%0*d", width, v)
		}
		switch verb {
		case 'Y':
			pad(t.Year(), 4)
		case 'y':
			pad(t.Year()%100, 2)
		case 'm':
			pad(int(t.Month()), 2)
		case 'd':
			pad(t.Day(), 2)
		case 'e':
			fmt.Fprintf(&b, "%2d", t.Day())
		case 'H':
			pad(t.Hour(), 2)
		case 'I':
			hour := t.Hour() % 12
			if hour == 0 {
				hour = 12
			}
			pad(hour, 2)
		case 'M':
			pad(t.Minute(), 2)
		case 'S':
			pad(t.Second(), 2)
		case 'f':
			// Python strftime microseconds, always six digits.
			fmt.Fprintf(&b, "%06d", t.Nanosecond()/1000)
		case 'j':
			pad(t.YearDay(), 3)
		case 'a':
			b.WriteString(t.Format("Mon"))
		case 'A':
			b.WriteString(t.Format("Monday"))
		case 'b', 'h':
			b.WriteString(t.Format("Jan"))
		case 'B':
			b.WriteString(t.Format("January"))
		case 'p':
			if t.Hour() < 12 {
				b.WriteString("AM")
			} else {
				b.WriteString("PM")
			}
		case 'w':
			b.WriteString(strconv.Itoa(int(t.Weekday())))
		case 'z':
			b.WriteString(t.Format("-0700"))
		case 'Z':
			b.WriteString(t.Format("MST"))
		case '%':
			b.WriteByte('%')
		default:
			// Unknown directive: emit it verbatim (including any "-" flag).
			b.WriteByte('%')
			if noPad {
				b.WriteByte('-')
			}
			b.WriteByte(verb)
		}
	}
	return b.String()
}
