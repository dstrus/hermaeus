package store

import (
	"fmt"
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	for _, c := range []struct {
		in   any
		want string
	}{
		{1791685096.5, "1791685096500"},      // Unix seconds
		{int64(1791685096), "1791685096000"}, // the same, read as an integer
		{2440587.5, "0"},                     // the Julian day of the Unix epoch
		{2461335.5, "1792627200000"},         // midnight UTC on 2026-10-22
		{"2026-10-15 04:54:35", "1792040075000"},
		{[]byte("2026-10-15 04:54:35"), "1792040075000"},
		{time.Date(2026, 10, 15, 4, 54, 35, 0, time.UTC), "1792040075000"},
		{0.0, "none"}, {-1.0, "none"}, {nil, "none"}, {"yesterday", "none"},
	} {
		got := "none"
		if ms := Timestamp(c.in); ms != nil {
			got = fmt.Sprint(*ms)
		}
		if got != c.want {
			t.Errorf("Timestamp(%#v) = %s, want %s", c.in, got, c.want)
		}
	}
}
