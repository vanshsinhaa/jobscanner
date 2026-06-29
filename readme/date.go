package readme

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// dateFormats lists all absolute date formats we expect from scrapers
// (Microsoft, Amazon, Apple, Oracle). Tried in order; first match wins.
var dateFormats = []string{
	"2006-01-02T15:04:05Z07:00", // ISO 8601 with timezone offset
	"2006-01-02T15:04:05Z",      // ISO 8601 UTC
	"2006-01-02T15:04:05",       // ISO 8601 no timezone
	"2006-01-02",                // date only
	"01/02/2006",                // MM/DD/YYYY
	"January 2, 2006",           // long form
	"Jan 2, 2006",               // short month
}

// workdayRelRe matches "N Days Ago" and "N+ Days Ago" (case-insensitive).
// The + is discarded — "30+ Days Ago" is treated as exactly 30 days old.
var workdayRelRe = regexp.MustCompile(`(\d+)\+?\s+[Dd]ays?\s+[Aa]go`)

// parsePostingDate converts a raw PostedOn string to a time.Time for
// internal sort and filter logic only — never shown in the README.
// Returns (zero, false) for empty or unrecognized input.
func parsePostingDate(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}

	lower := strings.ToLower(raw)

	// Workday: "Posted Today"
	if strings.Contains(lower, "today") {
		return time.Now(), true
	}

	// Workday: "Posted N Days Ago" or "Posted N+ Days Ago"
	if m := workdayRelRe.FindStringSubmatch(raw); len(m) > 1 {
		if n, err := strconv.Atoi(m[1]); err == nil {
			return time.Now().AddDate(0, 0, -n), true
		}
	}

	// Absolute ISO and locale formats (Microsoft, Amazon, Apple, Oracle)
	for _, format := range dateFormats {
		if t, err := time.Parse(format, raw); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

// displayDate formats a raw PostedOn string for the README table cell.
//
//   - Workday relative ("Posted 5 Days Ago") → "5 Days Ago"
//   - Workday relative ("Posted Today")       → "Today"
//   - ISO date (Microsoft/Amazon/Apple/Oracle) → "Jun 21"
//   - Empty / unparseable                     → "Unknown"
func displayDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "Unknown"
	}

	// Workday relative: strip 7-char "Posted " prefix, show the rest verbatim.
	if strings.HasPrefix(strings.ToLower(raw), "posted ") {
		return raw[7:]
	}

	// Absolute date: parse and reformat as short "Mon DD".
	for _, format := range dateFormats {
		if t, err := time.Parse(format, raw); err == nil {
			return t.Format("Jan 02")
		}
	}

	return "Unknown"
}

// extractRowDate recovers a sortable time.Time from a raw markdown table row.
// Looks for an embedded <!-- posted:... --> comment first (written by new row
// construction). Falls back to the display date column for rows written before
// comment embedding was introduced.
func extractRowDate(row string) (time.Time, bool) {
	const prefix = "<!-- posted:"
	const suffix = " -->"
	if idx := strings.Index(row, prefix); idx != -1 {
		start := idx + len(prefix)
		if end := strings.Index(row[start:], suffix); end != -1 {
			raw := strings.TrimSpace(row[start : start+end])
			if t, ok := parsePostingDate(raw); ok {
				return t, true
			}
		}
	}

	// Fallback: parse the display date from the last non-empty, non-comment column.
	cols := strings.Split(row, "|")
	for i := len(cols) - 1; i >= 0; i-- {
		col := strings.TrimSpace(cols[i])
		if col == "" || strings.HasPrefix(col, "<!--") {
			continue
		}
		// Handles Workday relative strings ("Today", "2 Days Ago") directly.
		if t, ok := parsePostingDate(col); ok {
			return t, true
		}
		// "Mon DD" display format needs year injection.
		for _, format := range []string{"Jan 02", "Jan 2"} {
			if t, err := time.Parse(format, col); err == nil {
				now := time.Now()
				return time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), true
			}
		}
		break
	}
	return time.Time{}, false
}
