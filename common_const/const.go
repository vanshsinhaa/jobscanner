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

func JobIdFile() string                   { return filepath.Join(DataDir(), "job_ids.json") }
func SnowflakeHiringManagersFile() string { return filepath.Join(DataDir(), "snowflake.json") }
func DBPath() string                      { return filepath.Join(DataDir(), "jobs.db") }
func JSONExportPath() string              { return filepath.Join(DataDir(), "jobs.json") }
