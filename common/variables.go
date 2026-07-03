package common

import (
	"fmt"
	"os"
)

// companies with own job portals
var (
	Google       = "GOGL"
	Microsoft    = "MISF"
	Oracle       = "ORCL"
	Apple        = "APLE"
	Meta         = "META"
	Tesla        = "TSLA"
	Chime        = "CHME"
	Visa         = "VISA"
	Uber         = "UBER"
	Databricks   = "DTBS"
	Amazon       = "AMZN"
	Amex         = "AMEX"
	Snowflake    = "SFLK"
	Intuit       = "INTT"
	IBM          = "IBMM"
	ABB          = "ABBB"
	AMD          = "AMDD"
	AkunaCapital = "AKUN"
	Fortinet     = "FORT"
	Reddit       = "REDT"
	NetApp       = "NAPP"
	Nokia        = "NKIA"
	// Greenhouse boards
	Stripe    = "STRP"
	Anthropic = "ANTH"
	Pinterest = "PINT"
	Airbnb    = "ARBN"
	Lyft      = "LYFT"
	DoorDash  = "DASH"
	Instacart = "INSC"
	Coinbase  = "COIN"
	Robinhood = "RBHD"
	Square    = "SQRE" // Block board on Greenhouse
	Asana     = "ASNA"
	Figma     = "FIGM"
	XAI       = "XAII" // X (formerly Twitter) merged into xAI
	// Ashby boards
	OpenAI = "OPAI"
	Notion = "NOTN"
	// Lever boards
	Palantir = "PLTR"
	// Other portals
	PayPal    = "PYPL" // Eightfold PCSX
	Shopify   = "SHOP" // own site, sitemap-based
	Atlassian = "ATLS" // iCIMS via atlassian.com endpoint (Trello parent)
) //////////////////////// Edit here

var AllCompanies = make(map[string]bool)

var SitesCompanies = make(map[string]bool)

func checkDuplicatesComapnies() {
	values := []string{
		Google,
		Microsoft,
		Oracle,
		Apple,
		Meta,
		Tesla,
		Chime,
		Visa,
		Uber,
		Databricks,
		Amazon,
		Amex,
		Snowflake,
		Intuit,
		IBM,
		ABB,
		AMD,
		AkunaCapital,
		Fortinet,
		Reddit,
		NetApp,
		Nokia,
		Stripe,
		Anthropic,
		Pinterest,
		Airbnb,
		Lyft,
		DoorDash,
		Instacart,
		Coinbase,
		Robinhood,
		Square,
		Asana,
		Figma,
		XAI,
		OpenAI,
		Notion,
		Palantir,
		PayPal,
		Shopify,
		Atlassian,
	} ///////////////////////// Edit here

	for _, value := range values {
		if AllCompanies[value] || SitesCompanies[value] {
			fmt.Printf("Duplicate company code found: %s\n", value)
			os.Exit(1)
		} else {
			AllCompanies[value] = true
			SitesCompanies[value] = true
		}
	}
}
