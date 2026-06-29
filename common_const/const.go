package commonconst

import (
	"os"
	"path/filepath"
)

// DataDir returns the root directory for all local state files.
// Set DATA_DIR env var in CI to point at the user repo checkout.
// Defaults to "local_data" for local runs.
func DataDir() string {
	if d := os.Getenv("DATA_DIR"); d != "" {
		return d
	}
	return "local_data"
}

// ReadmePath returns the path to README.md.
// Set README_PATH env var in CI to point at the user repo checkout.
func ReadmePath() string {
	if p := os.Getenv("README_PATH"); p != "" {
		return p
	}
	return "README.md"
}

// TargetReadmePath is where the target companies section is written.
// In CI, set TARGET_README_PATH to the dev repo README so the personal feed
// goes there instead of the public jobs repo.
// Locally defaults to README.md (the dev repo README in the working directory).
func TargetReadmePath() string {
	if p := os.Getenv("TARGET_README_PATH"); p != "" {
		return p
	}
	return "README.md"
}

func JobIdFile() string                   { return filepath.Join(DataDir(), "job_ids.json") }
func SnowflakeHiringManagersFile() string { return filepath.Join(DataDir(), "snowflake.json") }
func DBPath() string                      { return filepath.Join(DataDir(), "jobs.db") }
func JSONExportPath() string              { return filepath.Join(DataDir(), "jobs.json") }
// TargetCompaniesFile returns the path to target_companies.json.
// In CI, set TARGET_COMPANIES_FILE to point at the scraper repo's local_data so
// the personal config is read from the dev repo, not the public jobs repo.
func TargetCompaniesFile() string {
	if f := os.Getenv("TARGET_COMPANIES_FILE"); f != "" {
		return f
	}
	return filepath.Join(DataDir(), "target_companies.json")
}
