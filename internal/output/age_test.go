package output

import (
	"testing"
	"time"
)

func TestShortDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{0, "0s"},
		{-time.Hour, "0s"},
		{45 * time.Second, "45s"},
		{59 * time.Second, "59s"},
		{60 * time.Second, "1m"},
		{5 * time.Minute, "5m"},
		{59*time.Minute + 59*time.Second, "59m"},
		{60 * time.Minute, "1h"},
		{2 * time.Hour, "2h"},
		{23*time.Hour + 59*time.Minute, "23h"},
		{24 * time.Hour, "1d"},
		{5 * 24 * time.Hour, "5d"},
		{364 * 24 * time.Hour, "364d"},
		{365 * 24 * time.Hour, "1y"},
		{3 * 365 * 24 * time.Hour, "3y"},
	}
	for _, tc := range cases {
		if got := ShortDuration(tc.in); got != tc.want {
			t.Errorf("ShortDuration(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
