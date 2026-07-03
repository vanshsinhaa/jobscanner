package database

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// internRe uses word boundaries to match intern/co-op keywords as complete words.
// This prevents false positives: "International"/"Internal" don't match because the
// next char after "intern" is a word char and breaks the boundary.
// Plural forms (Interns, Internships, Co-ops) are explicitly covered — \binternship\b
// does NOT match "Internships" because "p" is followed by "s" (word char, no boundary).
var internRe = regexp.MustCompile(`(?i)\b(intern(?:ship)?s?|co-?ops?|co ops?)\b`)

// newGradYearRe catches cycle-year phrasings that the keyword list can't:
// "Class of 2027", "2027 Grads", "New Grads 2027", "Graduates 2027".
// Year-adjacent grad wording is unambiguous — a senior role never says "2027 Grads".
var newGradYearRe = regexp.MustCompile(`(?i)\b(class of 20\d{2}|20\d{2}\s+grad(?:uate)?s?|grad(?:uate)?s?\s+20\d{2})\b`)

// cycleYearRe extracts recruiting-cycle years (2020–2035) from titles, e.g.
// "Software Engineer Intern (Summer 2027)" or "2027 SDE New Grad".
var cycleYearRe = regexp.MustCompile(`\b20(?:2\d|3[0-5])\b`)

// ClassifyRole returns "intern", "new_grad", or "general" based on job title keywords.
// Called at insert time (database/insert_data.go) and at display time (readme/process_readme.go).
func ClassifyRole(title string) string {
	if internRe.MatchString(title) {
		return "intern"
	}

	if newGradYearRe.MatchString(title) {
		return "new_grad"
	}

	t := strings.ToLower(title)

	// Compound phrases only — "associate" and "junior" alone produce too many false positives
	// ("Associate Product Manager" at senior level, "Junior" = 5+ YOE in some regions).
	for _, kw := range []string{
		"new grad", "new graduate",
		"entry level", "entry-level",
		"university grad", "university graduate",
		"university hire", "university program",
		"campus hire",
		"graduate software engineer", "graduate engineer",
		"associate software engineer", "associate engineer",
	} {
		if strings.Contains(t, kw) {
			return "new_grad"
		}
	}

	return "general"
}

// CycleYear returns the recruiting-cycle year named in a title ("Summer 2027 Intern"
// -> 2027), or 0 if the title names no year. Ranges like "2026/2027" return the
// later year — the posting is live for the newer cycle.
func CycleYear(title string) int {
	max := 0
	for _, m := range cycleYearRe.FindAllString(title, -1) {
		if y, err := strconv.Atoi(m); err == nil && y > max {
			max = y
		}
	}
	return max
}

// IsStaleCycle reports whether a title names a recruiting cycle that has already
// passed (e.g. a "Summer 2026 Intern" posting still up in 2027). Year == current
// year is never stale: Fall/Winter cohorts of the current year are still active.
// Titles with no year always pass. Used at display time only — stale postings stay
// in the DB and job_ids.json so they don't reappear if the company reposts them.
func IsStaleCycle(title string) bool {
	y := CycleYear(title)
	return y > 0 && y < time.Now().Year()
}
