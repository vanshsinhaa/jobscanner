package sites

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"time"

	"github.com/vanshsinhaa/jobscanner/common"
)

// Microsoft careers migrated from gcsservices.careers.microsoft.com (dead, TLS cert mismatch)
// to Eightfold AI at apply.careers.microsoft.com. The /api/pcsx/search endpoint is public —
// no auth, no cookies required. Seniority/employment_type URL params are silently ignored by
// the API; only the text query filter works reliably.
const microsoftAPIBase = "https://apply.careers.microsoft.com/api/pcsx/search"

// maxMicrosoftAgeDays is the scraper-level recency cutoff. Jobs older than this are skipped
// before they ever reach job_ids.json, preventing stale closed-cycle postings from being
// permanently marked seen. The README layer applies its own 14-day display filter on top.
const maxMicrosoftAgeDays = 45

type microsoftPCSXResponse struct {
	Status int `json:"status"`
	Data   struct {
		Positions []microsoftPCSXJob `json:"positions"`
		Count     int                `json:"count"`
	} `json:"data"`
}

type microsoftPCSXJob struct {
	ID           int64    `json:"id"`
	DisplayJobID string   `json:"displayJobId"`
	Name         string   `json:"name"`
	Locations    []string `json:"locations"`
	PostedTs     int64    `json:"postedTs"`
}

func GetMicrosoftJobs() ([]common.JobPosting, error) {
	fmt.Println("Processing: ", "Microsoft")

	cutoff := time.Now().AddDate(0, 0, -maxMicrosoftAgeDays).Unix()

	// Two targeted text searches — the only reliable filter mechanism on this Eightfold
	// instance. "intern" catches SWE/research intern titles; "university" catches
	// Microsoft University Hire / new-grad postings.
	queries := []string{"intern", "university"}

	seen := make(map[string]bool)
	var all []common.JobPosting

	for _, q := range queries {
		jobs, err := microsoftFetchAll(q, cutoff)
		if err != nil {
			fmt.Printf("error fetching Microsoft jobs (query=%q): %v\n", q, err)
			continue
		}
		for _, job := range jobs {
			if !seen[job.JobId] {
				seen[job.JobId] = true
				all = append(all, job)
			}
		}
	}

	return all, nil
}

func microsoftFetchAll(query string, cutoff int64) ([]common.JobPosting, error) {
	jobs, total, err := microsoftPage(query, 1, cutoff)
	if err != nil {
		return nil, err
	}

	pages := int(math.Ceil(float64(total) / 20.0))
	for pg := 2; pg <= pages; pg++ {
		more, _, pageErr := microsoftPage(query, pg, cutoff)
		if pageErr != nil {
			fmt.Printf("warn: Microsoft page %d (query=%q): %v\n", pg, query, pageErr)
			continue
		}
		jobs = append(jobs, more...)
	}

	return jobs, nil
}

func microsoftPage(query string, page int, cutoff int64) ([]common.JobPosting, int, error) {
	client := common.GetClient()

	params := url.Values{}
	params.Set("domain", "microsoft.com")
	params.Set("query", query)
	params.Set("pg", strconv.Itoa(page))
	params.Set("pgSz", "20")

	resp, err := client.R().
		SetHeader("Accept", "application/json").
		Get(microsoftAPIBase + "?" + params.Encode())
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}

	var parsed microsoftPCSXResponse
	if err := json.Unmarshal(resp.Body(), &parsed); err != nil {
		return nil, 0, fmt.Errorf("json parse failed: %w", err)
	}

	var postings []common.JobPosting
	for _, job := range parsed.Data.Positions {
		if job.PostedTs < cutoff {
			continue
		}

		location := "Unknown"
		if len(job.Locations) > 0 {
			location = job.Locations[0]
		}

		postings = append(postings, common.JobPosting{
			JobId:        common.Microsoft + ":" + job.DisplayJobID,
			JobTitle:     job.Name,
			Location:     location,
			PostedOn:     time.Unix(job.PostedTs, 0).UTC().Format("2006-01-02T15:04:05Z"),
			ExternalPath: "https://apply.careers.microsoft.com/careers/job/" + strconv.FormatInt(job.ID, 10),
			Company:      "Microsoft",
		})
	}

	return postings, parsed.Data.Count, nil
}
