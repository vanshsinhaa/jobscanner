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
| **Databricks** | PhD GenAI Research Scientist Intern | San Francisco, California | <a href="https://databricks.com/company/careers/open-positions/job?gh_jid=7011263002" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Pinterest** | PhD Fall Machine Learning Intern (ATG — Visual, Multimodal, and Recommender Systems) | San Francisco, CA, US; Palo Alto, CA, US; Seattle, WA, US; New York, NY, US | <a href="https://www.pinterestcareers.com/jobs/?gh_jid=7255640" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | Machine Learning and Artificial Intelligence Masters Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664221-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Machine Learning and Artificial Intelligence PhD Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664223-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Business, Marketing & Creative Masters Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664247-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Business, Marketing & Creative Undergrad Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664241-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Legal Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664232-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Machine Learning and Artificial Intelligence Undergrad Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664780-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Hardware PhD Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664421-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Hardware Masters Engineering Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664419-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Hardware Undergrad Engineering Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200663981-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Hardware Technologies PhD Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664414-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Hardware Technologies Masters Engineering Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664383-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Hardware Technologies Undergrad Engineering Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200663968-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Engineering Program Management Undergrad Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664330-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Engineering Program Management Masters Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664336-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Software PhD Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664323-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Software Engineering Masters Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664320-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Software Undergrad Engineering Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664785-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Operations Manufacturing Design Undergrad Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664002-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Operations Manufacturing Design PhD Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664003-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Meta** | Research Scientist Intern PhD,  Applied Research | Menlo Park, CA | <a href="https://www.metacareers.com/jobs/2633206137040139" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, AI Alignment | Bellevue, WA, Menlo Park, CA, Seattle, WA, Boston, MA, New York, NY, San Francisco, CA | <a href="https://www.metacareers.com/jobs/1782902493113620" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, NMR Analysis Automation | Redmond, WA | <a href="https://www.metacareers.com/jobs/1418337243438665" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Photorealistic Telepresence (PhD) | Sausalito, CA, Pittsburgh, PA, Redmond, WA | <a href="https://www.metacareers.com/jobs/2022109075207025" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Embedded Contextual AI on Wearables (PhD) | Sunnyvale, CA, Redmond, WA | <a href="https://www.metacareers.com/jobs/2160167211413098" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Monetization Generative AI - LLM (PhD) | Bellevue, WA, Menlo Park, CA, Seattle, WA | <a href="https://www.metacareers.com/jobs/2916726525182155" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Optical System Design (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/1710381673750348" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, State Estimation for Dexterous Manipulation (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/2774289902955470" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, FAIR - Language & Multimodal Foundations (PhD) | Menlo Park, CA, Seattle, WA, New York, NY | <a href="https://www.metacareers.com/jobs/24536664159369645" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Machine Perception for Input and Interaction (PhD) | Redmond, WA, Seattle, WA | <a href="https://www.metacareers.com/jobs/779670167783218" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Applied Vision and Image Quality (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/1422892385992613" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Airbnb** | Legal Intern, Brazil | São Paulo, Brazil | <a href="https://careers.airbnb.com/positions/7886515?gh_jid=7886515" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 04 |
| **Airbnb** | Sales Operations Intern, Italy | Milan, Italy | <a href="https://careers.airbnb.com/positions/7979270?gh_jid=7979270" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 05 |
| **Palantir** | Deployment Strategist, Internship | Paris, France | <a href="https://jobs.lever.co/palantir/774cf5c9-bf6a-4d77-bf60-d50ef1beb1a0" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Deployment Strategist, Internship - US Government | Honolulu, HI | <a href="https://jobs.lever.co/palantir/a49d4181-a289-435a-b581-7f5af0497c8e" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship | Paris, France | <a href="https://jobs.lever.co/palantir/1b6f1d82-d459-4dea-8bc2-8d2ffe6f881a" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - AUS Government | Sydney, Australia | <a href="https://jobs.lever.co/palantir/395a4483-fc3d-4b77-a500-501923fd0976" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - Commercial | New York, NY | <a href="https://jobs.lever.co/palantir/4d29249a-d7e8-4c39-880d-3b35d7b2f6f6" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - Commercial | Chicago, IL | <a href="https://jobs.lever.co/palantir/d5486403-c050-4920-b2e0-91b69b61ebb2" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - Defense Tech | Washington, D.C. | <a href="https://jobs.lever.co/palantir/cccfe1bd-f15b-4fe5-b044-c793e7961c1b" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - France | New York, NY | <a href="https://jobs.lever.co/palantir/ac0dc094-2480-43c2-8495-26ade227ff4f" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - Intel | Washington, D.C. | <a href="https://jobs.lever.co/palantir/9e40d77f-b07c-437b-98e7-def9b0184d89" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - Poland | New York, NY | <a href="https://jobs.lever.co/palantir/d582cd84-14fd-4aa3-b413-15982d286bd9" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - US Government | Honolulu, HI | <a href="https://jobs.lever.co/palantir/315f695d-04d1-4a9a-848e-cb2bec7a997e" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - US Government | New York, NY | <a href="https://jobs.lever.co/palantir/e0010393-c300-446f-bf67-fa2ef067f16f" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, Internship - US Government | Washington, D.C. | <a href="https://jobs.lever.co/palantir/e6ff8bf2-135e-474d-ad37-24f490ae1dd2" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Privacy and Civil Liberties Software Engineer, Internship | New York, NY | <a href="https://jobs.lever.co/palantir/09846827-b931-4a9f-bd64-c3bb8860187b" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship | Denver, CO | <a href="https://jobs.lever.co/palantir/373eb939-6f57-4836-8479-be79a5e07249" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship | New York, NY | <a href="https://jobs.lever.co/palantir/7d69cf8a-06fd-4f05-bd84-27149db29c4d" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship | Washington, D.C. | <a href="https://jobs.lever.co/palantir/bdcfb29f-4f27-42de-933f-7f83a359b9f0" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship | Palo Alto, CA | <a href="https://jobs.lever.co/palantir/e27af7ab-41fc-40c9-b31d-02c6cb1c505c" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship - Defense Tech | New York, NY | <a href="https://jobs.lever.co/palantir/8bcf4f33-0a79-4248-bbfd-49ac4be9dd8e" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship - Defense Tech | Palo Alto, CA | <a href="https://jobs.lever.co/palantir/a483f41b-0da9-42ea-8ed6-cbf6eb93cc6d" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship - Defense Tech | Washington, D.C. | <a href="https://jobs.lever.co/palantir/f17e98d0-046a-4e6e-9d65-ed0b12dd0ff7" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship - Infrastructure | New York, NY | <a href="https://jobs.lever.co/palantir/b229baac-494b-4a0d-9a13-2e38806e06f3" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship - Infrastructure | Palo Alto, CA | <a href="https://jobs.lever.co/palantir/f221738b-e97c-4ce3-a12a-17ada2b855e4" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship - Production Infrastructure | Seattle, WA | <a href="https://jobs.lever.co/palantir/373367a9-3160-49d8-b7af-2efec062fad1" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship - Production Infrastructure | New York, NY | <a href="https://jobs.lever.co/palantir/37964982-9b4c-471e-a1d8-fb8f45d7f116" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, Internship - Production Infrastructure | Washington, D.C. | <a href="https://jobs.lever.co/palantir/3ab9e715-1ea9-4c6c-ad50-7340eac14e86" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Year at Palantir - Forward Deployed Software Engineer, Internship - Commercial | Chicago, IL | <a href="https://jobs.lever.co/palantir/75cc1c09-8ebd-44c8-b3bc-d122cd1fecb3" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Year at Palantir - Forward Deployed Software Engineer, Internship - Commercial | New York, NY | <a href="https://jobs.lever.co/palantir/e6789b17-62fb-4226-a079-f8c17ff19e2d" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Year at Palantir - Forward Deployed Software Engineer, Internship - USG | Washington, D.C. | <a href="https://jobs.lever.co/palantir/5c4c65c5-77da-4d36-856c-4ade87631019" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Year at Palantir - Forward Deployed Software Engineer, Internship - USG | New York, NY | <a href="https://jobs.lever.co/palantir/5c7bb70c-83ea-43e7-8055-0c8f319f4333" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Year at Palantir - Software Engineer, Internship | New York, NY | <a href="https://jobs.lever.co/palantir/655f9937-a4ce-4e7d-80e2-a6659af07329" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Notion** | Software Engineer Intern (Fall 2026) | San Francisco, California | <a href="https://jobs.ashbyhq.com/notion/5b15697c-fa91-4511-9482-c98a6ff29f90" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 06 |
| **Instacart** | Machine Learning Engineer, PhD Intern (Fall) | United States - Remote | <a href="https://instacart.careers/job/?gh_jid=5917202" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 24 |
| **Instacart** | Machine Learning PhD Intern, Economics (Fall) | United States - Remote | <a href="https://instacart.careers/job/?gh_jid=7532267" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 24 |
| **Microsoft** | Critical Environment Technician Intern | Ireland, Dublin, Dublin | <a href="https://apply.careers.microsoft.com/careers/job/1970393556869574" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Microsoft** | Data Center Technicians Intern | Spain, Madrid, Madrid | <a href="https://apply.careers.microsoft.com/careers/job/1970393556873445" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 06 |
| **Microsoft** | Critical Environment Ops INTERN | Canada, Ontario, Greater Toronto | <a href="https://apply.careers.microsoft.com/careers/job/1970393556866710" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Microsoft** | Applied Sciences INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556917519" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 30 |
| **Microsoft** | Data Science INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556917520" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 30 |
| **Microsoft** | Software Engineering INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556911730" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 01 |
| **Microsoft** | Government Affairs: Events and Operations Internship Opportunities | Belgium, Brussels Region, Brussels | <a href="https://apply.careers.microsoft.com/careers/job/1970393556887862" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 22 |
| **Adobe** | 2026 AI/ML Intern - Machine Learning Engineer/Researcher  Intern | 3 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-AI-ML-Intern---Machine-Learning-Engineer-Intern_R160706" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Adobe** | 2026 Intern - Research Scientist/Engineer | 7 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-Intern---Research-Scientist-Engineer_R160317" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Adobe** | 2026 AI/ML Intern - Machine Learning Engineer | 7 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-AI-ML-Intern---Machine-Learning-Engineer_R158493" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Stripe** | Software Engineer, Intern | Sydney, Australia | <a href="https://stripe.com/jobs/search?gh_jid=7532256" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Square** | Applied Research Intern, Proactive Intelligence & Customer World Models (PhD / Graduate Co-op) | Bay Area, CA, United States of America | <a href="http://block.xyz/careers/jobs/5108007008?gh_jid=5108007008" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 03 |
| **Square** | Applied Research Intern, Proactive Intelligence & Customer World Models (PhD / Graduate Co-op) | Toronto, Ontario , Canada | <a href="http://block.xyz/careers/jobs/5108009008?gh_jid=5108009008" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 03 |
| **Apple** | Software Engineer - Universal Media | United States of America | <a href="https://jobs.apple.com/en-us/details/200670706-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 01 |
| **Apple** | Backend Software Engineer - Universal Media | United States of America | <a href="https://jobs.apple.com/en-us/details/200669674-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Apple** | Computer Vision Software Engineer — Camera Technologies & Systems | United States of America | <a href="https://jobs.apple.com/en-us/details/200658419-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 18 |
| **Apple** | Machine Learning Engineer, Web Indexing Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200668682-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Machine Learning Engineer, Web Indexing Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200668682-3760" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Machine Learning Engineer, Proactive | United States of America | <a href="https://jobs.apple.com/en-us/details/200622577-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 03 |
| **Robinhood** | Customer Experience Associate (New Grad) | Westlake, TX | <a href="https://boards.greenhouse.io/robinhood/jobs/8024530?t=gh_src=&gh_jid=8024530" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 02 |
| **Meta** | Visiting Researcher, FAIR (University Grad) | Menlo Park, CA | <a href="https://www.metacareers.com/jobs/1546430453541264" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Production Engineer (University Grad) | Bellevue, WA, Menlo Park, CA, New York, NY | <a href="https://www.metacareers.com/jobs/1570473568087646" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Data Engineer, Product Analytics (University Grad) | Bellevue, WA, Menlo Park, CA, New York, NY | <a href="https://www.metacareers.com/jobs/1468691051611430" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Deployment Strategist, New Grad - Intel, US Government | Washington, D.C. | <a href="https://jobs.lever.co/palantir/5d8286d6-992a-404b-94af-99c173d40299" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, New Grad - Commercial | New York, NY | <a href="https://jobs.lever.co/palantir/2e6b0ac8-83e9-4be5-a3aa-cf319f751728" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, New Grad - Commercial | Chicago, IL | <a href="https://jobs.lever.co/palantir/e500bcf3-19d8-4d3c-b340-4d76e4a55b40" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, New Grad - Intel, US Government | Washington, D.C. | <a href="https://jobs.lever.co/palantir/fbca0358-083a-4222-bdbb-3bd729b48382" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, New Grad - US Government | Washington, D.C. | <a href="https://jobs.lever.co/palantir/cbe90327-3e6e-451c-a54c-1d3cbcef5aeb" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Forward Deployed Software Engineer, New Grad - US Government | New York, NY | <a href="https://jobs.lever.co/palantir/d1ac83d0-e923-42a5-8e6d-58dd0cab25ca" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Privacy & Civil Liberties Engineer - New Grad | New York, NY | <a href="https://jobs.lever.co/palantir/95e0d2b0-437a-4096-a5c6-0f247f426c90" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad | New York, NY | <a href="https://jobs.lever.co/palantir/94984771-0704-446c-88c6-91ce748f6d92" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad | Denver, CO | <a href="https://jobs.lever.co/palantir/c34b424e-caf2-455a-b104-ae1096ccca29" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad - Defense | New York, NY | <a href="https://jobs.lever.co/palantir/0a838e66-1ab0-4fc4-b4d3-4671c0352278" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad - Defense | Washington, D.C. | <a href="https://jobs.lever.co/palantir/18d901fc-93bb-4d18-9f04-c72031e20d79" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad - Defense | Palo Alto, CA | <a href="https://jobs.lever.co/palantir/f362d7aa-360d-4059-ab38-f482742693b3" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad - Infrastructure | New York, NY | <a href="https://jobs.lever.co/palantir/4abf26b4-795c-420a-bf22-1ab98db268b4" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad - Infrastructure | Palo Alto, CA | <a href="https://jobs.lever.co/palantir/7d75bed5-45d8-4876-840a-2d92ea79c98d" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad - Production Infrastructure | Washington, D.C. | <a href="https://jobs.lever.co/palantir/15844944-fb69-4b57-9531-e988650b20c6" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad - Production Infrastructure | Seattle, WA | <a href="https://jobs.lever.co/palantir/4d5a144e-87ea-45e2-a68c-3fad590629af" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Palantir** | Software Engineer, New Grad - Production Infrastructure | New York, NY | <a href="https://jobs.lever.co/palantir/e1a6c138-98bf-45e2-97f7-2c70371cc38a" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Notion** | Software Engineer, New Grad (AI) | San Francisco, California | <a href="https://jobs.ashbyhq.com/notion/7e6dc7fe-7ddd-42c1-8928-13f7bddb9ec9" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 27 |
| **Notion** | Software Engineer, New Grad | San Francisco, California | <a href="https://jobs.ashbyhq.com/notion/a6311f97-4850-4674-a5f3-d9fe5f6f2555" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 23 |
| **Snowflake** | AI Research Scientist, New Grad – Agents & Reinforcement Learning | Bellevue, Washington, United States | <a href="https://jobs.ashbyhq.com/snowflake/1bad12df-f443-426f-9d09-e96fc780d698" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Adobe** | 2026 University Graduate - Machine Learning Engineer | Seattle | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/Seattle/XMLNAME-2026-University-Graduate---Machine-Learning-Engineer_R160133" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Stripe** | Operations Associate, New Grad (Mexico) | Mexico City, Mexico | <a href="https://stripe.com/jobs/search?gh_jid=7544547" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Stripe** | Software Engineer, New Grad, Developer & End User Experience Platform | Toronto | <a href="https://stripe.com/jobs/search?gh_jid=7991718" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Stripe** | Tech Operations Associate, New Grad (Mexico) | Mexico City, Mexico | <a href="https://stripe.com/jobs/search?gh_jid=7718947" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Google** | Network Operations Engineer, University Graduate |  | <a href="https://www.google.com/about/careers/applications/jobs/results/124862995078488774-network-operations-engineer-university-graduate?location=United+States&target_level=EARLY&target_level=INTERN_AND_APPRENTICE&sort_by=date&page=3" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Google** | Network Operations Residency Program, University Graduate, August 2026 Start |  | <a href="https://www.google.com/about/careers/applications/jobs/results/118981017938600646-network-operations-residency-program-university-graduate-august-2026-start?location=United+States&target_level=EARLY&target_level=INTERN_AND_APPRENTICE&sort_by=date&page=5" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
<!-- target-rows-end -->
