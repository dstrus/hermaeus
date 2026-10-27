package store

import "time"

// Timestamp reads a real-world time as SkyrimNet stores it, returning Unix
// milliseconds, or nil for none. SkyrimNet uses several forms: Unix seconds
// (memories), Julian days (character records, and the default of several
// columns), and text such as "2026-10-15 04:54:35" in UTC (re-identifications).
// Scan the column into an any to pass it here.
func Timestamp(v any) *int64 {
	var ms int64
	switch t := v.(type) {
	case float64:
		switch {
		case t <= 0:
			return nil
		case t > 1e8: // Unix seconds
			ms = int64(t * 1000)
		default: // a Julian day
			ms = int64((t - 2440587.5) * 86400000)
		}
	case int64:
		return Timestamp(float64(t))
	case string:
		parsed, err := time.Parse(time.DateTime, t)
		if err != nil {
			return nil
		}
		ms = parsed.UnixMilli()
	case []byte:
		return Timestamp(string(t))
	case time.Time:
		ms = t.UnixMilli()
	default:
		return nil
	}
	return &ms
}
