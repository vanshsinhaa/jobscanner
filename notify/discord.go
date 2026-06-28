package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vanshsinhaa/jobscanner/common"
	"github.com/vanshsinhaa/jobscanner/database"
)

// SendCISummary posts one batch summary embed per CI run.
// Silent when webhookURL is empty or no new jobs were found.
// ≤5 new jobs: per-job embeds with clickable apply links.
// >5 new jobs: summary embed + compact list embeds by category.
func SendCISummary(newJobs []common.JobPosting, webhookURL string) error {
	if webhookURL == "" {
		return nil
	}
	if len(newJobs) == 0 {
		return post(webhookURL, map[string]any{
			"embeds": []map[string]any{{
				"title":       "seems dry in here...",
				"description": "No new jobs found this run.",
				"color":       0x95A5A6,
				"footer":      map[string]any{"text": fmt.Sprintf("vanshsinhaa/jobs · %s UTC", time.Now().UTC().Format("Jan 02 15:04"))},
			}},
		})
	}

	var internJobs, generalJobs []common.JobPosting
	for _, j := range newJobs {
		if rt := database.ClassifyRole(j.JobTitle); rt == "intern" || rt == "new_grad" {
			internJobs = append(internJobs, j)
		} else {
			generalJobs = append(generalJobs, j)
		}
	}

	// Few jobs: show each one with a clickable apply link.
	if len(newJobs) <= 5 {
		var embeds []map[string]any
		for _, j := range newJobs {
			embeds = append(embeds, buildJobEmbed(j))
		}
		return post(webhookURL, map[string]any{"embeds": embeds})
	}

	// Many jobs: summary counts + compact list per category.
	embeds := []map[string]any{buildSummaryEmbed(len(newJobs), len(internJobs), len(generalJobs))}
	if len(internJobs) > 0 {
		embeds = append(embeds, buildInternListEmbed(internJobs))
	}
	if len(generalJobs) > 0 {
		embeds = append(embeds, buildGeneralListEmbed(generalJobs))
	}
	return post(webhookURL, map[string]any{"embeds": embeds})
}

// SendWatchAlert posts per-job embeds for intern/new-grad roles found in a watch-mode sweep.
// General-only sweeps are silent. Batches into one embed if >10 intern jobs (flood guard).
func SendWatchAlert(newJobs []common.JobPosting, webhookURL string) error {
	if webhookURL == "" || len(newJobs) == 0 {
		return nil
	}

	var internJobs []common.JobPosting
	for _, j := range newJobs {
		if rt := database.ClassifyRole(j.JobTitle); rt == "intern" || rt == "new_grad" {
			internJobs = append(internJobs, j)
		}
	}
	if len(internJobs) == 0 {
		return nil
	}

	// Flood guard: first setup run can surface hundreds of intern jobs at once.
	if len(internJobs) > 10 {
		return post(webhookURL, map[string]any{"embeds": []map[string]any{buildInternListEmbed(internJobs)}})
	}

	for _, j := range internJobs {
		if err := post(webhookURL, map[string]any{"embeds": []map[string]any{buildJobEmbed(j)}}); err != nil {
			return err
		}
	}
	return nil
}

func buildSummaryEmbed(total, internCount, generalCount int) map[string]any {
	return map[string]any{
		"title": fmt.Sprintf("%d new jobs added", total),
		"color": 0x57F287,
		"fields": []map[string]any{
			{"name": "Intern / New Grad", "value": fmt.Sprintf("%d", internCount), "inline": true},
			{"name": "General SWE", "value": fmt.Sprintf("%d", generalCount), "inline": true},
		},
		"footer": map[string]any{
			"text": fmt.Sprintf("vanshsinhaa/jobs · %s UTC", time.Now().UTC().Format("Jan 02 15:04")),
		},
	}
}

func buildInternListEmbed(jobs []common.JobPosting) map[string]any {
	const max = 10
	var lines []string
	for i, j := range jobs {
		if i >= max {
			lines = append(lines, fmt.Sprintf("*+%d more*", len(jobs)-max))
			break
		}
		lines = append(lines, fmt.Sprintf("**%s** — %s", j.Company, j.JobTitle))
	}
	return map[string]any{
		"title":       "New intern / new-grad openings",
		"color":       0x5865F2,
		"description": strings.Join(lines, "\n"),
	}
}

func buildGeneralListEmbed(jobs []common.JobPosting) map[string]any {
	const max = 10
	var lines []string
	for i, j := range jobs {
		if i >= max {
			lines = append(lines, fmt.Sprintf("*+%d more*", len(jobs)-max))
			break
		}
		lines = append(lines, fmt.Sprintf("**%s** — %s", j.Company, j.JobTitle))
	}
	return map[string]any{
		"title":       "New general SWE openings",
		"color":       0xE67E22,
		"description": strings.Join(lines, "\n"),
	}
}

func buildJobEmbed(j common.JobPosting) map[string]any {
	return map[string]any{
		"title": fmt.Sprintf("%s — %s", j.Company, j.JobTitle),
		"color": 0x5865F2,
		"url":   j.ExternalPath,
		"fields": []map[string]any{
			{"name": "Location", "value": j.Location, "inline": true},
			{"name": "Posted", "value": j.PostedOn, "inline": true},
		},
	}
}

func post(webhookURL string, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal failed: %w", err)
	}
	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("discord: post failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord: returned %d", resp.StatusCode)
	}
	return nil
}
