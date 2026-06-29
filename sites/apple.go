package sites

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/vanshsinhaa/jobscanner/common"
)

// Apple migrated search from /api/role/search to /api/v1/search.
// The new endpoint requires:
//   1. A session established via GET /api/v1/CSRFToken (sets jssid cookie + X-Apple-CSRF-Token header)
//   2. The POST body must include a "format" field — without it the API returns 0 results silently
//   3. The same http.Client (with CookieJar) for both requests so jssid travels to the POST
//   4. Teams filter format changed: {"teams.teamID":...,"teams.subTeamID":...} → {"team":...,"subTeam":...}
//   5. Response is wrapped: {"res": {"searchResults":[...], "totalRecords":N}}

const (
	appleCSRFURL    = "https://jobs.apple.com/api/v1/CSRFToken"
	appleSearchURL  = "https://jobs.apple.com/api/v1/search"
	appleJobBaseURL = "https://jobs.apple.com/en-us/details/"
	maxAppleAgeDays = 60
)

var appleUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"

type appleV1Response struct {
	Res struct {
		SearchResults []appleV1Job `json:"searchResults"`
		TotalRecords  int          `json:"totalRecords"`
	} `json:"res"`
}

type appleV1Job struct {
	ReqID         string           `json:"reqId"`
	PostingTitle  string           `json:"postingTitle"`
	Locations     []appleV1Location `json:"locations"`
	PostDateInGMT string           `json:"postDateInGMT"`
}

type appleV1Location struct {
	City          string `json:"city"`
	StateProvince string `json:"stateProvince"`
	CountryName   string `json:"countryName"`
}

type appleTeamFilter struct {
	Team    string `json:"team"`
	SubTeam string `json:"subTeam"`
}

// appleTeams covers both intern and university queries. Apple posts internships under
// STDNT (Student Programs) AND under SFTWR/MLAI when the role is managed by that
// engineering team directly. The "intern" query text-search already scopes results —
// FTE roles that happen to mention "intern" in their description are filtered out at
// classification time (ClassifyRole checks the title, not the description).
var appleTeams = []appleTeamFilter{
	{Team: "teamsAndSubTeams-STDNT", SubTeam: "subTeam-INTRN"},
	{Team: "teamsAndSubTeams-STDNT", SubTeam: "subTeam-CORP"},
	{Team: "teamsAndSubTeams-STDNT", SubTeam: "subTeam-ACR"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-CLD"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-COS"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-DSR"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-EPM"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-ISTECH"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-SEC"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-SQAT"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-WSFT"},
	{Team: "teamsAndSubTeams-SFTWR", SubTeam: "subTeam-AF"},
	{Team: "teamsAndSubTeams-MLAI", SubTeam: "subTeam-MLI"},
	{Team: "teamsAndSubTeams-MLAI", SubTeam: "subTeam-DLRL"},
	{Team: "teamsAndSubTeams-MLAI", SubTeam: "subTeam-NLP"},
	{Team: "teamsAndSubTeams-MLAI", SubTeam: "subTeam-CV"},
	{Team: "teamsAndSubTeams-MLAI", SubTeam: "subTeam-AR"},
}

func GetAppleJobs() ([]common.JobPosting, error) {
	fmt.Println("Processing: ", "Apple")

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	csrfToken, err := appleFetchCSRF(client)
	if err != nil {
		return nil, fmt.Errorf("apple CSRF: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -maxAppleAgeDays)

	seen := make(map[string]bool)
	var all []common.JobPosting

	for _, query := range []string{"intern", "university"} {
		jobs, err := appleFetchAll(client, csrfToken, query, cutoff)
		if err != nil {
			fmt.Printf("apple query %q error: %v\n", query, err)
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

func appleFetchCSRF(client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", appleCSRFURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", appleUA)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://jobs.apple.com/en-us/search")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	token := resp.Header.Get("X-Apple-CSRF-Token")
	if token == "" {
		return "", fmt.Errorf("X-Apple-CSRF-Token header not found in response")
	}
	return token, nil
}

func appleFetchAll(client *http.Client, csrfToken, query string, cutoff time.Time) ([]common.JobPosting, error) {
	jobs, total, err := applePage(client, csrfToken, query, 1)
	if err != nil {
		return nil, err
	}

	pages := int(math.Ceil(float64(total) / 20.0))
	for pg := 2; pg <= pages; pg++ {
		more, _, pageErr := applePage(client, csrfToken, query, pg)
		if pageErr != nil {
			fmt.Printf("warn: Apple page %d (query=%q): %v\n", pg, query, pageErr)
			continue
		}
		jobs = append(jobs, more...)
	}

	// For the "university" query, tag jobs as new_grad. University hires sometimes
	// have generic titles that ClassifyRole can't identify from keywords alone.
	// The "intern" query is NOT tagged here — STDNT-team intern postings always
	// contain "intern"/"internship" in the title and ClassifyRole handles them.
	if query == "university" {
		for i := range jobs {
			if jobs[i].RoleType == "" {
				jobs[i].RoleType = "new_grad"
			}
		}
	}

	var recent []common.JobPosting
	for _, j := range jobs {
		t, err := time.Parse("2006-01-02T15:04:05.000Z", j.PostedOn)
		if err != nil {
			t, err = time.Parse(time.RFC3339Nano, j.PostedOn)
		}
		if err == nil && t.Before(cutoff) {
			continue
		}
		recent = append(recent, j)
	}
	return recent, nil
}

func applePage(client *http.Client, csrfToken, query string, page int) ([]common.JobPosting, int, error) {
	payload := map[string]any{
		"query": query,
		"filters": map[string]any{
			"locations": []string{"postLocation-USA"},
			"teams":     appleTeams,
		},
		"page":   page,
		"locale": "en-us",
		"sort":   "newest",
		"format": map[string]string{
			"longDate":   "MMMM D, YYYY",
			"mediumDate": "MMM D, YYYY",
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", appleSearchURL, bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", appleUA)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Referer", "https://jobs.apple.com/en-us/search")
	req.Header.Set("X-Apple-CSRF-Token", csrfToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var parsed appleV1Response
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, 0, fmt.Errorf("json parse failed: %w", err)
	}

	var postings []common.JobPosting
	for _, job := range parsed.Res.SearchResults {
		location := "United States"
		if len(job.Locations) > 0 {
			loc := job.Locations[0]
			switch {
			case loc.City != "" && loc.StateProvince != "":
				location = loc.City + ", " + loc.StateProvince
			case loc.City != "":
				location = loc.City
			case loc.CountryName != "":
				location = loc.CountryName
			}
		}

		postedOn := job.PostDateInGMT
		if t, err := time.Parse("2006-01-02T15:04:05.000Z", job.PostDateInGMT); err == nil {
			postedOn = t.UTC().Format("2006-01-02T15:04:05Z")
		}

		postings = append(postings, common.JobPosting{
			JobId:        common.Apple + ":" + job.ReqID,
			JobTitle:     job.PostingTitle,
			Location:     location,
			PostedOn:     postedOn,
			ExternalPath: appleJobBaseURL + job.ReqID,
			Company:      "Apple",
		})
	}

	return postings, parsed.Res.TotalRecords, nil
}
