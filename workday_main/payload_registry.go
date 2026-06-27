package workdaymain

import "github.com/vanshsinhaa/jobscanner/common"

var WorkdayPayloads = map[string]common.WorkdayPayload{}

func RegisterPayload(companyCode string, payload common.WorkdayPayload) {
	WorkdayPayloads[companyCode] = payload
}
