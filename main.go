package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vanshsinhaa/jobscanner/database"
	"github.com/vanshsinhaa/jobscanner/notify"
	"github.com/vanshsinhaa/jobscanner/process"
	"github.com/vanshsinhaa/jobscanner/readme"
)

var (
	watchMode    bool
	interval     time.Duration
	targetReport bool
)

func init() {
	flag.BoolVar(&watchMode, "watch", false, "run in daemon mode, polling on interval")
	flag.DurationVar(&interval, "interval", 15*time.Minute, "polling interval for watch mode")
	flag.BoolVar(&targetReport, "target-report", false, "print target company coverage for the last 7 days and exit")
}

func main() {
	flag.Parse()
	if targetReport {
		runTargetReport()
		return
	}
	if watchMode {
		runWatchMode(interval)
		return
	}
	runOnce()
}

// runOnce is the standard single-run mode used by CI and local one-shot runs.
// It scrapes, inserts, updates the README, and exports jobs.json.
func runOnce() {
	newJobs, err := process.ProcessJobsWithDB()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("New jobs this run: %d\n", len(newJobs))

	if err := readme.ReadMeProcessNewJobs(); err != nil {
		fmt.Println("error while processing readme file with new jobs:", err)
	}

	if err := database.ExportJSON(); err != nil {
		fmt.Println("warn: json export failed:", err)
	}
	if err := readme.WriteTargetCompanySection(); err != nil {
		fmt.Println("warn: target section failed:", err)
	}
	if err := notify.SendCISummary(newJobs, os.Getenv("DISCORD_WEBHOOK_URL")); err != nil {
		fmt.Println("warn: discord notify failed:", err)
	}
}

// runWatchMode is the local daemon mode. Each iteration calls ScrapeAllJobs() directly,
// bypassing the sync.Once cache so fresh jobs are found every poll.
func runWatchMode(d time.Duration) {
	fmt.Printf("Watch mode: polling every %s\n", d)
	for {
		fmt.Printf("\n[%s] Running scrape...\n", time.Now().Format("15:04:05"))

		jobs, err := process.ScrapeAllJobs()
		if err != nil {
			fmt.Println("scrape error:", err)
			time.Sleep(d)
			continue
		}

		fmt.Printf("New jobs this sweep: %d\n", len(jobs))

		// ScrapeAllJobs already inserted all scraped jobs into the DB before dedup.
		// Use ReadMeProcessNewJobs() (DB-backed) so the README gets the full current
		// picture, not only the dedup-filtered new-only slice.
		if err := readme.ReadMeProcessNewJobs(); err != nil {
			fmt.Println("readme error:", err)
		}
		if err := database.ExportJSON(); err != nil {
			fmt.Println("warn: json export failed:", err)
		}
		if err := readme.WriteTargetCompanySection(); err != nil {
			fmt.Println("warn: target section failed:", err)
		}
		if err := notify.SendWatchAlert(jobs, os.Getenv("DISCORD_WEBHOOK_URL")); err != nil {
			fmt.Println("warn: discord notify failed:", err)
		}

		time.Sleep(d)
	}
}

func runTargetReport() {
	targets, err := database.LoadTargetCompanies()
	if err != nil {
		fmt.Println("error loading target companies:", err)
		return
	}
	if len(targets) == 0 {
		fmt.Println("No target companies configured. Create local_data/target_companies.json with a JSON array of company names.")
		return
	}
	if err := database.SyncTargetCompanies(targets); err != nil {
		fmt.Println("error syncing target companies:", err)
		return
	}
	report, err := database.TargetCompanyReport()
	if err != nil {
		fmt.Println("error running target report:", err)
		return
	}
	fmt.Printf("\n%-30s  %9s  %s\n", "Company", "Jobs (7d)", "Last Seen")
	fmt.Println(strings.Repeat("-", 62))
	for _, r := range report {
		lastSeen := "never"
		if r.LastSeen != "" {
			lastSeen = r.LastSeen
		}
		fmt.Printf("%-30s  %9d  %s\n", r.Name, r.JobsFound, lastSeen)
	}
}
