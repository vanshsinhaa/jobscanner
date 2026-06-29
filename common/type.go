package common

// Job represents the structure to hold job listing data
type JobPosting struct {
	Company      string `json:"company,omitempty"`
	JobId        string `json:"jobId"`
	JobTitle     string `json:"title"`
	Location     string `json:"locationsText,omitempty"`
	PostedOn     string `json:"postedOn,omitempty"`
	ExternalPath string `json:"externalPath"`
	// RoleType may be set by scrapers that know the role category from context
	// (e.g. Apple sets "intern" for intern-query results, "new_grad" for university-query).
	// When non-empty, InsertIntoDB uses it directly instead of calling ClassifyRole.
	// Empty means ClassifyRole should infer from the title.
	RoleType string `json:"roleType,omitempty"`
}

// JobsResponse represents the structure of the full response
type JobsResponse struct {
	JobPostings []JobPosting `json:"jobPostings"`
	Total       int          `json:"total"`
}

// WorkdayPayload
type WorkdayPayload struct {
	Company string
	CmpCode string
	PreURL  string
	JobsURL string
	PayLoad string
}
