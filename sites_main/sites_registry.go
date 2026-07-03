package sitesmain

import (
	"fmt"

	"github.com/vanshsinhaa/jobscanner/common"
	"github.com/vanshsinhaa/jobscanner/sites"
)

// New function to fetch jobs based on company name
func FetchJobsByCompany(company string) ([]common.JobPosting, error) {
	switch company {
	case common.Google:
		return sites.GetGoogleJobs()
	case common.Microsoft:
		return sites.GetMicrosoftJobs()
	case common.Oracle:
		return sites.GetOracleJobs()
	case common.Apple:
		return sites.GetAppleJobs()
	case common.Meta:
		return sites.GetMetaJobs()
	case common.Tesla:
		return sites.GetTeslaJobs()
	case common.Chime:
		return sites.GetChimeJobs()
	case common.Visa:
		return sites.GetVisaJobs()
	case common.Uber:
		return sites.GetUberJobs()
	case common.Databricks:
		return sites.GetDatabricksJobs()
	case common.Amazon:
		return sites.GetAmazonJobs()
	case common.Amex:
		return sites.GetAmexJobs()
	case common.Snowflake:
		return sites.GetSnowflakeJobs()
	case common.Intuit:
		return sites.GetIntuitJobs()
	case common.IBM:
		return sites.GetIBMJobs()
	case common.ABB:
		return sites.GetABBJobs()
	case common.AMD:
		return sites.GetAMDJobs()
	case common.AkunaCapital:
		return sites.GetAkunaCapitalJobs()
	case common.Fortinet:
		return sites.GetFortinetJobs()
	case common.Reddit:
		return sites.GetRedditJobs()
	case common.NetApp:
		return sites.GetNetAppJobs()
	case common.Nokia:
		return sites.GetNokiaJobs()
	case common.Stripe:
		return sites.GetStripeJobs()
	case common.Anthropic:
		return sites.GetAnthropicJobs()
	case common.Pinterest:
		return sites.GetPinterestJobs()
	case common.Airbnb:
		return sites.GetAirbnbJobs()
	case common.Lyft:
		return sites.GetLyftJobs()
	case common.DoorDash:
		return sites.GetDoorDashJobs()
	case common.Instacart:
		return sites.GetInstacartJobs()
	case common.Coinbase:
		return sites.GetCoinbaseJobs()
	case common.Robinhood:
		return sites.GetRobinhoodJobs()
	case common.Square:
		return sites.GetSquareJobs()
	case common.Asana:
		return sites.GetAsanaJobs()
	case common.Figma:
		return sites.GetFigmaJobs()
	case common.XAI:
		return sites.GetXAIJobs()
	case common.OpenAI:
		return sites.GetOpenAIJobs()
	case common.Notion:
		return sites.GetNotionJobs()
	case common.Palantir:
		return sites.GetPalantirJobs()
	case common.PayPal:
		return sites.GetPayPalJobs()
	case common.Shopify:
		return sites.GetShopifyJobs()
	case common.Atlassian:
		return sites.GetAtlassianJobs()
	default:
		return nil, fmt.Errorf("unknown company: %s", company)
	} //////////////////////// Edit here
}
