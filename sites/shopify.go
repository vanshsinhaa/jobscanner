package sites

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/vanshsinhaa/jobscanner/common"
)

// Shopify hosts jobs on its own domain (Ashby-backed internally, but the public
// Ashby posting API is disabled and the careers page is a streamed React app).
// The careers sitemap lists every live posting as /careers/{title-slug}_{uuid};
// the title is reconstructed from the slug. Location is not available this way.
var shopifyJobURLRe = regexp.MustCompile(`https://www\.shopify\.com/careers/([a-z0-9-]+)_([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`)

func GetShopifyJobs() ([]common.JobPosting, error) {
	fmt.Println("Processing: ", "Shopify")
	client := common.GetClient()

	resp, err := client.R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)").
		Get("https://www.shopify.com/careers/sitemap.xml")
	if err != nil {
		return nil, fmt.Errorf("error fetching shopify sitemap: %v", err)
	}

	matches := shopifyJobURLRe.FindAllStringSubmatch(string(resp.Body()), -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("shopify sitemap contained no job URLs (format changed?)")
	}

	seen := make(map[string]bool)
	var jobPostings []common.JobPosting
	for _, m := range matches {
		fullURL, slug, id := m[0], m[1], m[2]
		if seen[id] {
			continue
		}
		seen[id] = true
		jobPostings = append(jobPostings, common.JobPosting{
			Company:      "Shopify",
			JobId:        common.Shopify + ":" + id,
			JobTitle:     slugToTitle(slug),
			Location:     "See posting",
			ExternalPath: fullURL,
		})
	}

	return jobPostings, nil
}

// slugToTitle converts "senior-software-engineer-intern" to "Senior Software Engineer Intern".
func slugToTitle(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
