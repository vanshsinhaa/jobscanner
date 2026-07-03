# Scraper Status Tracker

Last full audit: **2026-07-03** (local baseline run of every registered scraper).

## Summary

- 63 companies scraped across 4 engine types: Workday CXS, Greenhouse, Ashby, Lever, plus per-site custom scrapers (Oracle HCM, Eightfold PCSX, Phenom widgets, iCIMS, HTML/sitemap).
- Every company in `local_data/target_companies.json` now has coverage except **LinkedIn** and **Tesla** (see Blocked below).
- Target-section matching supports brand aliases (`database/target_companies.go: targetAliases`) and sub-brand retagging (`process/data_process.go: retagSubBrands`).

## Fixed in 2026-07-03 audit

| Company | Symptom | Root cause | Fix |
|---|---|---|---|
| ABB | 0 jobs | Phenom facet values renamed: country `USA` → `United States of America`; category `Digital` removed | Updated `selected_fields` in `sites/abb.go` |
| American Express | 0 jobs | Left Eightfold; careers.americanexpress.com now fronts Oracle Cloud HCM (`egug.fa.us2.oraclecloud.com`, site CX_1) | Rewrote `sites/americanexpress.go` against Oracle HCM REST (Technology category facet `300000081299474`). NOTE: `expand=requisitionList.secondaryLocations` is mandatory or Oracle returns an empty list |
| Chime | 0 jobs | Old HTML careers site is bot-protected (403); Chime is on Greenhouse | `sites/chime.go` → Greenhouse board `chime` |
| IBM | 0 jobs | Search API scope renamed `careers` → `careers2` | Updated scope + widened level filter to `Entry Level` + `Internship` in `sites/ibm.go` |
| Snowflake | 0 jobs | Phenom facets renamed: region `AMS` → `Americas`; categories `IT`/`Security`/`Product` gone | Updated payload in `sites/snowflake.go` (Engineering, Enterprise Technology, Data/Analytics/AI × Americas) |
| Google | junk rows ("Help link" etc.) | `.WpHeLc` selector also matches nav anchors | Filter: only anchors with `Learn more about` aria-label + `jobs/results/` href (`sites/google.go`) |
| Bose | 0 jobs (HTTP 422) | Workday tenant migrated wd1 → **wd503** | URL bump in `workday/bose.go` |
| Symbotic | 0 jobs (HTTP 422) | Workday tenant migrated wd1 → **wd504** | URL bump in `workday/symbotic.go` |
| Walmart | 0 jobs (ERR_TENANT_MIGRATED) | Workday tenant migrated wd5 → **wd504** | URL bump in `workday/walmart.go` |
| Twitter | 0 jobs (HTTP 422) | `twitter.wd5/X` tenant deleted — X Corp merged into xAI (careers.x.com → x.ai/careers) | Removed `workday/twitter.go`; added **xAI** Greenhouse scraper (`sites/greenhouse.go`); alias `twitter → xai` |

## Added in 2026-07-03 audit (target-company coverage)

| Company | Engine | Board/endpoint | Notes |
|---|---|---|---|
| Stripe | Greenhouse | `stripe` | |
| Anthropic | Greenhouse | `anthropic` | |
| Pinterest | Greenhouse | `pinterest` | |
| Airbnb | Greenhouse | `airbnb` | |
| Lyft | Greenhouse | `lyft` | |
| DoorDash | Greenhouse | `doordashusa` | |
| Instacart | Greenhouse | `instacart` | |
| Coinbase | Greenhouse | `coinbase` | |
| Robinhood | Greenhouse | `robinhood` | |
| Square | Greenhouse | `block` | Block is Square's parent; scraper emits Company="Square" |
| Asana | Greenhouse | `asana` | |
| Figma | Greenhouse | `figma` | |
| xAI | Greenhouse | `xai` | Covers X/Twitter roles post-merger |
| OpenAI | Ashby | `openai` | |
| Notion | Ashby | `notion` | |
| Palantir | Lever | `palantir` | |
| PayPal | Eightfold PCSX | `paypal.eightfold.ai/api/pcsx/search` | Legacy SmartApply endpoint returns 0 for PCSX tenants |
| Zoom | Workday | `zoom.wd5 / Zoom` | `workday/zoom.go` |
| Atlassian (Trello) | iCIMS JSON | `atlassian.com/endpoint/careers/listings` | Alias `trello → atlassian` |
| Shopify | Own site (sitemap) | `shopify.com/careers/sitemap.xml` | Careers app is a streamed React app with the Ashby public API disabled; titles reconstructed from URL slugs, location not available |

## Target matching aliases

`target_companies.json` name → scraped company name:

- `Amex` → "American Express"
- `Snap` → "Snapchat"
- `Twitter` → "xAI"
- `Trello` → "Atlassian"
- `Annapurna Labs` → retagged from Amazon postings whose title mentions "Annapurna"
- `Slack` → retagged from Salesforce Workday postings whose title mentions "Slack"

## Blocked / known limitations

| Company | Status | Detail |
|---|---|---|
| Tesla | **Blocked** | `tesla.com/cua-api` behind Akamai bot management; plain HTTP (any headers) gets 403. Needs a real browser/session; scraper left in place and fails gracefully |
| LinkedIn | **Blocked** | No public ATS/job-board API; careers.linkedin.com is their own platform behind auth/bot walls. In target list but no scraper |
| ManTech | **Removed from rotation** | `mantech.wd1/External` responds but returns `total: 0` for all API queries while the public page still renders jobs (server-side allowlisting suspected). Code deleted from active list (`common/variables_workday.go`) |
| Shopify | Partial | No location data (sitemap-only). Titles derived from slugs |
| Microsoft | By design | ~12–15 jobs: 45-day recency cutoff + intern/university text queries only (Eightfold ignores level filters) |

## Verification protocol

Run the isolated baseline harness (no DB writes, no README writes):

```
go run ./cmd/baseline
```

It prints a per-company table (count, latency, sample titles) and a ZERO-OR-ERROR section. Any company in that section needs investigation — see the probes in this file's history for the endpoint-testing playbook (curl the API with empty filters first; if that works the facet values are stale, if 404/422 the tenant/site moved).
