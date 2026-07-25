package dates

import (
	"testing"
	"time"
)

func TestStrftimeFormat(t *testing.T) {
	// Expected strings verified with Python 3:
	// datetime(2024, 7, 9, 5, 3, 7, 12500).strftime(fmt)
	ts := time.Date(2024, 7, 9, 5, 3, 7, 12500*1000, time.UTC)
	cases := []struct {
		format string
		want   string
	}{
		{"%Y-%m-%d", "2024-07-09"},
		{"%y-%m", "24-07"},
		{"%d %b %Y", "09 Jul 2024"},
		{"%B %A %a", "July Tuesday Tue"},
		{"%H:%M:%S", "05:03:07"},
		{"%H:%M:%S.%f", "05:03:07.012500"},
		{"%I %p", "05 AM"},
		{"%j", "191"},
		{"%w", "2"},
		{"%-d.%-m.%Y", "9.7.2024"},
		{"100%%", "100%"},
		{"plain text", "plain text"},
		{"%q", "%q"}, // unknown directive stays verbatim
	}
	for _, tc := range cases {
		if got := strftimeFormat(ts, tc.format); got != tc.want {
			t.Errorf("strftimeFormat(%q) = %q, want %q", tc.format, got, tc.want)
		}
	}
}

func TestStrftimeFormatPM(t *testing.T) {
	ts := time.Date(2024, 7, 9, 13, 0, 0, 0, time.UTC)
	if got := strftimeFormat(ts, "%I %p"); got != "01 PM" {
		t.Errorf("strftimeFormat(%%I %%p) = %q, want %q", got, "01 PM")
	}
	midnight := time.Date(2024, 7, 9, 0, 0, 0, 0, time.UTC)
	if got := strftimeFormat(midnight, "%I %p"); got != "12 AM" {
		t.Errorf("strftimeFormat(%%I %%p) at midnight = %q, want %q", got, "12 AM")
	}
}
