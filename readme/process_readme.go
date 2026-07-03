package readme

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/vanshsinhaa/jobscanner/common"
	commonconst "github.com/vanshsinhaa/jobscanner/common_const"
	"github.com/vanshsinhaa/jobscanner/database"
)

const (
	maxTableRows       = 1500
	maxInternTableRows = 500
	maxJobAgeDays      = 14  // general SWE: 2-week window; roles fill fast, stale listings aren't useful
	maxInternJobAgeDays = 60 // intern/new-grad: 60-day window; programs post months before the start date

	// Section headers used to locate each table in the README.
	// The code searches for the table separator row within the bounded region
	// [sectionHeader, endAnchor) and writes rows immediately after the separator.
	// This keeps HTML comments OUT of the table body, which is critical:
	// a comment between the separator row and the first data row breaks GitHub's
	// markdown table parser and causes all rows to render as prose.
	internSectionHeader  = "## \U0001F393 Intern & New Grad Opportunities"
	generalSectionHeader = "## \U0001F4BC All SWE Opportunities"

	// End anchors mark where each table's rows stop.
	// They appear AFTER the last data row, which ends the table cleanly.
	internEndAnchor  = "<!-- intern-rows-end -->"
	generalEndAnchor = "<!-- general-rows-end -->"

	// Standard table separator shared by both tables.
	tableSep = "| --- | --- | --- | :---: | :---: |"
)

// allowedMonths returns current + previous month so isRowInAllowedMonth never
// needs a manual update at month boundaries. Computed fresh per call.
func allowedMonths() map[string]bool {
	now := time.Now()
	return map[string]bool{
		now.Format("Jan"):                   true,
		now.AddDate(0, -1, 0).Format("Jan"): true,
	}
}

// ReadMeProcessNewJobs writes all jobs from the current run's DB to the jobs repo README.
// Sourcing from the DB (not the dedup-filtered new-only list) ensures the jobs repo README
// always reflects the full current scrape: all companies' jobs appear, not just first-seen ones.
// The kept-rows mechanism in appendJobsToReadme still provides resilience for temporarily-
// failing scrapers — rows from prior runs persist until isRowInAllowedMonth ages them out.
func ReadMeProcessNewJobs() error {
	jobs, err := database.GetAllJobs()
	if err != nil {
		return fmt.Errorf("readme: get all jobs from db: %w", err)
	}
	return appendJobsToReadme(jobs)
}

// isRowInAllowedMonth decides whether to keep an existing README row between runs.
// Three date formats can appear in the date column after Phase 2:
//  1. "Mon DD" (ISO-derived)        â†’ check allowedMonths map
//  2. Workday relative ("5 Days Ago", "Today", "Yesterday") â†’ always keep (inherently recent)
//  3. "Unknown"                     â†’ always keep (age unknowable; job is in job_ids.json)
func isRowInAllowedMonth(row string) bool {
	row = strings.TrimSpace(row)
	if row == "" || !strings.HasPrefix(row, "|") {
		return false
	}

	cols := strings.Split(row, "|")
	dateStr := ""
	for i := len(cols) - 1; i >= 0; i-- {
		col := strings.TrimSpace(cols[i])
		if col != "" && !strings.HasPrefix(col, "<!--") {
			dateStr = col
			break
		}
	}

	if dateStr == "Unknown" {
		return true
	}

	lower := strings.ToLower(dateStr)
	if lower == "today" || lower == "yesterday" || strings.Contains(lower, "days ago") {
		return true
	}

	if len(dateStr) < 3 {
		return false
	}
	return allowedMonths()[dateStr[:3]]
}

// extractLink pulls the href URL from a table row for use as a dedup key.
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

// extractTitle gets the job title column from a markdown table row.
// Row format: | **Company** | Job Title | Location | Link | Date |
func extractTitle(row string) string {
	cols := strings.Split(row, "|")
	if len(cols) < 3 {
		return ""
	}
	return strings.TrimSpace(cols[2])
}

// sortRows sorts markdown table rows newest-first using extractRowDate.
// Rows without a parseable date sort to the end. Uses SliceStable so equal-date
// rows keep their relative order (new rows added this run stay ahead of same-day
// kept rows from previous runs).
func sortRows(rows []string) {
	sort.SliceStable(rows, func(i, j int) bool {
		ti, oki := extractRowDate(rows[i])
		tj, okj := extractRowDate(rows[j])
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
}

// replaceSection writes rows into a table identified by its section header.
// It finds the tableSep within [sectionHeader, endAnchor), then replaces
// everything between the separator and the end anchor with the given rows.
// The separator row itself and everything before it are left untouched.
func replaceSection(content, sectionHeader, endAnchor string, rows []string) (string, error) {
	si := strings.Index(content, sectionHeader)
	if si == -1 {
		return "", fmt.Errorf("section header %q not found", sectionHeader)
	}

	// Work within the region starting at si to avoid matching the wrong table.
	rest := content[si:]
	ei := strings.Index(rest, endAnchor)
	if ei == -1 {
		return "", fmt.Errorf("end anchor %q not found after section %q", endAnchor, sectionHeader)
	}

	// Find the separator row inside this section (before end anchor).
	sepi := strings.Index(rest[:ei], tableSep)
	if sepi == -1 {
		return "", fmt.Errorf("table separator not found in section %q", sectionHeader)
	}

	// Write point: the character immediately after the separator row.
	writeAt := si + sepi + len(tableSep)

	var sb strings.Builder
	sb.WriteString(content[:writeAt])
	sb.WriteString("\n")
	for _, row := range rows {
		sb.WriteString(row)
		sb.WriteString("\n")
	}
	// Resume from the end anchor (inclusive) so it is preserved in the output.
	sb.WriteString(rest[ei:])
	return sb.String(), nil
}

// parseExistingRows extracts table rows from between a section's separator row
// and its end anchor. Applies the allowed-month filter and deduplicates against seen.
func parseExistingRows(content, sectionHeader, endAnchor string, seen map[string]bool) []string {
	si := strings.Index(content, sectionHeader)
	if si == -1 {
		return nil
	}
	rest := content[si:]
	ei := strings.Index(rest, endAnchor)
	if ei == -1 {
		return nil
	}
	sepi := strings.Index(rest[:ei], tableSep)
	if sepi == -1 {
		return nil
	}
	section := rest[sepi+len(tableSep) : ei]

	var kept []string
	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		if !isRowInAllowedMonth(trimmed) {
			continue
		}
		// Age out rows whose title names a past recruiting cycle (kept-rows
		// survive scraper outages, so they need the same cycle filter as new rows).
		if database.IsStaleCycle(extractTitle(trimmed)) {
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

// parseAndReclassifyGeneralRows reads existing general table rows and re-classifies
// each one by title. Any row matching intern/new-grad keywords is returned in the
// intern slice so it gets moved to the intern table. This handles the self-healing
// migration from the old single-table format and future keyword updates automatically.
func parseAndReclassifyGeneralRows(content string, seen map[string]bool) (intern, general []string) {
	si := strings.Index(content, generalSectionHeader)
	if si == -1 {
		return nil, nil
	}
	rest := content[si:]
	ei := strings.Index(rest, generalEndAnchor)
	if ei == -1 {
		return nil, nil
	}
	sepi := strings.Index(rest[:ei], tableSep)
	if sepi == -1 {
		return nil, nil
	}
	section := rest[sepi+len(tableSep) : ei]

	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
			continue
		}
		if !isRowInAllowedMonth(trimmed) {
			continue
		}
		if database.IsStaleCycle(extractTitle(trimmed)) {
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

// WriteJobsToReadme writes a given job slice directly to the README.
// Use this in watch mode where the jobs come from process.ScrapeAllJobs()
// rather than the sync.Once-cached process.GetProcessedNewJobs().
func WriteJobsToReadme(jobs []common.JobPosting) error {
	return appendJobsToReadme(jobs)
}

func appendJobsToReadme(jobPostings []common.JobPosting) error {
	file, err := os.ReadFile(commonconst.ReadmePath())
	if err != nil {
		return fmt.Errorf("error reading README: %v", err)
	}
	content := strings.ReplaceAll(string(file), "\xEF\xBB\xBF", "")

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

	// Drop jobs older than their role's cutoff. Jobs without a parseable date always pass.
	// Intern/new-grad roles use a longer window — programs post months before the start date.
	cutoffGeneral := time.Now().AddDate(0, 0, -maxJobAgeDays)
	cutoffIntern := time.Now().AddDate(0, 0, -maxInternJobAgeDays)
	var filtered []common.JobPosting
	for _, job := range jobPostings {
		// Cycle filter: "Summer 2026 Intern" postings left up during the 2027
		// cycle are closed programs — drop them regardless of posting date.
		if database.IsStaleCycle(job.JobTitle) {
			continue
		}
		t, ok := parsePostingDate(job.PostedOn)
		if !ok {
			filtered = append(filtered, job)
			continue
		}
		rt := job.RoleType
		if rt == "" {
			rt = database.ClassifyRole(job.JobTitle)
		}
		cutoff := cutoffGeneral
		if rt == "intern" || rt == "new_grad" {
			cutoff = cutoffIntern
		}
		if t.Before(cutoff) {
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
		row := fmt.Sprintf("| **%s** | %s | %s | <a href=\"%s\" target=\"_blank\"><img src=\"https://i.imgur.com/u1KNU8z.png\" width=\"118\" alt=\"Apply\"></a> | %s |<!-- posted:%s -->",
			job.Company, job.JobTitle, job.Location, job.ExternalPath, displayDate(job.PostedOn), job.PostedOn)
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

	// Existing intern rows are already correctly classified â€” keep as-is.
	keptInternRows := parseExistingRows(content, internSectionHeader, internEndAnchor, seen)

	// Re-classify existing general rows; move any intern/new-grad into the intern table.
	reclassifiedIntern, keptGeneralRows := parseAndReclassifyGeneralRows(content, seen)

	// Merge new + kept for each table, then sort the merged slice so the overall
	// table is newest-first regardless of which run wrote each row.
	internRows := append(append(newInternRows, keptInternRows...), reclassifiedIntern...)
	sortRows(internRows)
	if len(internRows) > maxInternTableRows {
		internRows = internRows[:maxInternTableRows]
	}
	generalRows := append(newGeneralRows, keptGeneralRows...)
	sortRows(generalRows)
	if len(generalRows) > maxTableRows {
		generalRows = generalRows[:maxTableRows]
	}

	if len(internRows) == 0 && len(generalRows) == 0 {
		fmt.Println("no new data!!!")
		return nil
	}

	content, err = replaceSection(content, internSectionHeader, internEndAnchor, internRows)
	if err != nil {
		return fmt.Errorf("intern table: %w", err)
	}
	// Re-search anchors in the updated content string before writing general table.
	content, err = replaceSection(content, generalSectionHeader, generalEndAnchor, generalRows)
	if err != nil {
		return fmt.Errorf("general table: %w", err)
	}

	if err = os.WriteFile(commonconst.ReadmePath(), []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing README: %v", err)
	}

	fmt.Printf("Job postings written: %d intern/new-grad, %d general\n", len(internRows), len(generalRows))
	return nil
}
