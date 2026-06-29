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
| **Meta** | Research Scientist Intern, Machine Perception for Input and Interaction (PhD) | Redmond, WA, Seattle, WA | <a href="https://www.metacareers.com/jobs/779670167783218" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Applied Vision and Image Quality (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/1422892385992613" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Databricks** | PhD GenAI Research Scientist Intern | San Francisco, California | <a href="https://databricks.com/company/careers/open-positions/job?gh_jid=7011263002" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Microsoft** | Software Engineering INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556911730" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 23 |
| **Microsoft** | Data Center Technicians Intern | Netherlands, Noord-Holland, Middenmeer | <a href="https://apply.careers.microsoft.com/careers/job/1970393556867635" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 20 |
| **Microsoft** | IMDA-CLT Program: Solution Engineering INTERN | Singapore, Singapore, Singapore | <a href="https://apply.careers.microsoft.com/careers/job/1970393556869379" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Microsoft** | Critical Environment Ops INTERN | Canada, Ontario, Greater Toronto | <a href="https://apply.careers.microsoft.com/careers/job/1970393556866710" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Microsoft** | Government Affairs: Events and Operations Internship Opportunities | Belgium, Brussels Region, Brussels | <a href="https://apply.careers.microsoft.com/careers/job/1970393556887862" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 22 |
| **Meta** | Research Scientist Intern, Photorealistic Telepresence (PhD) | Sausalito, CA, Pittsburgh, PA, Redmond, WA | <a href="https://www.metacareers.com/jobs/2022109075207025" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Embedded Contextual AI on Wearables (PhD) | Sunnyvale, CA, Redmond, WA | <a href="https://www.metacareers.com/jobs/2160167211413098" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern PhD,  Applied Research | Menlo Park, CA | <a href="https://www.metacareers.com/jobs/2633206137040139" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, AI Alignment | Bellevue, WA, Menlo Park, CA, Seattle, WA, Boston, MA, New York, NY, San Francisco, CA | <a href="https://www.metacareers.com/jobs/1782902493113620" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Monetization Generative AI - LLM (PhD) | Bellevue, WA, Menlo Park, CA, Seattle, WA | <a href="https://www.metacareers.com/jobs/2916726525182155" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Optical System Design (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/1710381673750348" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, State Estimation for Dexterous Manipulation (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/2774289902955470" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, FAIR - Language & Multimodal Foundations (PhD) | Menlo Park, CA, Seattle, WA, New York, NY | <a href="https://www.metacareers.com/jobs/24536664159369645" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Apple** | Software Development Engineer - Data | United States of America | <a href="https://jobs.apple.com/en-us/details/200667248-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 08 |
| **Apple** | Security Software Engineer, OS Security | United States of America | <a href="https://jobs.apple.com/en-us/details/200667231-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 08 |
| **Apple** | Detection and Response Software Engineer | | United States of America | <a href="https://jobs.apple.com/en-us/details/200655126-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 06 |
| **Apple** | Software Engineer (Data Solutions), AI & Data Platforms (AiDP) | United States of America | <a href="https://jobs.apple.com/en-us/details/200667022-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 05 |
| **Apple** | Software Engineer (Data Solutions), AI & Data Platforms (AiDP) | United States of America | <a href="https://jobs.apple.com/en-us/details/200667022-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 05 |
| **Apple** | Cloud Database Engineer - Apple Ads | United States of America | <a href="https://jobs.apple.com/en-us/details/200666675-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 05 |
| **Apple** | Sr. Software Engineer, Core Location, Sensing & Connectivity  | United States of America | <a href="https://jobs.apple.com/en-us/details/200630136-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | Senior Machine Learning Engineer, AI, SIML | United States of America | <a href="https://jobs.apple.com/en-us/details/200651112-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | Senior Machine Learning Engineer, AI, SIML | United States of America | <a href="https://jobs.apple.com/en-us/details/200651112-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | Bluetooth Firmware Engineer, Wireless Technologies & Ecosystems | United States of America | <a href="https://jobs.apple.com/en-us/details/200666184-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | Program Manager - SWE Infrastructure | United States of America | <a href="https://jobs.apple.com/en-us/details/200666799-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | Sr. CI Tooling & Infrastructure Engineer - Xcode | United States of America | <a href="https://jobs.apple.com/en-us/details/200666807-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | Software Engineer - AI Agents & Automation, Maps Data Tooling (iOS) | United States of America | <a href="https://jobs.apple.com/en-us/details/200663946-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | Senior Software Engineer,  Software Developer Foundation | United States of America | <a href="https://jobs.apple.com/en-us/details/200666694-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Apple** | AirPlay Engineering Investigations Lead | United States of America | <a href="https://jobs.apple.com/en-us/details/200660710-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 03 |
| **Apple** | Software Development Engineer - Test, ASE Media Platform Quality - Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200666649-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 03 |
| **Apple** | Senior Compliance and Automation Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200665946-3577" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 03 |
| **Apple** | Software Engineer (Customer Success), Developer Engagement | United States of America | <a href="https://jobs.apple.com/en-us/details/200666610-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 03 |
| **Apple** | Forward Deployed Engineer, Maps Client QE Intelligence | United States of America | <a href="https://jobs.apple.com/en-us/details/200647290-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 02 |
| **Apple** | Security Adoption Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200665938-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 02 |
| **Apple** | Senior Engineering Program Manager, Find My | United States of America | <a href="https://jobs.apple.com/en-us/details/200630263-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 02 |
| **Apple** | Data Analytics Engineer - Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200662203-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 02 |
| **Apple** | Tools and Automation Engineer -- UI Automation | United States of America | <a href="https://jobs.apple.com/en-us/details/200665914-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 01 |
| **Apple** | System Test Engineer, 5G/6G Wireless Interoperability - Wireless Technologies and Ecosystems | United States of America | <a href="https://jobs.apple.com/en-us/details/200665837-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 01 |
| **Apple** | Tools & Automation Engineer - iCloud Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200663792-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 01 |
| **Apple** | Sr Program Manager, Commerce Program Management - Partners, Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200656992-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 29 |
| **Apple** | Full Stack Engineer, FindMy | United States of America | <a href="https://jobs.apple.com/en-us/details/200665699-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 28 |
| **Apple** | Director, Automation & Tools - Camera & Photos | United States of America | <a href="https://jobs.apple.com/en-us/details/200665545-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 28 |
| **Apple** | Systems Architect, Retail and Marcom Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200665367-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 28 |
| **Apple** | Special Projects - Research Scientist/Engineer (Robotics) | United States of America | <a href="https://jobs.apple.com/en-us/details/200663675-3760" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 27 |
| **Apple** | Database Systems SRE, ASE Cassandra SRE | United States of America | <a href="https://jobs.apple.com/en-us/details/200665417-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 27 |
| **Apple** | Sr Application Engineer — IT Inventory Platform, IS&T | United States of America | <a href="https://jobs.apple.com/en-us/details/200664459-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 26 |
| **Apple** | Sr Application Engineer — IT Inventory Platform, IS&T | United States of America | <a href="https://jobs.apple.com/en-us/details/200664459-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 26 |
| **Apple** | Business Development Manager, Apple Card | United States of America | <a href="https://jobs.apple.com/en-us/details/200665240-2459" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 26 |
| **Apple** | Business Development Manager, Apple Card | United States of America | <a href="https://jobs.apple.com/en-us/details/200665240-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 26 |
| **Apple** | Clinical Specialist - Health | United States of America | <a href="https://jobs.apple.com/en-us/details/200665229-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 26 |
| **Apple** | Sr. Software QA Engineer, Siri Client Platforms | United States of America | <a href="https://jobs.apple.com/en-us/details/200664913-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 26 |
| **Apple** | Engineering Program Manager, Quality Assurance - Customer Systems  | United States of America | <a href="https://jobs.apple.com/en-us/details/200663756-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 26 |
| **Apple** | Full Stack Software Engineer - Camera & Photos Tools & AI Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200665015-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 24 |
| **Apple** | Program Manager, IT SOX Compliance – Apple Ads | United States of America | <a href="https://jobs.apple.com/en-us/details/200664954-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Staff AI Software Engineer, Siri User Experiences | United States of America | <a href="https://jobs.apple.com/en-us/details/200664946-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Internationalization Engineering Manager - Intelligence Features | United States of America | <a href="https://jobs.apple.com/en-us/details/200664938-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Machine Learning and Artificial Intelligence Masters Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664221-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Machine Learning and Artificial Intelligence PhD Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664223-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Business, Marketing & Creative Masters Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664247-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Business, Marketing & Creative Undergrad Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664241-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Legal Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664232-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | International Engineering Lead - Intelligence Features | United States of America | <a href="https://jobs.apple.com/en-us/details/200659194-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Principal Product Manager - Claris | United States of America | <a href="https://jobs.apple.com/en-us/details/200643356-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 22 |
| **Apple** | Business Infrastructure Demand Planner | United States of America | <a href="https://jobs.apple.com/en-us/details/200664727-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Business Development & Partnership Manager (Fintech) | United States of America | <a href="https://jobs.apple.com/en-us/details/200664797-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Machine Learning and Artificial Intelligence Undergrad Internships | United States of America | <a href="https://jobs.apple.com/en-us/details/200664780-3810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Senior Data Engineer  | United States of America | <a href="https://jobs.apple.com/en-us/details/200663993-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
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
| **Apple** | Anti Abuse Lead, Mail Operations (Trust & Safety) | United States of America | <a href="https://jobs.apple.com/en-us/details/200646760-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Anti Abuse Lead, Mail Operations (Trust & Safety) | United States of America | <a href="https://jobs.apple.com/en-us/details/200646760-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Engineering Manager, Data Protection - FoundationDB | United States of America | <a href="https://jobs.apple.com/en-us/details/200661187-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | AIML Privacy & Data Governance Eng Lead, Evaluation | United States of America | <a href="https://jobs.apple.com/en-us/details/200664385-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Senior Applied Researcher | United States of America | <a href="https://jobs.apple.com/en-us/details/200663871-3577" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Sr. Manager, Automation & Tools - Camera & Photos | United States of America | <a href="https://jobs.apple.com/en-us/details/200660461-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 21 |
| **Apple** | Software Engineer, Crypto Services - Key Management, Enterprise Technology Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200664495-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 20 |
| **Apple** | DevOps Engineer , Employee Experience & Productivity | United States of America | <a href="https://jobs.apple.com/en-us/details/200664573-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 20 |
| **Apple** | Solution Engineer, G&A Solutions Engineering (GSE) | United States of America | <a href="https://jobs.apple.com/en-us/details/200626160-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 20 |
| **Apple** | Software Engineer, Data Solutions ASE | United States of America | <a href="https://jobs.apple.com/en-us/details/200664338-2459" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 20 |
| **Apple** | Principal AI Architect, App Store Data | United States of America | <a href="https://jobs.apple.com/en-us/details/200664236-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 20 |
| **Apple** | Software Development Engineer - Traffic Network Proxying, ASE | United States of America | <a href="https://jobs.apple.com/en-us/details/200664341-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 19 |
| **Apple** | Technical Product Manager- Issuers (Wallet, Payments & Commerce)  | United States of America | <a href="https://jobs.apple.com/en-us/details/200664301-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 19 |
| **Apple** | Sr. Engineering Program Manager, Security, Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200663998-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 18 |
| **Apple** | Site Reliability Engineer (Edge Services), Infrastructure Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200663929-0953" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 18 |
| **Apple** | Site Reliability Engineer (Edge Services), Infrastructure Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200663929-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 18 |
| **Apple** | Site Reliability Engineer (Edge Services), Infrastructure Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200663929-0776" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 18 |
| **Apple** | Site Reliability Engineer (Edge Services), Infrastructure Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200663929-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 18 |
| **Apple** | Production Services Engineer - Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200663488-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 18 |
| **Apple** | Tools and Automation Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200663616-0365" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 18 |
| **Apple** | Software Engineer iCloud - Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200663794-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 18 |
| **Apple** | Software Engineer, Employee Experience & Productivity | United States of America | <a href="https://jobs.apple.com/en-us/details/200660609-0670" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 15 |
| **Apple** | ASE Compute - Site Reliability Engineering (SRE) Manager | United States of America | <a href="https://jobs.apple.com/en-us/details/200662234-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 14 |
| **Apple** | Manager, Product Management - AI & Data Platforms | United States of America | <a href="https://jobs.apple.com/en-us/details/200662445-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 14 |
| **Apple** | Senior Engineering Manager, Platform Engineering - iCloud | United States of America | <a href="https://jobs.apple.com/en-us/details/200633146-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 14 |
| **Apple** | Architect, Perimeter and Network Security, Enterprise Technology Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200663398-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 14 |
| **Apple** | Software Engineer, Triage Intelligence and Debug Engineering, CoreOS | United States of America | <a href="https://jobs.apple.com/en-us/details/200663192-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 14 |
| **Apple** | Systems Software Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200663092-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 14 |
| **Apple** | Network Engineer, , Apple Cloud Network | United States of America | <a href="https://jobs.apple.com/en-us/details/200662900-3121" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 14 |
| **Apple** | Software Engineer - Capture, Vision Products Software | United States of America | <a href="https://jobs.apple.com/en-us/details/200663107-3401" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 13 |
| **Apple** | Software Development Engineer - iOS for Keyboard | United States of America | <a href="https://jobs.apple.com/en-us/details/200663112-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 13 |
| **Apple** | Senior Software Engineer,  Software Developer Foundation | United States of America | <a href="https://jobs.apple.com/en-us/details/200660873-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 12 |
| **Apple** | Sr Software Engineer (Platform Intelligence), Developer Workflows | United States of America | <a href="https://jobs.apple.com/en-us/details/200662563-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 11 |
| **Apple** | Senior Software Engineer (Observability Solutions) - Enterprise Technology Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200661000-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 11 |
| **Apple** | Senior Full Stack Software Development Engineer - AI, Search & Knowledge | United States of America | <a href="https://jobs.apple.com/en-us/details/200662684-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 11 |
| **Apple** | Engineering Program Manager, AI and Data Platform, Apple Services Engineering  | United States of America | <a href="https://jobs.apple.com/en-us/details/200651225-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 11 |
| **Apple** | Payments Engineering Project Manager (Wallet, Payments & Commerce) | United States of America | <a href="https://jobs.apple.com/en-us/details/200662179-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 11 |
| **Apple** | Senior Software Engineer, Intelligent Automation & Developer Platforms | United States of America | <a href="https://jobs.apple.com/en-us/details/200662429-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 08 |
| **Apple** | Software Engineer - Intelligent Engineering Workflows | United States of America | <a href="https://jobs.apple.com/en-us/details/200655381-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 08 |
| **Apple** | Sr Engineering Manager (Customer Engagement), AI & Data Platforms (AiDP) | United States of America | <a href="https://jobs.apple.com/en-us/details/200662407-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 08 |
| **Apple** | Machine Learning Engineer - Agentic AI | United States of America | <a href="https://jobs.apple.com/en-us/details/200662397-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 08 |
| **Apple** | Sr. Ruby on Rails Engineer, ASE | United States of America | <a href="https://jobs.apple.com/en-us/details/200654205-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 08 |
| **Apple** | Engineering Project Manager (SAP Vertex),  IS&T Enterprise Systems | United States of America | <a href="https://jobs.apple.com/en-us/details/200656402-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 08 |
| **Apple** | Manager, Machine Learning Infrastructure - SIML | United States of America | <a href="https://jobs.apple.com/en-us/details/200661779-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 08 |
| **Apple** | Manager, Machine Learning Infrastructure - SIML | United States of America | <a href="https://jobs.apple.com/en-us/details/200661779-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 08 |
| **Apple** | Senior Software Data Engineer, App Store | United States of America | <a href="https://jobs.apple.com/en-us/details/200662206-3577" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 07 |
| **Apple** | Software Data Engineer, App Store | United States of America | <a href="https://jobs.apple.com/en-us/details/200661956-3577" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 07 |
| **Apple** | Senior iOS Engineering Manager - AI Adoption | United States of America | <a href="https://jobs.apple.com/en-us/details/200624967-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 07 |
| **Apple** | Senior Software Engineer — Observability  | United States of America | <a href="https://jobs.apple.com/en-us/details/200661910-1435" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 07 |
| **Apple** | Senior Software Engineer — Observability  | United States of America | <a href="https://jobs.apple.com/en-us/details/200661910-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 07 |
| **Apple** | Software Automation Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200661716-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 06 |
| **Apple** | Swift Compiler Backend Engineer, Languages & Runtimes | United States of America | <a href="https://jobs.apple.com/en-us/details/200661730-3401" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 05 |
| **Apple** | Staff AI Software Engineer, Siri Core Modeling | United States of America | <a href="https://jobs.apple.com/en-us/details/200661622-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 05 |
| **Apple** | SWE Engineering Project Manager, International Experience & Operations Features | United States of America | <a href="https://jobs.apple.com/en-us/details/200649007-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 05 |
| **Apple** | Software Engineer - Cloud Infrastructure, Golang | United States of America | <a href="https://jobs.apple.com/en-us/details/200661492" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 05 |
| **Apple** | Staff Applied Scientist, AI Quality & Meta Evaluation  | United States of America | <a href="https://jobs.apple.com/en-us/details/200661039-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 04 |
| **Apple** | Applied AI Engineer - iCloud Data | United States of America | <a href="https://jobs.apple.com/en-us/details/200661135-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 04 |
| **Apple** | Applied AI Engineer - iCloud Data | United States of America | <a href="https://jobs.apple.com/en-us/details/200661135-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 04 |
| **Apple** | Engineering Manager, App Store Data | United States of America | <a href="https://jobs.apple.com/en-us/details/200661200-3577" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 01 |
| **Apple** | Tooling & Infrastructure Engineer, Test Engineering & Efficiency | United States of America | <a href="https://jobs.apple.com/en-us/details/200652988-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | May 01 |
| **Apple** | Senior Full Stack Engineer - Special Projects | United States of America | <a href="https://jobs.apple.com/en-us/details/200660601-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 30 |
| **Apple** | System Integration Lead | United States of America | <a href="https://jobs.apple.com/en-us/details/200660901-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 30 |
| **Apple** | Engineering Program Manager, Search Quality and Infrastructure, Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200660943-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 30 |
| **Apple** | Engineering Program Manager, Search Quality and Infrastructure, Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200660943-3760" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 30 |
| **Apple** | Senior Product Manager, Find My | United States of America | <a href="https://jobs.apple.com/en-us/details/200660914-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 30 |
| **Apple** | Engineering Project Manager - AI Features Internationalization, L&RE | United States of America | <a href="https://jobs.apple.com/en-us/details/200660570-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 30 |
| **Apple** | Wallet and Payments Engineering Program Manager | United States of America | <a href="https://jobs.apple.com/en-us/details/200670177-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Apple** | Software Engineer, Quality, Retail and Marcom Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200670205-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Apple** | Senior ML Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200670136-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Apple** | On-Device ML Integration Engineer, Graphics, Games and Machine Learning | United States of America | <a href="https://jobs.apple.com/en-us/details/200642883-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Apple** | Head of iCloud Business Strategy and Growth | United States of America | <a href="https://jobs.apple.com/en-us/details/200670018-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Apple** | Senior Machine Learning Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200669181-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Apple** | Health Software QA Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200670041-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Apple** | Software Development Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200669105-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Apple** | Senior Software Development Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200668915-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Apple** | Data Scientist | United States of America | <a href="https://jobs.apple.com/en-us/details/200668804-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Apple** | Software Engineer - Applications | United States of America | <a href="https://jobs.apple.com/en-us/details/200669180-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Apple** | Software Engineer- Systems | United States of America | <a href="https://jobs.apple.com/en-us/details/200669153-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 25 |
| **Apple** | Software Engineer, SDLC Analytics | United States of America | <a href="https://jobs.apple.com/en-us/details/200650958-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 24 |
| **Apple** | Data Scientist - Strategic Security Risk | United States of America | <a href="https://jobs.apple.com/en-us/details/200669799-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 24 |
| **Apple** | Research Operations Engineer (Applied Sensing and Health), Sensing & Connectivity | United States of America | <a href="https://jobs.apple.com/en-us/details/200668356-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 24 |
| **Apple** | Engineering Program Manager, Foundational Security, Apple Services Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200669235-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 23 |
| **Apple** | Software Application Support Engineer, Retail & Marcom Engineering  | United States of America | <a href="https://jobs.apple.com/en-us/details/200668353-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 23 |
| **Apple** | DataServices Engineer - Apple Wallet, Payments & Commerce | United States of America | <a href="https://jobs.apple.com/en-us/details/200669652-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 23 |
| **Apple** | DataServices Engineer - Apple Wallet, Payments & Commerce | United States of America | <a href="https://jobs.apple.com/en-us/details/200669652-1435" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 23 |
| **Apple** | Senior Machine Learning Engineer, Developer Product Analytics | United States of America | <a href="https://jobs.apple.com/en-us/details/200668686-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 23 |
| **Apple** | Technical Product Manager - eCommerce & PSPs (Wallet, Payments & Commerce) | United States of America | <a href="https://jobs.apple.com/en-us/details/200669597-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 23 |
| **Apple** | Senior Device Fleet and Infrastructure Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200669467-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 22 |
| **Apple** | Clinical Study Operations Engineering Program Manager - Health | United States of America | <a href="https://jobs.apple.com/en-us/details/200668430-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 22 |
| **Apple** |  Senior Audio Systems Architect | United States of America | <a href="https://jobs.apple.com/en-us/details/200668256-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 22 |
| **Apple** | Manager, Memory & Symbolication Tools | United States of America | <a href="https://jobs.apple.com/en-us/details/200669201-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 19 |
| **Apple** | Sr. ML Production Model Automation Engineer, Siri Speech | United States of America | <a href="https://jobs.apple.com/en-us/details/200669222-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 18 |
| **Apple** | Software Development Engineer - Frameworks & AI/ML, Wireless Technologies & Ecosystems | United States of America | <a href="https://jobs.apple.com/en-us/details/200646833-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 18 |
| **Apple** | Program Manager, Employee Experience & Productivity | United States of America | <a href="https://jobs.apple.com/en-us/details/200669164-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 18 |
| **Apple** | Senior CoreOS Engineer - AppleCare Enterprise Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200659687-0240" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 17 |
| **Apple** | Senior CoreOS Engineer - AppleCare Enterprise Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200659687-3401" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 17 |
| **Apple** | Engineering Project Manager, Sales Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200649813-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 17 |
| **Apple** | DevOps Engineer, G&A Solutions Engineering (GSE) | United States of America | <a href="https://jobs.apple.com/en-us/details/200667899-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Data Scientist, Employee Experience & Productivity | United States of America | <a href="https://jobs.apple.com/en-us/details/200668306-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Senior Cloud Engineer (Wallet, Payments and Commerce)  | United States of America | <a href="https://jobs.apple.com/en-us/details/200668294-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 15 |
| **Apple** | Senior Cloud Engineer (Wallet, Payments and Commerce)  | United States of America | <a href="https://jobs.apple.com/en-us/details/200668294-1435" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 15 |
| **Apple** | Senior Cloud Engineer (Wallet, Payments and Commerce)  | United States of America | <a href="https://jobs.apple.com/en-us/details/200668294-2459" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 15 |
| **Apple** | Engineering Project Manager (SAP FI - Financial Accounting),  IS&T Enterprise Systems | United States of America | <a href="https://jobs.apple.com/en-us/details/200651802-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 15 |
| **Apple** | Certification Engineering Program Manager, Wireless Technologies & Ecosystems | United States of America | <a href="https://jobs.apple.com/en-us/details/200653574-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 15 |
| **Apple** | Partner and Performance Program Manager | United States of America | <a href="https://jobs.apple.com/en-us/details/200661700-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 15 |
| **Apple** | Partner and Performance Program Manager | United States of America | <a href="https://jobs.apple.com/en-us/details/200661700-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 15 |
| **Apple** | Strategic Partnerships Lead, Corporate Development  | United States of America | <a href="https://jobs.apple.com/en-us/details/200646043-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 13 |
| **Apple** | Platform Security Certification Engineer | United States of America | <a href="https://jobs.apple.com/en-us/details/200668286-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 12 |
| **Apple** | Senior DevOps Engineer, Infrastructure Services | United States of America | <a href="https://jobs.apple.com/en-us/details/200666781-3121" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 12 |
| **Apple** | Bluetooth Firmware Engineer, Wireless Technologies & Ecosystems | United States of America | <a href="https://jobs.apple.com/en-us/details/200667156-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 12 |
| **Apple** | Offensive Security Researcher, Kernel & Embedded Security | United States of America | <a href="https://jobs.apple.com/en-us/details/200667546-2459" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 11 |
| **Apple** | Engineering Leader, Data Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200667770-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 11 |
| **Apple** | Senior Data Engineer — Music Services Operations Analytics & Strategy | United States of America | <a href="https://jobs.apple.com/en-us/details/200667098-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 11 |
| **Apple** | Applied AI Engineer (EDA), Platform Architecture | United States of America | <a href="https://jobs.apple.com/en-us/details/200667745-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 10 |
| **Apple** | Applied AI Engineer (EDA), Platform Architecture | United States of America | <a href="https://jobs.apple.com/en-us/details/200667745-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 10 |
| **Apple** | Applied AI Engineer (EDA), Platform Architecture | United States of America | <a href="https://jobs.apple.com/en-us/details/200667797-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 10 |
| **Apple** | Applied AI Engineer (EDA), Platform Architecture | United States of America | <a href="https://jobs.apple.com/en-us/details/200667797-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 10 |
| **Apple** | Manager, Engineering Project Management, Build Health and Availability Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200667594-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 10 |
| **Apple** | Software Engineer, Information Systems & Technology | United States of America | <a href="https://jobs.apple.com/en-us/details/200653252-3956" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 10 |
| **Apple** | Software Engineer, Information Systems & Technology | United States of America | <a href="https://jobs.apple.com/en-us/details/200653729-0157" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 10 |
| **Apple** | Generative AI Applied Scientist, SIML - ISE | United States of America | <a href="https://jobs.apple.com/en-us/details/200632699-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 09 |
| **Apple** | Systems Engineer - Evaluation Engineering | United States of America | <a href="https://jobs.apple.com/en-us/details/200667461-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 09 |
| **Apple** | Sr. Software Engineer - Apple Vision Pro Developer Ecosystem, Vision Products Group | United States of America | <a href="https://jobs.apple.com/en-us/details/200659830-3401" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 08 |
| **Adobe** | 2026 AI/ML Intern - Machine Learning Engineer/Researcher  Intern | 3 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-AI-ML-Intern---Machine-Learning-Engineer-Intern_R160706" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Adobe** | 2026 Intern - Research Scientist/Engineer | 7 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-Intern---Research-Scientist-Engineer_R160317" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Adobe** | 2026 AI/ML Intern - Machine Learning Engineer | 7 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-AI-ML-Intern---Machine-Learning-Engineer_R158493" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Google** |  Network Operations Residency Program, University Graduate, August 2026 Start |  | <a href="https://www.google.com/about/careers/applications/jobs/results/118981017938600646-network-operations-residency-program-university-graduate-august-2026-start?location=United+States&target_level=EARLY&target_level=INTERN_AND_APPRENTICE&sort_by=date&page=4" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Network Production Engineer (University Grad) | Menlo Park, CA | <a href="https://www.metacareers.com/jobs/4416232018698029" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Data Engineer, Product Analytics (University Grad) | Bellevue, WA, Menlo Park, CA, New York, NY | <a href="https://www.metacareers.com/jobs/1468691051611430" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Production Engineer (University Grad) | Bellevue, WA, Menlo Park, CA, New York, NY | <a href="https://www.metacareers.com/jobs/1570473568087646" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Visiting Researcher, FAIR (University Grad) | Menlo Park, CA | <a href="https://www.metacareers.com/jobs/1546430453541264" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Apple** | Backend Software Engineer - Universal Media | United States of America | <a href="https://jobs.apple.com/en-us/details/200669674-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Apple** | Computer Vision Software Engineer — Camera Technologies & Systems | United States of America | <a href="https://jobs.apple.com/en-us/details/200658419-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 18 |
| **Apple** | Machine Learning Engineer, Web Indexing Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200668682-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Machine Learning Engineer, Web Indexing Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200668682-3760" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Machine Learning Engineer, Proactive | United States of America | <a href="https://jobs.apple.com/en-us/details/200622577-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 03 |
| **Adobe** | 2026 University Graduate - Machine Learning Engineer | Seattle | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/Seattle/XMLNAME-2026-University-Graduate---Machine-Learning-Engineer_R160133" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
<!-- target-rows-end -->
