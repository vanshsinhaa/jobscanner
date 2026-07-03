package workday

import (
	"github.com/vanshsinhaa/jobscanner/common"
	workdaymain "github.com/vanshsinhaa/jobscanner/workday_main"
)

func init() {
	workdaymain.RegisterPayload(common.Zoom, common.WorkdayPayload{
		Company: "Zoom",
		CmpCode: common.Zoom,
		PreURL:  "https://zoom.wd5.myworkdayjobs.com/en-US/Zoom",
		JobsURL: "https://zoom.wd5.myworkdayjobs.com/wday/cxs/zoom/Zoom/jobs",
		PayLoad: `{
  "appliedFacets": {},
  "limit": 20,
  "offset": %d,
  "searchText": ""
}`,
	})
}
