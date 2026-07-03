# Contributing

Thanks for helping keep the board alive. Scrapers rot — companies migrate ATS platforms, rename facets, and delete tenants without warning — so the most valuable contributions are **new company scrapers** and **fixes for broken ones**. Both are usually small.

## Data repo skeleton

When forking your own board (README quickstart step 1), your data repo's `README.md` must contain exactly these two sections — the table writers locate them by header text, separator row, and end anchor:

```markdown
## 🎓 Intern & New Grad Opportunities

| Company | Role | Location | Apply | Posted |
| --- | --- | --- | :---: | :---: |
<!-- intern-rows-end -->

## 💼 All SWE Opportunities

| Company | Role | Location | Apply | Posted |
| --- | --- | --- | :---: | :---: |
<!-- general-rows-end -->
```

## Dev setup

```bash
git clone https://github.com/vanshsinhaa/jobscanner
cd jobscanner
go build ./...
go run ./cmd/baseline   # runs every scraper in isolation — no DB or README writes
```

`cmd/baseline` prints a per-company table (job count, latency, sample titles) and a `ZERO OR ERROR` section. That section is the to-do list.

## Adding a company

### Step 0: find out what ATS they use

Open the company's careers page, open your browser's network tab, and look at where the job list request goes. Shortcuts — try these public endpoints directly (replace `TOKEN` with the company slug):

| ATS | Test endpoint | Sign it's this one |
|---|---|---|
| Greenhouse | `https://boards-api.greenhouse.io/v1/boards/TOKEN/jobs` | careers links contain `gh_jid=` or `greenhouse.io` |
| Ashby | `https://api.ashbyhq.com/posting-api/job-board/TOKEN` | links to `jobs.ashbyhq.com` or `ashby_jid=` |
| Lever | `https://api.lever.co/v0/postings/TOKEN?mode=json` | links to `jobs.lever.co` |
| Workday | `https://TENANT.wdN.myworkdayjobs.com/wday/cxs/TENANT/SITE/jobs` (POST) | URL is `*.myworkdayjobs.com` |
| Eightfold | `https://TENANT.eightfold.ai/api/pcsx/search?domain=DOMAIN&start=0` — if that says "PCSX is not enabled", try `/api/apply/v2/jobs?domain=DOMAIN&start=0&num=10` | careers site is `*.eightfold.ai` or has `pid=` job URLs |
| Oracle HCM | `https://HOST/hcmRestApi/resources/latest/recruitingCEJobRequisitions?onlyData=true&expand=requisitionList.secondaryLocations&finder=findReqs;siteNumber=CX_N,limit=10,offset=0` | careers URL contains `/en/sites/CX_` |
| Phenom | POST `https://careers.COMPANY.com/widgets` with a `refineSearch` payload | careers site has `/widgets` XHR calls |

### Step 1: Greenhouse / Ashby / Lever (the easy 90%)

1. **Company code** — add a 4-letter code to `common/variables.go` (both the `var` block and the `values` list; the init check will crash on duplicates).
2. **Wrapper** — one line in the matching engine file:

```go
// sites/greenhouse.go
func GetAcmeJobs() ([]common.JobPosting, error) {
	return fetchGreenhouseJobs("Acme", common.Acme, "acme")
}
```

(`fetchAshbyJobs` / `fetchLeverJobs` for the other two — identical shape.)

3. **Dispatch** — add a case to `sites_main/sites_registry.go`:

```go
case common.Acme:
	return sites.GetAcmeJobs()
```

The first argument to `fetch*Jobs` is the display name. **It must match the name people would put in `target_companies.json`** — that lookup is an exact case-insensitive match. If the board is owned by a parent brand (like Square postings on Block's board), either emit the tracked name or add an alias in `database/target_companies.go`.

### Step 2: Workday

Copy any file in `workday/` (e.g. `zoom.go` for a no-facet example, `nvidia.go` for a facet-filtered one), change tenant/site/URLs, register the code in `common/variables_workday.go`. Facet GUIDs come from the careers site's network tab — apply filters in the UI and copy the `appliedFacets` object from the request body. Keep `"offset": %d` — the engine paginates by substituting it.

### Step 3: verify

```bash
go build ./...
go run ./cmd/baseline
```

Your company should appear with a plausible count and real titles. Spot-check a sample URL from the output — it should land on the actual posting.

## Fixing a broken scraper

The debugging playbook, with the root cause of every historical outage, is in [`plan/tracker.md`](plan/tracker.md). The short version:

1. **`curl` the endpoint with filters removed / empty facets.**
   - Works unfiltered but 0 with filters → **facet values went stale.** Re-pull the facet/aggregation list and update the payload (this happened to ABB, Snowflake, IBM).
   - HTTP 422/410 on Workday → **tenant migrated datacenters.** Search `site:myworkdayjobs.com COMPANY` and bump the `wdN` host (Bose wd1→wd503, Walmart wd5→wd504).
   - 404 / config-JSON-instead-of-jobs → **company changed ATS entirely.** Re-run the Step 0 discovery (Amex went Eightfold→Oracle; Chime went custom→Greenhouse).
2. **Eightfold quirk**: PCSX tenants return `count: 0` from the legacy `/api/apply/v2/jobs` endpoint without erroring. Always try `/api/pcsx/search` too.
3. **Oracle HCM quirk**: without `expand=requisitionList.secondaryLocations` the API reports a correct `TotalJobsCount` but an **empty** `requisitionList`.
4. Update `plan/tracker.md` with the root cause — the ledger is how the next person avoids re-diagnosing.

## PR checklist

- [ ] `go build ./...` and `go vet ./...` pass
- [ ] `go run ./cmd/baseline` shows your company with a non-zero count (paste the line in the PR)
- [ ] Display name matches what a user would write in `target_companies.json` (or an alias is added)
- [ ] `plan/tracker.md` updated (new row for additions, root cause for fixes)
- [ ] No secrets, no personal data, no scraping of authenticated/private endpoints

## Ground rules for scrapers

- Public, unauthenticated endpoints only. If it needs a login, a real browser, or CAPTCHA evasion, we don't scrape it (see Tesla/LinkedIn in the tracker).
- Respect the shared HTTP client's retry/backoff (`common.GetClient()`) — don't hand-roll hammering loops.
- Emit stable `JobId`s (`CODE:external-id`) — dedup depends on them never changing for the same posting.
