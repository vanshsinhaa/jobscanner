# jobscanner

Automated SWE internship and job tracker. Scrapes 50+ companies every hour and publishes a live job board to **[vanshsinhaa/jobs](https://github.com/vanshsinhaa/jobs)** — no manual work required.

Built on top of [go-get-jobs](https://github.com/neyaadeez/go-get-jobs) with a full rewrite: SQLite backend, GitHub Actions CI, Discord webhooks, intern/new-grad classification, recency sorting, and a personalized target companies feed.

---

## Live job board

**[github.com/vanshsinhaa/jobs](https://github.com/vanshsinhaa/jobs)**

Updated hourly. Two tables:
- **Intern & New Grad** — roles classified by title (intern, co-op, new grad, entry level)
- **All SWE** — everything else

---

## Features

- **Hourly CI scraping** — GitHub Actions runs on a cron schedule, no server needed
- **50+ companies** — 23 custom scrapers + 30 Workday companies
- **Intern/new-grad classification** — word-boundary regex separates intern roles from general SWE
- **Recency sorting** — newest postings first; jobs older than 14 days filtered out
- **Deduplication** — `job_ids.json` persists between runs so each job only appears once
- **Discord notifications** — get a summary embed in Discord every time new jobs drop
- **Target companies feed** — personal watchlist updated hourly, lives in this repo's README
- **Watch mode** — daemon flag for local polling with per-job Discord alerts
- **JSON export** — `jobs.json` published alongside the README for programmatic access

---

## Architecture

Three repos, each with a single responsibility:

```
vanshsinhaa/jobscanner   <- you are here
  scraper code, CI workflow, personal target feed

vanshsinhaa/jobs         <- public output
  README.md (job board), job_ids.json, jobs.json

vanshsinhaa/jobs-web     <- planned
  static website reading from jobs.json
```

**How a CI run works:**

1. GitHub spins up a fresh Ubuntu runner at the top of every hour
2. Checks out this repo (`scraper/`) and the jobs repo (`jobs/`)
3. Runs `go run main.go` — scrapes all companies, deduplicates, classifies roles
4. Writes the public job board to `jobs/README.md`
5. Writes the target companies feed to `scraper/README.md` (this file, below)
6. Commits and pushes both repos
7. Fires a Discord summary embed via webhook

Your machine does not need to be on.

---

## Setup — fork your own

### 1. Fork and create your data repo

Fork this repo. Then create a new public repo (e.g. `yourname/jobs`) — this is where the job board will be published.

In your `jobs` repo, create `local_data/job_ids.json` containing `[]` and a `README.md` with this structure:

```markdown
## Intern & New Grad Opportunities

| Company | Role | Location | Apply | Posted |
| --- | --- | --- | :---: | :---: |
<!-- intern-rows-end -->

## All SWE Opportunities

| Company | Role | Location | Apply | Posted |
| --- | --- | --- | :---: | :---: |
<!-- general-rows-end -->
```

### 2. Set GitHub Actions secrets

In your forked `jobscanner` repo — Settings → Secrets and variables → Actions:

| Secret | What it is |
|---|---|
| `JOBS_REPO_TOKEN` | GitHub PAT with **Contents: Read and Write** on your jobs repo. Fine-grained PATs default to read-only — set write explicitly. |
| `DISCORD_WEBHOOK_URL` | Discord channel webhook URL. Optional — omit to disable notifications. |

### 3. Configure your target companies

Edit `local_data/target_companies.json` — a JSON array of company names:

```json
["Google", "Meta", "Stripe", "Anthropic", "Figma"]
```

The target feed at the bottom of this README updates automatically every CI run.

### 4. Enable scheduled runs

Go to the Actions tab of your forked repo. If you see a banner about workflows being disabled on forks, click **Enable workflows**. The cron runs at the top of every UTC hour.

---

## Running locally

```bash
go run main.go
```

Reads state from `local_data/`, writes to `README.md`. Set env vars to redirect output in CI or to a separate data directory:

```bash
DATA_DIR=/path/to/jobs/local_data README_PATH=/path/to/jobs/README.md go run main.go
```

### Watch mode

Polls on a configurable interval and sends per-job Discord alerts for new intern roles:

```bash
go run main.go --watch --interval=15m
```

### Target company report

Prints a coverage table showing which target companies returned jobs in the last 7 days — useful for spotting broken scrapers:

```bash
go run main.go --target-report
```

---

## Configuration

| Env var | Default | Purpose |
|---|---|---|
| `DATA_DIR` | `local_data` | Directory for `job_ids.json`, `jobs.db`, `jobs.json` |
| `README_PATH` | `README.md` | Where the public intern/SWE tables are written |
| `TARGET_README_PATH` | `README.md` | Where the target companies feed is written |
| `DISCORD_WEBHOOK_URL` | _(empty, disabled)_ | Discord webhook for run summaries and watch-mode alerts |

---

## Project structure

```
go-get-jobs/
├── main.go                      entry point; --watch and --target-report flags
├── common/                      JobPosting struct, HTTP client (resty, 3 retries + backoff)
├── common_const/                env-var-backed path functions
├── database/                    SQLite layer (modernc.org/sqlite, pure Go, no CGO)
│   ├── init.go                  schema: jobs + target_companies tables
│   ├── insert_data.go           INSERT OR IGNORE with role classification at insert time
│   ├── classify.go              intern / new_grad / general by title keyword (word-boundary regex)
│   ├── export.go                writes jobs.json after each run
│   └── target_companies.go      target feed queries and 7-day coverage report
├── process/                     orchestration, dedup via job_ids.json, sync.Once CI cache
├── readme/                      README writers
│   ├── process_readme.go        two-table writer (intern + general)
│   ├── target_section.go        target companies section writer
│   └── date.go                  posting date parsing and display formatting
├── notify/                      Discord webhooks
│   └── discord.go               SendCISummary (CI) + SendWatchAlert (watch mode)
├── sites/                       23 custom scrapers
├── workday/                     30 Workday company configs
├── workday_main/                generic Workday CXS POST scraper
└── local_data/
    ├── target_companies.json    your personal watchlist — edit this
    ├── job_ids.json             dedup state, persists between CI runs
    └── jobs.json                JSON export of all current jobs
```

---

## Known scraper issues

| Company | Issue |
|---|---|
| Microsoft | TLS certificate mismatch on `gcsservices.careers.microsoft.com` |
| Apple | CSRF token required — returns 0 jobs |
| Tesla | 403 Access Denied |
| Splunk | Returns HTML instead of JSON |

Pull requests welcome.

---

## 🎯 My Target Companies

_Updated automatically every hour by CI._

| Company | Role | Location | Apply | Posted |
| --- | --- | --- | :---: | :---: |
<!-- target-rows-end -->
