package sites

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vanshsinhaa/jobscanner/common"
)

type akunaGreenhouseResponse struct {
	Jobs []akunaGreenhouseJob `json:"jobs"`
}

type akunaGreenhouseJob struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	AbsoluteURL string `json:"absolute_url"`
	Location    struct {
		Name string `json:"name"`
	} `json:"location"`
	FirstPublished string `json:"first_published"`
}

func GetAkunaCapitalJobs() ([]common.JobPosting, error) {
	fmt.Println("Processing: ", "Akuna Capital")
	client := common.GetClient()

	resp, err := client.R().Get("https://boards-api.greenhouse.io/v1/boards/akunacapital/jobs")
	if err != nil {
		return nil, fmt.Errorf("error fetching Akuna Capital jobs: %v", err)
	}

	var ghResp akunaGreenhouseResponse
	if err := json.Unmarshal(resp.Body(), &ghResp); err != nil {
		return nil, fmt.Errorf("error parsing Akuna Capital response: %v", err)
	}

	var jobPostings []common.JobPosting
	for _, job := range ghResp.Jobs {
		jobPostings = append(jobPostings, common.JobPosting{
			Company:      "Akuna Capital",
			JobId:        common.AkunaCapital + ":" + fmt.Sprintf("%d", job.ID),
			JobTitle:     job.Title,
			Location:     strings.TrimSpace(job.Location.Name),
			ExternalPath: job.AbsoluteURL,
			PostedOn:     job.FirstPublished,
		})
	}

	return jobPostings, nil
}
