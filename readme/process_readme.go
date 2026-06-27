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

// maxTableRows caps the total number of job rows in the README so that it
// stays under GitHub's 512KB markdown rendering limit (~1700 rows).
const maxTableRows = 1500

// maxJobAgeDays is the maximum age of a posting to include in the README.
// Only applies to jobs with a parseable PostedOn. Jobs with "Unknown" dates
// are always included — we cannot know their age.
const maxJobAgeDays = 14

// allowedMonths defines which months to keep in the README.
// Since the README only stores "Mon DD" without a year, we can't distinguish
// years reliably. Set this to the months you want to retain.
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

// isRowInAllowedMonth checks if a table row's date (last column, "Mon DD" format)
// belongs to one of the allowed months.
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

	// Always keep rows whose date is "Unknown" — job_ids.json already marks them
	// as seen so they can never be re-added if pruned here.
	if dateStr == "Unknown" {
		return true
	}

	if len(dateStr) < 3 {
		return false
	}

	month := dateStr[:3]
	return allowedMonths[month]
}

// extractLink extracts the href URL from a table row to use as a unique key.
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

func appendJobsToReadme(jobPostings []common.JobPosting) error {
	file, err := os.ReadFile("README.md")
	if err != nil {
		return fmt.Errorf("error reading README.md: %v", err)
	}

	content := string(file)

	tableMarker := "| --- | --- | --- | :---: | :---: |"
	splitContent := strings.Split(content, tableMarker)

	if len(splitContent) < 2 {
		return fmt.Errorf("table marker not found")
	}

	// Sort by posting date, newest first.
	// Jobs with no parseable PostedOn (Workday "Unknown", etc.) fall to the end.
	sort.Slice(jobPostings, func(i, j int) bool {
		ti, oki := parsePostingDate(jobPostings[i].PostedOn)
		tj, okj := parsePostingDate(jobPostings[j].PostedOn)
		if !oki && !okj {
			return false
		}
		if !oki {
			return false // no date → end of list
		}
		if !okj {
			return true
		}
		return ti.After(tj)
	})

	// Drop jobs older than maxJobAgeDays. "Posted 30+ Days Ago" (parsed as 30)
	// is the primary casualty. Jobs with no parseable date are always kept.
	cutoff := time.Now().AddDate(0, 0, -maxJobAgeDays)
	var filteredJobs []common.JobPosting
	for _, job := range jobPostings {
		t, ok := parsePostingDate(job.PostedOn)
		if ok && t.Before(cutoff) {
			continue
		}
		filteredJobs = append(filteredJobs, job)
	}
	jobPostings = filteredJobs

	// Stable partition: intern/new-grad rows first (date-sorted within group),
	// general roles after. The date sort ran above, so order within each group is preserved.
	var priorityJobs, generalJobs []common.JobPosting
	for _, job := range jobPostings {
		if rt := database.ClassifyRole(job.JobTitle); rt == "intern" || rt == "new_grad" {
			priorityJobs = append(priorityJobs, job)
		} else {
			generalJobs = append(generalJobs, job)
		}
	}
	jobPostings = append(priorityJobs, generalJobs...)

	seen := make(map[string]bool)

	// Build new rows using the actual PostedOn value from each scraper.
	var newRows []string
	for _, job := range jobPostings {
		row := fmt.Sprintf("| **%s** | %s | %s | <a href=\"%s\" target=\"_blank\"><img src=\"https://i.imgur.com/u1KNU8z.png\" width=\"118\" alt=\"Apply\"></a> | %s |",
			job.Company, job.JobTitle, job.Location, job.ExternalPath, displayDate(job.PostedOn))
		link := extractLink(row)
		if link != "" && !seen[link] {
			seen[link] = true
			newRows = append(newRows, row)
		}
	}

	// Filter existing rows: keep allowed months + deduplicate by link
	existingLines := strings.Split(splitContent[1], "\n")
	var keptRows []string
	var footerLines []string
	for _, line := range existingLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || !strings.HasPrefix(trimmed, "|") {
			footerLines = append(footerLines, line)
			continue
		}
		if !isRowInAllowedMonth(trimmed) {
			continue
		}
		link := extractLink(trimmed)
		if link != "" && seen[link] {
			continue // duplicate
		}
		if link != "" {
			seen[link] = true
		}
		keptRows = append(keptRows, line)
	}

	if len(newRows) == 0 && len(keptRows) == 0 {
		fmt.Println("no new data!!!")
		return nil
	}

	// Merge new + existing rows, capped at maxTableRows (newest first)
	allRows := append(newRows, keptRows...)
	if len(allRows) > maxTableRows {
		allRows = allRows[:maxTableRows]
	}

	var sb strings.Builder
	sb.WriteString(splitContent[0])
	sb.WriteString(tableMarker)
	sb.WriteString("\n")
	for _, row := range allRows {
		sb.WriteString(row)
		sb.WriteString("\n")
	}
	sb.WriteString(strings.Join(footerLines, "\n"))

	err = os.WriteFile("README.md", []byte(sb.String()), 0644)
	if err != nil {
		return fmt.Errorf("error writing to README.md: %v", err)
	}

	fmt.Println("Job postings appended successfully!")
	return nil
}
