package database

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	commonconst "github.com/vanshsinhaa/jobscanner/common_const"
)

type JobExport struct {
	Company  string `json:"company"`
	Title    string `json:"title"`
	Location string `json:"location"`
	URL      string `json:"url"`
	PostedOn string `json:"posted_on"`
	RoleType string `json:"role_type"`
}

type ExportPayload struct {
	GeneratedAt  time.Time   `json:"generated_at"`
	Total        int         `json:"total"`
	InternCount  int         `json:"intern_count"`
	GeneralCount int         `json:"general_count"`
	Jobs         []JobExport `json:"jobs"`
}

// ExportJSON writes all jobs from SQLite to jobs.json at JSONExportPath().
// Called after each scrape run. Non-fatal to the caller — it only writes a data file.
func ExportJSON() error {
	db := GetDB()
	rows, err := db.Query(`SELECT company, title, location, external_url, posted_on, role_type
		FROM jobs ORDER BY inserted_on DESC`)
	if err != nil {
		return fmt.Errorf("export query failed: %w", err)
	}
	defer rows.Close()

	var jobs []JobExport
	var internCount, generalCount int
	for rows.Next() {
		var j JobExport
		var postedOn *string
		if err := rows.Scan(&j.Company, &j.Title, &j.Location, &j.URL, &postedOn, &j.RoleType); err != nil {
			continue
		}
		if postedOn != nil {
			j.PostedOn = *postedOn
		}
		if j.RoleType == "intern" || j.RoleType == "new_grad" {
			internCount++
		} else {
			generalCount++
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("export row scan failed: %w", err)
	}

	payload := ExportPayload{
		GeneratedAt:  time.Now().UTC(),
		Total:        len(jobs),
		InternCount:  internCount,
		GeneralCount: generalCount,
		Jobs:         jobs,
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal failed: %w", err)
	}
	return os.WriteFile(commonconst.JSONExportPath(), data, 0644)
}
