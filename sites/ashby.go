package sites

import (
	"encoding/json"
	"fmt"

	"github.com/vanshsinhaa/jobscanner/common"
)

// Ashby public posting API: https://api.ashbyhq.com/posting-api/job-board/{org}
// Returns every listed job in a single response.
type ashbyMain struct {
	Jobs []ashbyJob `json:"jobs"`
}

type ashbyJob struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Location    string `json:"location"`
	JobURL      string `json:"jobUrl"`
	PublishedAt string `json:"publishedAt"`
	IsListed    bool   `json:"isListed"`
}

// fetchAshbyJobs pulls all listed jobs from an Ashby job board.
func fetchAshbyJobs(company, code, org string) ([]common.JobPosting, error) {
	fmt.Println("Processing: ", company)
	client := common.GetClient()

	url := "https://api.ashbyhq.com/posting-api/job-board/" + org
	resp, err := client.R().SetHeader("Accept", "application/json").Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching ashby jobs (%s): %v", company, err)
	}

	var ab ashbyMain
	if err := json.Unmarshal(resp.Body(), &ab); err != nil {
		return nil, fmt.Errorf("error parsing ashby response (%s): %v", company, err)
	}

	var jobPostings []common.JobPosting
	for _, job := range ab.Jobs {
		if !job.IsListed {
			continue
		}
		jobPostings = append(jobPostings, common.JobPosting{
			Company:      company,
			JobId:        code + ":" + job.ID,
			JobTitle:     job.Title,
			Location:     job.Location,
			PostedOn:     job.PublishedAt,
			ExternalPath: job.JobURL,
		})
	}

	return jobPostings, nil
}

func GetOpenAIJobs() ([]common.JobPosting, error) {
	return fetchAshbyJobs("OpenAI", common.OpenAI, "openai")
}

func GetNotionJobs() ([]common.JobPosting, error) {
	return fetchAshbyJobs("Notion", common.Notion, "notion")
}
