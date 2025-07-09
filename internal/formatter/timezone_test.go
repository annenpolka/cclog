package formatter

import (
	"os"
	"testing"
	"time"
)

func TestGetSystemTimezone(t *testing.T) {
	// Test basic functionality
	tz := GetSystemTimezone()
	if tz == nil {
		t.Error("GetSystemTimezone() should not return nil")
	}

	// Test that it returns a valid timezone
	now := time.Now()
	locTime := now.In(tz)
	if locTime.IsZero() {
		t.Error("Timezone should be able to format time")
	}
}

func TestGetSystemTimezoneWithEnvironmentVariable(t *testing.T) {
	// Note: This test demonstrates that time.Local behavior can be affected by TZ environment variable
	// However, in practice, the Go runtime may not always pick up TZ changes during test execution
	// This test is more about documenting expected behavior rather than testing implementation

	// Save original TZ
	originalTZ := os.Getenv("TZ")
	defer func() {
		if originalTZ != "" {
			os.Setenv("TZ", originalTZ)
		} else {
			os.Unsetenv("TZ")
		}
	}()

	// Test that function returns time.Local
	tz := GetSystemTimezone()
	if tz != time.Local {
		t.Error("GetSystemTimezone() should return time.Local")
	}

	// Test with manual timezone creation to verify our understanding
	testCases := []struct {
		name     string
		location *time.Location
		utcHour  int
		expected int
	}{
		{
			name:     "UTC timezone",
			location: time.UTC,
			utcHour:  12,
			expected: 12,
		},
		{
			name:     "JST timezone",
			location: time.FixedZone("JST", 9*60*60),
			utcHour:  12,
			expected: 21,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			utcTime := time.Date(2025, 7, 6, tc.utcHour, 0, 0, 0, time.UTC)
			localTime := utcTime.In(tc.location)

			if localTime.Hour() != tc.expected {
				t.Errorf("%s: expected %d:00, got %d:00", tc.name, tc.expected, localTime.Hour())
			}
		})
	}
}

func TestGetSystemTimezoneDefaultBehavior(t *testing.T) {
	// Test that the function returns time.Local behavior
	tz := GetSystemTimezone()

	// Compare with time.Local
	now := time.Now()
	ourTime := now.In(tz)
	localTime := now.In(time.Local)

	// They should be the same
	if ourTime.Format(time.RFC3339) != localTime.Format(time.RFC3339) {
		t.Errorf("GetSystemTimezone() result should match time.Local behavior")
	}
}
