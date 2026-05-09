package output

import (
	"fmt"
	"time"
)

// Age returns a kubectl-style relative duration since t (e.g. "5s", "3m",
// "2h", "5d", "1y"). Negative durations clamp to "0s".
func Age(t time.Time) string {
	return ShortDuration(time.Since(t))
}

// ShortDuration formats d as a single most-significant unit:
// seconds, minutes, hours, days, or years.
func ShortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	sec := int64(d.Seconds())
	switch {
	case sec < 60:
		return fmt.Sprintf("%ds", sec)
	case sec < 60*60:
		return fmt.Sprintf("%dm", sec/60)
	case sec < 60*60*24:
		return fmt.Sprintf("%dh", sec/3600)
	case sec < 60*60*24*365:
		return fmt.Sprintf("%dd", sec/86400)
	default:
		return fmt.Sprintf("%dy", sec/(86400*365))
	}
}
