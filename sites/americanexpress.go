package sites

import (
	"encoding/json"
	"fmt"

	"github.com/vanshsinhaa/jobscanner/common"
)

// American Express moved careers off Eightfold onto Oracle Cloud HCM
// (careers.americanexpress.com -> egug.fa.us2.oraclecloud.com, site CX_1).
// The old aexp.eightfold.ai API still responds but always returns 0 positions.
// 300000081299474 is the "Technology" category facet ID.
type amexOracleMain struct {
	Items []amexOracleItem `json:"items"`
}

type amexOracleItem struct {
	TotalJobsCount  int64                  `json:"TotalJobsCount"`
	RequisitionList []amexOracleRequisition `json:"requisitionList"`
}

type amexOracleRequisition struct {
	ID              string `json:"Id"`
	Title           string `json:"Title"`
	PostedDate      string `json:"PostedDate"`
	PrimaryLocation string `json:"PrimaryLocation"`
}

func GetAmexJobs() ([]common.JobPosting, error) {
	fmt.Println("Processing: ", "American Express")
	client := common.GetClient()

	// expand=requisitionList.secondaryLocations is required: without it Oracle reports
	// TotalJobsCount but returns an empty requisitionList.
	url := "https://egug.fa.us2.oraclecloud.com/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList.secondaryLocations&finder=findReqs;siteNumber=CX_1,selectedCategoriesFacet=300000081299474,limit=200,offset=0,sortBy=POSTING_DATES_DESC"

	resp, err := client.R().SetHeader("Accept", "application/json").Get(url)
	if err != nil {
		return nil, fmt.Errorf("error creating API request(Amex jobs): %v", err)
	}

	var amexJobs amexOracleMain
	if err := json.Unmarshal(resp.Body(), &amexJobs); err != nil {
		return nil, fmt.Errorf("error parsing json (Amex jobs): %v", err)
	}
	if len(amexJobs.Items) == 0 {
		return nil, fmt.Errorf("amex oracle response contained no items")
	}

	var jobPostings []common.JobPosting
	for _, job := range amexJobs.Items[0].RequisitionList {
		jobPostings = append(jobPostings, common.JobPosting{
			Company:      "American Express",
			JobId:        common.Amex + ":" + job.ID,
			JobTitle:     job.Title,
			Location:     job.PrimaryLocation,
			PostedOn:     job.PostedDate,
			ExternalPath: "https://careers.americanexpress.com/en/sites/CX_1/job/" + job.ID,
		})
	}

	return jobPostings, nil
}
