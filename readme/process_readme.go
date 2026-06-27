package readme

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/neyaadeez/go-get-jobs/common"
	"github.com/neyaadeez/go-get-jobs/database"
	"github.com/neyaadeez/go-get-jobs/process"
)

const (
	// maxTableRows caps the general table so the README stays under GitHub's 512KB render limit.
	maxTableRows = 1500
	// maxInternTableRows caps the intern/new-grad table separately.
	maxInternTableRows = 500
	// maxJobAgeDays filters out stale postings with a parseable date.
	maxJobAgeDays = 14

	// HTML comment anchors delimiting each table's row section in the README.
	// replaceSection writes rows between start (inclusive) and end (exclusive),
	// leaving everything outside the pair untouched.
	internStartAnchor  = "<!-- intern-table-start -->"
	internEndAnchor    = "<!-- intern-table-end -->"
	generalStartAnchor = "<!-- general-table-start -->"
	generalEndAnchor   = "<!-- general-table-end -->"
)

// allowedMonths controls which existing README rows survive between runs.
// Rows with a "Mon DD" date whose month is not listed here are pruned.
// Update this every couple of months (or convert to a rolling window in a future phase).
var allowedMonths = map[string]bool{
	"May": true,
	"Jun": true,
}

func ReadMeProcessNewJobs() error {
	jobs, err := process.GetProcessedNewJobs()
	if err != nil {
		fmt.Println("error while getting new processed jobs: ", err.Error())
	}
	return appendJobsToReadme(jobs)
}

// isRowInAllowedMonth checks whether an existing table row should be kept between runs.
// Three date formats can appear in the date column after Phase 2:
//  1. "Mon DD" (ISO-derived)  → check allowedMonths
//  2. Workday relative ("Today", "Yesterday", "N Days Ago") → always keep (inherently recent)
//  3. "Unknown" → always keep (can't determine age; job is already in job_ids.json)
func isRowInAllowedMonth(row string) bool {
	row = strings.TrimSpace(row)
	if row == "" || !strings.HasPrefix(row, "|") {
		return false
	}

	cols := strings.Split(row, "|")
	dateStr := ""
	for i := len(cols) - 1; i >= 0; i-- {
		col := strings.TrimSpace(cols[i])
		if col != "" {
			dateStr = col
			break
		}
	}

	if dateStr == "Unknown" {
		return true
	}

	// Workday relative strings are inherently recent — always keep them.
	lower := strings.ToLower(dateStr)
	if lower == "today" || lower == "yesterday" || strings.Contains(lower, "days ago") {
		return true
	}

	if len(dateStr) < 3 {
		return false
	}
	return allowedMonths[dateStr[:3]]
}

// extractLink pulls the href URL out of a table row, used as the dedup key.
func extractLink(row string) string {
	idx := strings.Index(row, "href=\"")
	if idx == -1 {
		return ""
	}
	start := idx + len("href=\"")
	end := strings.Index(row[start:], "\"")
	if end == -1 {
		return ""
	}
	return row[start : start+end]
}

// extractTitle pulls the job title column from a markdown table row.
// Row format: | **Company** | Job Title | Location | Link | Date |
// After splitting on "|": ["", " **Company** ", " Job Title ", ...]
func extractTitle(row string) string {
	cols := strings.Split(row, "|")
	if len(cols) < 3 {
		return ""
	}
	return strings.TrimSpace(cols[2])
}

// replaceSection replaces the content between startAnchor and endAnchor with rows.
// Everything outside the anchor pair is left exactly as-is.
func replaceSection(content, startAnchor, endAnchor string, rows []string) (string, error) {
	si := strings.Index(content, startAnchor)
	ei := strings.Index(content, endAnchor)
	if si == -1 {
		return "", fmt.Errorf("start anchor %q not found in README", startAnchor)
	}
	if ei == -1 {
		return "", fmt.Errorf("end anchor %q not found in README", endAnchor)
	}

	var sb strings.Builder
	sb.WriteString(content[:si+len(startAnchor)])
	sb.WriteString("\n")
	for _, row := range rows {
		sb.WriteString(row)
		sb.WriteString("\n")
	}
	sb.WriteString(content[ei:])
	return sb.String(), nil
}

// parseExistingRows reads table rows from between the given anchors.
// Applies the allowed-month filter and deduplicates against seen.
// Used for the intern table, whose rows are already correctly classified.
func parseExistingRows(content, startAnchor, endAnchor string, seen map[string]bool) []string {
	si := strings.Index(content, startAnchor)
	ei := strings.Index(content, endAnchor)
	if si == -1 || ei == -1 || ei <= si {
		return nil
	}
	section := content[si+len(startAnchor) : ei]
	var kept []string
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		if !isRowInAllowedMonth(trimmed) {
			continue
		}
		link := extractLink(trimmed)
		if link != "" && seen[link] {
			continue
		}
		if link != "" {
			seen[link] = true
		}
		kept = append(kept, line)
	}
	return kept
}

// parseAndReclassifyGeneralRows reads existing general table rows and re-classifies each one.
// Any row whose title now matches intern/new-grad keywords is returned in the intern slice
// so it gets moved to the intern table. This handles the one-time migration from the
// old single-table format and also self-corrects as keywords evolve.
func parseAndReclassifyGeneralRows(content string, seen map[string]bool) (intern, general []string) {
	si := strings.Index(content, generalStartAnchor)
	ei := strings.Index(content, generalEndAnchor)
	if si == -1 || ei == -1 || ei <= si {
		return nil, nil
	}
	section := content[si+len(generalStartAnchor) : ei]
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		if !isRowInAllowedMonth(trimmed) {
			continue
		}
		link := extractLink(trimmed)
		if link != "" && seen[link] {
			continue
		}
		if link != "" {
			seen[link] = true
		}
		if rt := database.ClassifyRole(extractTitle(trimmed)); rt == "intern" || rt == "new_grad" {
			intern = append(intern, line)
		} else {
			general = append(general, line)
		}
	}
	return
}

func appendJobsToReadme(jobPostings []common.JobPosting) error {
	file, err := os.ReadFile("README.md")
	if err != nil {
		return fmt.Errorf("error reading README.md: %v", err)
	}
	content := string(file)

	// Sort newest-first; jobs with no parseable date fall to the end.
	sort.Slice(jobPostings, func(i, j int) bool {
		ti, oki := parsePostingDate(jobPostings[i].PostedOn)
		tj, okj := parsePostingDate(jobPostings[j].PostedOn)
		if !oki && !okj {
			return false
		}
		if !oki {
			return false
		}
		if !okj {
			return true
		}
		return ti.After(tj)
	})

	// Drop jobs older than maxJobAgeDays. "Posted 30+ Days Ago" is the primary casualty.
	// Jobs with no parseable date are always kept — we can't know their age.
	cutoff := time.Now().AddDate(0, 0, -maxJobAgeDays)
	var filtered []common.JobPosting
	for _, job := range jobPostings {
		t, ok := parsePostingDate(job.PostedOn)
		if ok && t.Before(cutoff) {
			continue
		}
		filtered = append(filtered, job)
	}
	jobPostings = filtered

	// One global seen map prevents any job from appearing in both tables.
	seen := make(map[string]bool)

	// Classify and build rows for newly scraped jobs.
	var newInternRows, newGeneralRows []string
	for _, job := range jobPostings {
		row := fmt.Sprintf("| **%s** | %s | %s | <a href=\"%s\" target=\"_blank\"><img src=\"https://i.imgur.com/u1KNU8z.png\" width=\"118\" alt=\"Apply\"></a> | %s |",
			job.Company, job.JobTitle, job.Location, job.ExternalPath, displayDate(job.PostedOn))
		link := extractLink(row)
		if link != "" && seen[link] {
			continue
		}
		if link != "" {
			seen[link] = true
		}
		if rt := database.ClassifyRole(job.JobTitle); rt == "intern" || rt == "new_grad" {
			newInternRows = append(newInternRows, row)
		} else {
			newGeneralRows = append(newGeneralRows, row)
		}
	}

	// Existing intern table rows are already correctly classified — keep as-is.
	keptInternRows := parseExistingRows(content, internStartAnchor, internEndAnchor, seen)

	// Re-classify existing general table rows. Any that now match intern/new-grad keywords
	// (e.g., rows written before two-table migration) are moved to the intern table.
	reclassifiedIntern, keptGeneralRows := parseAndReclassifyGeneralRows(content, seen)

	// Merge new + kept for each table, newest-scraped rows first, capped separately.
	internRows := append(append(newInternRows, keptInternRows...), reclassifiedIntern...)
	if len(internRows) > maxInternTableRows {
		internRows = internRows[:maxInternTableRows]
	}
	generalRows := append(newGeneralRows, keptGeneralRows...)
	if len(generalRows) > maxTableRows {
		generalRows = generalRows[:maxTableRows]
	}

	if len(internRows) == 0 && len(generalRows) == 0 {
		fmt.Println("no new data!!!")
		return nil
	}

	// Write intern table rows between their anchors.
	content, err = replaceSection(content, internStartAnchor, internEndAnchor, internRows)
	if err != nil {
		return fmt.Errorf("intern table: %w", err)
	}
	// Write general table rows (re-find anchors in the now-updated content string).
	content, err = replaceSection(content, generalStartAnchor, generalEndAnchor, generalRows)
	if err != nil {
		return fmt.Errorf("general table: %w", err)
	}

	if err = os.WriteFile("README.md", []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing README.md: %v", err)
	}

	fmt.Printf("Job postings written: %d intern/new-grad, %d general\n", len(internRows), len(generalRows))
	return nil
}
