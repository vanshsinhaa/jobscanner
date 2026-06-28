package readme

import (
	"fmt"
	"os"
	"strings"

	commonconst "github.com/vanshsinhaa/jobscanner/common_const"
	"github.com/vanshsinhaa/jobscanner/database"
)

const (
	targetSectionHeader = "## \U0001f3af My Target Companies"
	targetEndAnchor     = "<!-- target-rows-end -->"
)

// WriteTargetCompanySection queries the DB for all jobs at configured target companies
// and writes (or creates) the target section in the README. No-op if target_companies.json
// is missing or empty.
func WriteTargetCompanySection() error {
	targets, err := database.LoadTargetCompanies()
	if err != nil {
		return fmt.Errorf("target section: %w", err)
	}
	if len(targets) == 0 {
		return nil
	}

	if err := database.SyncTargetCompanies(targets); err != nil {
		return fmt.Errorf("target section: %w", err)
	}

	jobs, err := database.GetTargetCompanyJobs(targets)
	if err != nil {
		return fmt.Errorf("target section: %w", err)
	}

	var rows []string
	seen := make(map[string]bool)
	for _, j := range jobs {
		row := fmt.Sprintf("| **%s** | %s | %s | <a href=\"%s\" target=\"_blank\"><img src=\"https://i.imgur.com/u1KNU8z.png\" width=\"118\" alt=\"Apply\"></a> | %s |",
			j.Company, j.JobTitle, j.Location, j.ExternalPath, displayDate(j.PostedOn))
		link := extractLink(row)
		if link != "" && seen[link] {
			continue
		}
		if link != "" {
			seen[link] = true
		}
		rows = append(rows, row)
	}

	data, err := os.ReadFile(commonconst.ReadmePath())
	if err != nil {
		return fmt.Errorf("target section: read README: %w", err)
	}
	content := string(data)

	if !strings.Contains(content, targetEndAnchor) {
		content = insertTargetSection(content)
	}

	content, err = replaceSection(content, targetSectionHeader, targetEndAnchor, rows)
	if err != nil {
		return fmt.Errorf("target section: %w", err)
	}

	if err := os.WriteFile(commonconst.ReadmePath(), []byte(content), 0644); err != nil {
		return fmt.Errorf("target section: write README: %w", err)
	}

	fmt.Printf("Target company jobs written: %d\n", len(rows))
	return nil
}

// insertTargetSection prepends the target section skeleton before the intern section header.
// Called once to bootstrap the README when the section doesn't exist yet.
func insertTargetSection(content string) string {
	skeleton := targetSectionHeader + "\n\n" +
		"| Company | Role | Location | Apply | Posted |\n" +
		tableSep + "\n" +
		targetEndAnchor + "\n\n"

	idx := strings.Index(content, internSectionHeader)
	if idx == -1 {
		// No intern section found; prepend at top.
		return skeleton + content
	}
	return content[:idx] + skeleton + content[idx:]
}
