package sites

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/vanshsinhaa/jobscanner/common"
)

// Atlassian (Jira/Confluence/Trello) publishes all openings as a flat JSON list
// (iCIMS-backed) at atlassian.com/endpoint/careers/listings. Single response.
// The "Trello" target maps here via the alias table in database/target_companies.go.
type atlassianListing struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Locations []string `json:"locations"`
	Category  string   `json:"category"`
	ApplyURL  string   `json:"applyUrl"`
	PortalJob struct {
		PortalURL   string `json:"portalUrl"`
		UpdatedDate string `json:"updatedDate"`
	} `json:"portalJobPost"`
}

func GetAtlassianJobs() ([]common.JobPosting, error) {
	fmt.Println("Processing: ", "Atlassian")
	client := common.GetClient()

	url := "https://www.atlassian.com/endpoint/careers/listings"
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)").
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching atlassian jobs: %v", err)
	}

	var listings []atlassianListing
	if err := json.Unmarshal(resp.Body(), &listings); err != nil {
		return nil, fmt.Errorf("error parsing atlassian response: %v", err)
	}

	var jobPostings []common.JobPosting
	for _, l := range listings {
		externalURL := l.PortalJob.PortalURL
		if externalURL == "" {
			externalURL = l.ApplyURL
		}
		jobPostings = append(jobPostings, common.JobPosting{
			Company:      "Atlassian",
			JobId:        common.Atlassian + ":" + strconv.FormatInt(l.ID, 10),
			JobTitle:     l.Title,
			Location:     strings.Join(l.Locations, "; "),
			PostedOn:     l.PortalJob.UpdatedDate,
			ExternalPath: externalURL,
		})
	}

	return jobPostings, nil
}
