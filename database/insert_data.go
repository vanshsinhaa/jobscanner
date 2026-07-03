package database

import (
	"fmt"

	"github.com/vanshsinhaa/jobscanner/common"
)

// InsertIntoDB inserts job postings into SQLite.
// Duplicate job_id values are silently ignored via INSERT OR IGNORE.
// All rows go in a single transaction: one fsync instead of one per row,
// which turns a ~30-minute insert of ~14k rows into a sub-second commit.
func InsertIntoDB(jobs []common.JobPosting) error {
	db := GetDB()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin insert transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO jobs
		(company, job_id, title, location, posted_on, external_url, role_type)
		VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, j := range jobs {
		roleType := j.RoleType
		if roleType == "" {
			roleType = ClassifyRole(j.JobTitle)
		}

		var postedOn interface{}
		if j.PostedOn != "" {
			postedOn = j.PostedOn
		}

		if _, err := stmt.Exec(
			j.Company,
			j.JobId,
			j.JobTitle,
			j.Location,
			postedOn,
			j.ExternalPath,
			roleType,
		); err != nil {
			fmt.Printf("warn: failed to insert job %s: %v\n", j.JobId, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit insert transaction: %w", err)
	}
	return nil
}

// DeleteJobFromDB removes a single job by its job_id.
func DeleteJobFromDB(jobId string) error {
	db := GetDB()
	res, err := db.Exec(`DELETE FROM jobs WHERE job_id = ?`, jobId)
	if err != nil {
		return fmt.Errorf("failed to delete job %s: %w", jobId, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("no job found with job_id: %s", jobId)
	}
	fmt.Printf("deleted job %s\n", jobId)
	return nil
}
