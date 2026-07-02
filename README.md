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
| **Meta** | Research Scientist Intern, Photorealistic Telepresence (PhD) | Sausalito, CA, Pittsburgh, PA, Redmond, WA | <a href="https://www.metacareers.com/jobs/2022109075207025" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Embedded Contextual AI on Wearables (PhD) | Sunnyvale, CA, Redmond, WA | <a href="https://www.metacareers.com/jobs/2160167211413098" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Monetization Generative AI - LLM (PhD) | Bellevue, WA, Menlo Park, CA, Seattle, WA | <a href="https://www.metacareers.com/jobs/2916726525182155" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Optical System Design (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/1710381673750348" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, State Estimation for Dexterous Manipulation (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/2774289902955470" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, FAIR - Language & Multimodal Foundations (PhD) | Menlo Park, CA, Seattle, WA, New York, NY | <a href="https://www.metacareers.com/jobs/24536664159369645" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Machine Perception for Input and Interaction (PhD) | Redmond, WA, Seattle, WA | <a href="https://www.metacareers.com/jobs/779670167783218" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Applied Vision and Image Quality (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/1422892385992613" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Microsoft** | Critical Environment Technician Intern | Ireland, Dublin, Dublin | <a href="https://apply.careers.microsoft.com/careers/job/1970393556869574" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Microsoft** | Data Center Technicians Intern | Spain, Madrid, Madrid | <a href="https://apply.careers.microsoft.com/careers/job/1970393556873445" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 06 |
| **Microsoft** | Software Engineering INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556911730" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 01 |
| **Microsoft** | Software Engineering INTERN | Brazil, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556875247" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 02 |
| **Microsoft** | Data Science INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556917520" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 30 |
| **Microsoft** | Critical Environment Ops INTERN | Canada, Ontario, Greater Toronto | <a href="https://apply.careers.microsoft.com/careers/job/1970393556866710" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Microsoft** | Government Affairs: Events and Operations Internship Opportunities | Belgium, Brussels Region, Brussels | <a href="https://apply.careers.microsoft.com/careers/job/1970393556887862" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 22 |
| **Microsoft** | Applied Sciences INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556917519" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 30 |
| **Meta** | Research Scientist Intern PhD,  Applied Research | Menlo Park, CA | <a href="https://www.metacareers.com/jobs/2633206137040139" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, AI Alignment | Bellevue, WA, Menlo Park, CA, Seattle, WA, Boston, MA, New York, NY, San Francisco, CA | <a href="https://www.metacareers.com/jobs/1782902493113620" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, NMR Analysis Automation | Redmond, WA | <a href="https://www.metacareers.com/jobs/1418337243438665" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
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
| **Databricks** | PhD GenAI Research Scientist Intern | San Francisco, California | <a href="https://databricks.com/company/careers/open-positions/job?gh_jid=7011263002" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Adobe** | 2026 AI/ML Intern - Machine Learning Engineer/Researcher  Intern | 3 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-AI-ML-Intern---Machine-Learning-Engineer-Intern_R160706" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Adobe** | 2026 Intern - Research Scientist/Engineer | 7 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-Intern---Research-Scientist-Engineer_R160317" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Adobe** | 2026 AI/ML Intern - Machine Learning Engineer | 7 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-AI-ML-Intern---Machine-Learning-Engineer_R158493" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Meta** | Data Engineer, Product Analytics (University Grad) | Bellevue, WA, Menlo Park, CA, New York, NY | <a href="https://www.metacareers.com/jobs/1468691051611430" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Google** |  Network Operations Engineer, University Graduate |  | <a href="https://www.google.com/about/careers/applications/jobs/results/124862995078488774-network-operations-engineer-university-graduate?location=United+States&target_level=EARLY&target_level=INTERN_AND_APPRENTICE&sort_by=date&page=3" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Google** |  Network Operations Residency Program, University Graduate, August 2026 Start |  | <a href="https://www.google.com/about/careers/applications/jobs/results/118981017938600646-network-operations-residency-program-university-graduate-august-2026-start?location=United+States&target_level=EARLY&target_level=INTERN_AND_APPRENTICE&sort_by=date&page=5" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Visiting Researcher, FAIR (University Grad) | Menlo Park, CA | <a href="https://www.metacareers.com/jobs/1546430453541264" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Production Engineer (University Grad) | Bellevue, WA, Menlo Park, CA, New York, NY | <a href="https://www.metacareers.com/jobs/1570473568087646" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Apple** | Software Engineer - Universal Media | United States of America | <a href="https://jobs.apple.com/en-us/details/200670706-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 01 |
| **Apple** | Backend Software Engineer - Universal Media | United States of America | <a href="https://jobs.apple.com/en-us/details/200669674-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Apple** | Computer Vision Software Engineer — Camera Technologies & Systems | United States of America | <a href="https://jobs.apple.com/en-us/details/200658419-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 18 |
| **Apple** | Machine Learning Engineer, Web Indexing Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200668682-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Machine Learning Engineer, Web Indexing Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200668682-3760" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Machine Learning Engineer, Proactive | United States of America | <a href="https://jobs.apple.com/en-us/details/200622577-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 03 |
| **Adobe** | 2026 University Graduate - Machine Learning Engineer | Seattle | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/Seattle/XMLNAME-2026-University-Graduate---Machine-Learning-Engineer_R160133" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
<!-- target-rows-end -->
