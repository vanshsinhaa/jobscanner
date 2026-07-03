package sites

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/vanshsinhaa/jobscanner/common"
)

// Greenhouse Job Board API: https://boards-api.greenhouse.io/v1/boards/{board}/jobs
// Returns every published job in a single response — no pagination needed.
type greenhouseMain struct {
	Jobs []greenhouseJob `json:"jobs"`
}

type greenhouseJob struct {
	ID          int64              `json:"id"`
	Title       string             `json:"title"`
	AbsoluteURL string             `json:"absolute_url"`
	UpdatedAt   string             `json:"updated_at"`
	Location    greenhouseLocation `json:"location"`
}

type greenhouseLocation struct {
	Name string `json:"name"`
}

// fetchGreenhouseJobs pulls all published jobs from a Greenhouse job board.
// company is the display name (must match target_companies.json for target tracking),
// code the internal company code, board the Greenhouse board token.
func fetchGreenhouseJobs(company, code, board string) ([]common.JobPosting, error) {
	fmt.Println("Processing: ", company)
	client := common.GetClient()

	url := "https://boards-api.greenhouse.io/v1/boards/" + board + "/jobs"
	resp, err := client.R().SetHeader("Accept", "application/json").Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching greenhouse jobs (%s): %v", company, err)
	}

	var gh greenhouseMain
	if err := json.Unmarshal(resp.Body(), &gh); err != nil {
		return nil, fmt.Errorf("error parsing greenhouse response (%s): %v", company, err)
	}

	var jobPostings []common.JobPosting
	for _, job := range gh.Jobs {
		jobPostings = append(jobPostings, common.JobPosting{
			Company:      company,
			JobId:        code + ":" + strconv.FormatInt(job.ID, 10),
			JobTitle:     job.Title,
			Location:     job.Location.Name,
			PostedOn:     job.UpdatedAt,
			ExternalPath: job.AbsoluteURL,
		})
	}

	return jobPostings, nil
}

func GetStripeJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Stripe", common.Stripe, "stripe")
}

func GetAnthropicJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Anthropic", common.Anthropic, "anthropic")
}

func GetPinterestJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Pinterest", common.Pinterest, "pinterest")
}

func GetAirbnbJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Airbnb", common.Airbnb, "airbnb")
}

func GetLyftJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Lyft", common.Lyft, "lyft")
}

func GetDoorDashJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("DoorDash", common.DoorDash, "doordashusa")
}

func GetInstacartJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Instacart", common.Instacart, "instacart")
}

func GetCoinbaseJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Coinbase", common.Coinbase, "coinbase")
}

func GetRobinhoodJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Robinhood", common.Robinhood, "robinhood")
}

// Square postings live on Block's Greenhouse board (Block is Square's parent).
// Company is set to "Square" to match the target_companies.json entry.
func GetSquareJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Square", common.Square, "block")
}

func GetAsanaJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Asana", common.Asana, "asana")
}

func GetFigmaJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Figma", common.Figma, "figma")
}

// X Corp (Twitter) merged into xAI in March 2025; X/xAI roles are on the xai board.
// The "Twitter" target maps here via the alias table in database/target_companies.go.
func GetXAIJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("xAI", common.XAI, "xai")
}
