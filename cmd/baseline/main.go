// Command baseline runs every registered scraper in isolation and prints a
// per-company report (count, error, sample titles). It performs NO database
// writes and NO README updates — diagnostic use only.
package main

import (
	"fmt"
	"sort"
	"time"

	"github.com/vanshsinhaa/jobscanner/common"
	sitesmain "github.com/vanshsinhaa/jobscanner/sites_main"
	_ "github.com/vanshsinhaa/jobscanner/workday" // registers workday payloads via init()
	workdaymain "github.com/vanshsinhaa/jobscanner/workday_main"
)

var siteNames = map[string]string{
	common.Google: "Google", common.Microsoft: "Microsoft", common.Oracle: "Oracle",
	common.Apple: "Apple", common.Meta: "Meta", common.Tesla: "Tesla",
	common.Chime: "Chime", common.Visa: "Visa", common.Uber: "Uber",
	common.Databricks: "Databricks", common.Amazon: "Amazon", common.Amex: "Amex",
	common.Snowflake: "Snowflake", common.Intuit: "Intuit", common.IBM: "IBM",
	common.ABB: "ABB", common.AMD: "AMD", common.AkunaCapital: "AkunaCapital",
	common.Fortinet: "Fortinet", common.Reddit: "Reddit", common.NetApp: "NetApp",
	common.Nokia: "Nokia", common.Stripe: "Stripe", common.Anthropic: "Anthropic",
	common.Pinterest: "Pinterest", common.Airbnb: "Airbnb", common.Lyft: "Lyft",
	common.DoorDash: "DoorDash", common.Instacart: "Instacart", common.Coinbase: "Coinbase",
	common.Robinhood: "Robinhood", common.Square: "Square", common.Asana: "Asana",
	common.Figma: "Figma", common.XAI: "xAI", common.OpenAI: "OpenAI",
	common.Notion: "Notion", common.Palantir: "Palantir", common.PayPal: "PayPal",
	common.Shopify: "Shopify", common.Atlassian: "Atlassian",
}

type result struct {
	kind    string
	code    string
	name    string
	count   int
	dur     time.Duration
	err     error
	samples []string
}

func main() {
	var results []result

	for code := range common.SitesCompanies {
		name := siteNames[code]
		fmt.Printf("=== sites: %s (%s)\n", name, code)
		start := time.Now()
		jobs, err := sitesmain.FetchJobsByCompany(code)
		results = append(results, mkResult("sites", code, name, jobs, err, time.Since(start)))
	}

	for code := range common.WorkdayCompanies {
		p, ok := workdaymain.WorkdayPayloads[code]
		if !ok {
			results = append(results, result{kind: "workday", code: code, name: code, err: fmt.Errorf("no payload registered")})
			continue
		}
		fmt.Printf("=== workday: %s (%s)\n", p.Company, code)
		start := time.Now()
		jobs, err := workdaymain.GetWorkdayJobs(p)
		results = append(results, mkResult("workday", code, p.Company, jobs, err, time.Since(start)))
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].kind != results[j].kind {
			return results[i].kind < results[j].kind
		}
		return results[i].name < results[j].name
	})

	fmt.Println("\n\n========== BASELINE REPORT ==========")
	fmt.Printf("%-8s %-26s %6s  %8s  %s\n", "KIND", "COMPANY", "JOBS", "TIME", "ERROR / SAMPLES")
	for _, r := range results {
		status := ""
		if r.err != nil {
			status = "ERR: " + r.err.Error()
			if len(status) > 120 {
				status = status[:120]
			}
		} else if len(r.samples) > 0 {
			status = fmt.Sprintf("%.110s", r.samples[0])
		}
		fmt.Printf("%-8s %-26s %6d  %8s  %s\n", r.kind, r.name, r.count, r.dur.Round(time.Millisecond), status)
	}

	fmt.Println("\n--- ZERO OR ERROR ---")
	for _, r := range results {
		if r.count == 0 || r.err != nil {
			fmt.Printf("%s / %s (%s): count=%d err=%v\n", r.kind, r.name, r.code, r.count, r.err)
		}
	}
}

func mkResult(kind, code, name string, jobs []common.JobPosting, err error, dur time.Duration) result {
	r := result{kind: kind, code: code, name: name, count: len(jobs), dur: dur, err: err}
	for i, j := range jobs {
		if i >= 3 {
			break
		}
		r.samples = append(r.samples, fmt.Sprintf("%s | %s", j.JobTitle, j.ExternalPath))
	}
	return r
}
