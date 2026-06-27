package process

import (
	"fmt"

	"github.com/vanshsinhaa/jobscanner/common"
	"github.com/vanshsinhaa/jobscanner/workday"
	workdaymain "github.com/vanshsinhaa/jobscanner/workday_main"
)

func GetAllWorkdayJobs() ([]common.JobPosting, error) {
	workday.Init()
	var allJobs []common.JobPosting

	for company := range common.WorkdayCompanies {
		payload, ok := workdaymain.WorkdayPayloads[company]
		if !ok {
			fmt.Printf("No Workday payload found for company: %s\n", company)
			continue
		}

		jobs, err := workdaymain.GetWorkdayJobs(payload)
		if err != nil {
			fmt.Printf("Error fetching jobs for %s: %v\n", company, err)
			continue
		}

		allJobs = append(allJobs, jobs...)
	}

	return allJobs, nil
}
