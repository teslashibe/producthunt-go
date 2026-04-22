package mcp

import (
	"fmt"
	"time"

	"github.com/teslashibe/mcptool"
)

// parseDate accepts YYYY-MM-DD or RFC3339 and returns the corresponding UTC
// time. Empty input returns a zero time and an error.
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, &mcptool.Error{
			Code:    "invalid_input",
			Message: "date must be non-empty (YYYY-MM-DD or RFC3339)",
		}
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, &mcptool.Error{
		Code:    "invalid_input",
		Message: fmt.Sprintf("invalid date %q (want YYYY-MM-DD or RFC3339)", s),
	}
}
