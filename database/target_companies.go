package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	commonconst "github.com/vanshsinhaa/jobscanner/common_const"
	"github.com/vanshsinhaa/jobscanner/common"
)

type TargetCompanyStatus struct {
	Name      string
	JobsFound int
	LastSeen  string
}

// targetAliases maps a target_companies.json name (lowercased) to the additional
// scraped company names (lowercased) it should match. Needed where the brand the
// user tracks differs from the name the scraper emits:
//   - Amex postings carry "American Express"
//   - Snap postings carry "Snapchat" (Workday tenant name)
//   - Square postings come from Block's Greenhouse board (scraper already emits "Square")
//   - Twitter no longer exists: X Corp merged into xAI, scraper emits "xAI"
//   - Trello is an Atlassian product; scraper emits "Atlassian"
//   - Slack/Annapurna Labs are retagged sub-brands (process.retagSubBrands) and match directly
var targetAliases = map[string][]string{
	"amex":    {"american express"},
	"snap":    {"snapchat"},
	"twitter": {"xai"},
	"trello":  {"atlassian"},
}

// expandTargetNames lowercases the configured targets and appends alias names.
func expandTargetNames(targets []string) []string {
	seen := make(map[string]bool)
	var names []string
	add := func(n string) {
		n = strings.ToLower(n)
		if !seen[n] {
			seen[n] = true
			names = append(names, n)
		}
	}
	for _, t := range targets {
		add(t)
		for _, alias := range targetAliases[strings.ToLower(t)] {
			add(alias)
		}
	}
	return names
}

// LoadTargetCompanies reads target_companies.json. Returns empty slice if file doesn't exist.
func LoadTargetCompanies() ([]string, error) {
	data, err := os.ReadFile(commonconst.TargetCompaniesFile())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read target_companies.json: %w", err)
	}
	var names []string
	if err := json.Unmarshal(data, &names); err != nil {
		return nil, fmt.Errorf("parse target_companies.json: %w", err)
	}
	return names, nil
}

// SyncTargetCompanies upserts the configured names into the target_companies table.
func SyncTargetCompanies(names []string) error {
	db := GetDB()
	// Clear stale entries so removed companies don't linger.
	if _, err := db.Exec(`DELETE FROM target_companies`); err != nil {
		return fmt.Errorf("clear target_companies: %w", err)
	}
	stmt, err := db.Prepare(`INSERT OR IGNORE INTO target_companies (name) VALUES (?)`)
	if err != nil {
		return fmt.Errorf("prepare target_companies insert: %w", err)
	}
	defer stmt.Close()
	for _, name := range names {
		if _, err := stmt.Exec(name); err != nil {
			return fmt.Errorf("insert target company %q: %w", name, err)
		}
	}
	return nil
}

// GetTargetCompanyJobs returns intern/new_grad jobs in the DB for the given target companies.
// The SQL pre-filters by role_type, then Go post-filters with ClassifyRole(title) to remove
// FTE roles that were bulk-tagged 'intern' by broad text-search query context (e.g. Apple's
// "intern" full-text search matches FTE job descriptions that mention intern programs).
// new_grad DB entries are always kept — university hires often have generic titles that
// ClassifyRole cannot identify without the query-context signal.
func GetTargetCompanyJobs(targets []string) ([]common.JobPosting, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	db := GetDB()

	names := expandTargetNames(targets)
	placeholders := make([]string, len(names))
	args := make([]any, len(names))
	for i, n := range names {
		placeholders[i] = "?"
		args[i] = n
	}
	// intern/new_grad only; intern before new_grad, then newest first.
	// role_type is fetched so we can use it in the post-filter below.
	query := fmt.Sprintf(`
		SELECT company, title, location, external_url, COALESCE(posted_on, ''), role_type
		FROM jobs
		WHERE LOWER(company) IN (%s)
		  AND role_type IN ('intern', 'new_grad')
		ORDER BY CASE role_type WHEN 'intern' THEN 0 ELSE 1 END ASC,
		         inserted_on DESC`,
		strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("target company jobs query: %w", err)
	}
	defer rows.Close()

	var jobs []common.JobPosting
	for rows.Next() {
		var j common.JobPosting
		var dbRoleType string
		if err := rows.Scan(&j.Company, &j.JobTitle, &j.Location, &j.ExternalPath, &j.PostedOn, &dbRoleType); err != nil {
			continue
		}
		// Post-filter: exclude FTE roles that carry role_type='intern' only because
		// they were returned by a broad "intern" text search (their title has no
		// intern/co-op keyword and ClassifyRole classifies them as general).
		// new_grad entries bypass this check — their DB tag is the authoritative signal.
		if dbRoleType == "intern" && ClassifyRole(j.JobTitle) == "general" {
			continue
		}
		j.RoleType = dbRoleType
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

// TargetCompanyReport returns coverage for each configured company over the last 7 days.
// Useful locally where the DB accumulates across runs; in CI the DB is per-run only.
func TargetCompanyReport() ([]TargetCompanyStatus, error) {
	db := GetDB()
	targets, err := LoadTargetCompanies()
	if err != nil {
		return nil, err
	}

	var results []TargetCompanyStatus
	for _, t := range targets {
		names := expandTargetNames([]string{t})
		placeholders := make([]string, len(names))
		args := make([]any, len(names))
		for i, n := range names {
			placeholders[i] = "?"
			args[i] = n
		}
		query := fmt.Sprintf(`
			SELECT COUNT(id), COALESCE(MAX(inserted_on), '')
			FROM jobs
			WHERE LOWER(company) IN (%s)
			  AND inserted_on > datetime('now', '-7 days')`,
			strings.Join(placeholders, ","))

		var s TargetCompanyStatus
		s.Name = t
		if err := db.QueryRow(query, args...).Scan(&s.JobsFound, &s.LastSeen); err != nil {
			return nil, fmt.Errorf("target report query for %q: %w", t, err)
		}
		results = append(results, s)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].JobsFound < results[j].JobsFound })
	return results, nil
}
