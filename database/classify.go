package database

import (
	"regexp"
	"strings"
)

// internRe uses word boundaries to match intern/co-op keywords as complete words.
// This prevents false positives: "International"/"Internal" don't match because the
// next char after "intern" is a word char and breaks the boundary.
// Plural forms (Interns, Internships, Co-ops) are explicitly covered — \binternship\b
// does NOT match "Internships" because "p" is followed by "s" (word char, no boundary).
var internRe = regexp.MustCompile(`(?i)\b(intern(?:ship)?s?|co-?ops?|co ops?)\b`)

// ClassifyRole returns "intern", "new_grad", or "general" based on job title keywords.
// Called at insert time (database/insert_data.go) and at display time (readme/process_readme.go).
func ClassifyRole(title string) string {
	if internRe.MatchString(title) {
		return "intern"
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
		"associate software engineer", "associate engineer",
	} {
		if strings.Contains(t, kw) {
			return "new_grad"
		}
	}

	return "general"
}
