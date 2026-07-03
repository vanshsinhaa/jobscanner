package sites

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/vanshsinhaa/jobscanner/common"
)

// Lever public postings API: https://api.lever.co/v0/postings/{org}?mode=json
// Returns every published posting in a single response.
type leverJob struct {
	ID         string          `json:"id"`
	Text       string          `json:"text"`
	HostedURL  string          `json:"hostedUrl"`
	CreatedAt  int64           `json:"createdAt"`
	Categories leverCategories `json:"categories"`
}

type leverCategories struct {
	Location string `json:"location"`
}

// fetchLeverJobs pulls all published postings from a Lever job board.
func fetchLeverJobs(company, code, org string) ([]common.JobPosting, error) {
	fmt.Println("Processing: ", company)
	client := common.GetClient()

	url := "https://api.lever.co/v0/postings/" + org + "?mode=json"
	resp, err := client.R().SetHeader("Accept", "application/json").Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching lever jobs (%s): %v", company, err)
	}

	var postings []leverJob
	if err := json.Unmarshal(resp.Body(), &postings); err != nil {
		return nil, fmt.Errorf("error parsing lever response (%s): %v", company, err)
	}

	var jobPostings []common.JobPosting
	for _, job := range postings {
		jobPostings = append(jobPostings, common.JobPosting{
			Company:      company,
			JobId:        code + ":" + job.ID,
			JobTitle:     job.Text,
			Location:     job.Categories.Location,
			PostedOn:     strconv.FormatInt(job.CreatedAt, 10),
			ExternalPath: job.HostedURL,
		})
	}

	return jobPostings, nil
}

func GetPalantirJobs() ([]common.JobPosting, error) {
	return fetchLeverJobs("Palantir", common.Palantir, "palantir")
}
