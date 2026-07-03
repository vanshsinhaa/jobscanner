package sites

import (
	"github.com/vanshsinhaa/jobscanner/common"
)

// Chime moved its careers site off the old careers.chime.com HTML pages (now
// bot-protected, returns 403) onto Greenhouse. Use the public board API instead.
func GetChimeJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Chime", common.Chime, "chime")
}
