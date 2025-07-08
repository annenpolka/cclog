package formatter

import "time"

// GetSystemTimezone returns the system's local timezone
// This uses time.Local which automatically handles:
// - TZ environment variable
// - System timezone configuration (/etc/localtime on Unix)
// - Falls back to UTC if system timezone cannot be determined
func GetSystemTimezone() *time.Location {
	return time.Local
}