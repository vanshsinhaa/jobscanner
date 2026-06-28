package process

import (
	"fmt"
	"sync"

	"github.com/vanshsinhaa/jobscanner/common"
	"github.com/vanshsinhaa/jobscanner/database"
	sitesmain "github.com/vanshsinhaa/jobscanner/sites_main"
	"github.com/vanshsinhaa/jobscanner/workday"
	workdaymain "github.com/vanshsinhaa/jobscanner/workday_main"
)

var (
	cachedJobs  []common.JobPosting
	cachedError error
	onceGetJobs sync.Once
)

// ScrapeAllJobs hits every scraper, deduplicates against job_ids.json, and returns new jobs.
// No caching — safe to call each watch-mode iteration without sync.Once interference.
// Individual scraper failures are non-fatal: they are logged and the run continues.
// Only deduplication errors are returned.
func ScrapeAllJobs() ([]common.JobPosting, error) {
	var allJobs []common.JobPosting

	jobs, err := GetAllWorkdayJobs()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("All Workday Jobs: ", len(jobs))
	allJobs = append(allJobs, jobs...)

	for company := range common.SitesCompanies {
		jobs, err := sitesmain.FetchJobsByCompany(company)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("All %s Jobs: %d\n", company, len(jobs))
		allJobs = append(allJobs, jobs...)
	}

	result, err := processDublicateJobs(allJobs)
	if err != nil {
		fmt.Println(err.Error())
	}
	return result, err
}

// GetProcessedNewJobs caches the scrape result for single-run mode (CI).
// The sync.Once ensures scrapers run exactly once per process lifetime.
// In watch/daemon mode, call ScrapeAllJobs() directly each iteration to bypass the cache.
func GetProcessedNewJobs() ([]common.JobPosting, error) {
	onceGetJobs.Do(func() {
		cachedJobs, cachedError = ScrapeAllJobs()
	})
	return cachedJobs, cachedError
}

// ProcessJobsWithDB scrapes, deduplicates, inserts into SQLite, and returns the new jobs.
// Used in single-run (CI) mode. Returns the new jobs slice so main can log counts and
// pass to Discord notify (Phase 8).
func ProcessJobsWithDB() ([]common.JobPosting, error) {
	jobs, err := GetProcessedNewJobs()
	if err != nil {
		fmt.Println("error while processing new jobs: ", err.Error())
		return nil, err
	}
	fmt.Println("Processed Jobs (New Jobs): ", len(jobs))

	if err = database.InsertIntoDB(jobs); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return jobs, nil
}

func ProcessJobsWithDBForNewlyAddedJobPortal(company string, w bool) error {
	workday.Init()
	jobs, err := getProcessedNewJobsNewlyAddedJobPortal(company, w)
	if err != nil {
		fmt.Println("error while processing new jobs: ", err.Error())
		return err
	}
	fmt.Println("Processed Jobs (New Jobs): ", len(jobs))

	err = database.InsertIntoDB(jobs)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

func getProcessedNewJobsNewlyAddedJobPortal(company string, w bool) ([]common.JobPosting, error) {
	onceGetJobs.Do(func() {
		var allJobs []common.JobPosting

		if w {
			jobs, err := workdaymain.GetWorkdayJobs(workdaymain.WorkdayPayloads[company])
			if err != nil {
				fmt.Println(err.Error())
				cachedError = err
				return
			}
			fmt.Println(jobs[0])
			fmt.Println("All "+company+" Jobs: ", len(jobs))
			allJobs = append(allJobs, jobs...)

		} else {
			jobs, err := sitesmain.FetchJobsByCompany(company)
			if err != nil {
				fmt.Println(err.Error())
				cachedError = err
				return
			}
			fmt.Println("All "+company+" Jobs: ", len(jobs))
			allJobs = append(allJobs, jobs...)
		}

		fmt.Println(allJobs[0])
		cachedJobs, cachedError = processDublicateJobs(allJobs)
		if cachedError != nil {
			fmt.Println(cachedError.Error())
		}
	})

	return cachedJobs, cachedError
}
