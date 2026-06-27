package database

import "strings"

// ClassifyRole returns "intern", "new_grad", or "general" based on job title keywords.
// Called at insert time to tag each row in the DB, and again in the README writer
// to partition intern/new-grad rows to the top of the table.
func ClassifyRole(title string) string {
	t := strings.ToLower(title)

	for _, kw := range []string{"intern", "internship", "co-op", "coop", "co op"} {
		if strings.Contains(t, kw) {
			return "intern"
		}
	}

	// Compound phrases only — "associate" and "junior" alone produce too many false positives
	// (e.g., "Associate Product Manager" at senior level, "Junior" meaning 5+ YOE in some regions).
	for _, kw := range []string{
		"new grad", "new graduate",
		"entry level", "entry-level",
		"university grad", "university graduate",
		"campus hire",
		"associate software engineer", "associate engineer",
	} {
		if strings.Contains(t, kw) {
			return "new_grad"
		}
	}

	return "general"
}
