package sites

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/vanshsinhaa/jobscanner/common"
)

// PayPal runs on Eightfold's newer PCSX platform. The legacy SmartApply endpoint
// (/api/apply/v2/jobs) returns zero positions for PCSX tenants — use /api/pcsx/search.
// Pagination: 10 positions per page via the start offset.
type paypalPCSXMain struct {
	Status int64          `json:"status"`
	Data   paypalPCSXData `json:"data"`
}

type paypalPCSXData struct {
	Count     int64                `json:"count"`
	Positions []paypalPCSXPosition `json:"positions"`
}

type paypalPCSXPosition struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Locations []string `json:"locations"`
}

func GetPayPalJobs() ([]common.JobPosting, error) {
	fmt.Println("Processing: ", "PayPal")

	var jobPostings []common.JobPosting
	start := 0
	total := int64(1)

	for int64(start) < total {
		page, count, err := paypalPCSXPage(start)
		if err != nil {
			if start == 0 {
				return nil, err
			}
			fmt.Println("warn: paypal page fetch failed at offset", start, ":", err)
			break
		}
		total = count
		if len(page) == 0 {
			break
		}
		jobPostings = append(jobPostings, page...)
		start += 10
	}

	return jobPostings, nil
}

func paypalPCSXPage(start int) ([]common.JobPosting, int64, error) {
	client := common.GetClient()

	url := fmt.Sprintf("https://paypal.eightfold.ai/api/pcsx/search?domain=paypal.com&query=&location=&start=%d", start)
	resp, err := client.R().SetHeader("Accept", "application/json").Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("error fetching paypal jobs: %v", err)
	}

	var main paypalPCSXMain
	if err := json.Unmarshal(resp.Body(), &main); err != nil {
		return nil, 0, fmt.Errorf("error parsing paypal response: %v", err)
	}

	var jobPostings []common.JobPosting
	for _, p := range main.Data.Positions {
		id := strconv.FormatInt(p.ID, 10)
		jobPostings = append(jobPostings, common.JobPosting{
			Company:      "PayPal",
			JobId:        common.PayPal + ":" + id,
			JobTitle:     p.Name,
			Location:     strings.Join(p.Locations, "; "),
			ExternalPath: "https://paypal.eightfold.ai/careers/job/" + id + "?domain=paypal.com",
		})
	}

	return jobPostings, main.Data.Count, nil
}
