# jobscanner

[![Scrape Jobs](https://github.com/vanshsinhaa/jobscanner/actions/workflows/scrape.yml/badge.svg)](https://github.com/vanshsinhaa/jobscanner/actions/workflows/scrape.yml)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**A self-hosting job board.** Scrapes 65+ tech companies every hour on free GitHub Actions, classifies every posting (intern / new grad / general SWE), and publishes a live, deduplicated job board as a GitHub README — plus a personal watchlist feed for the companies *you* care about. No server. No database hosting. No manual work.

**Live board → [vanshsinhaa/jobs](https://github.com/vanshsinhaa/jobs)** · **Personal feed → [bottom of this README](#-my-target-companies)**

Built on top of [go-get-jobs](https://github.com/neyaadeez/go-get-jobs) with a full rewrite: SQLite backend, GitHub Actions CI, generic ATS engines, Discord webhooks, role classification, recency sorting, and target-company tracking.

---

## Why this exists

Job boards aggregate slowly and bury intern/new-grad roles under senior listings. Company career pages are fast but there are dozens of them. jobscanner scrapes the source directly — the actual ATS APIs behind each careers page — every hour, so postings show up on your board minutes after they go live, already classified and sorted newest-first.

---

## Features

- **65+ companies, 5 scraper engines** — generic engines for **Greenhouse, Ashby, Lever, and Workday** (adding a company on those platforms is a ~5 line change), plus custom scrapers for Oracle Cloud HCM, Eightfold, Phenom, iCIMS, and bespoke career sites (Google, Apple, Amazon, Meta, …)
- **Hourly CI scraping** — GitHub Actions cron; your machine never needs to be on
- **Intern / new-grad classification** — word-boundary regex on titles (`intern`, `co-op`, `new grad`, `entry level`, `Class of 2027`, `2027 Grads`, …) with false-positive guards ("International" ≠ intern)
- **Cycle-aware** — recruiting-cycle years are parsed from titles: postings for past cycles ("Summer 2026 Intern" still up during the 2027 cycle) are aged off the board automatically, so each new hiring season starts clean
- **Target-company watchlist** — a JSON list of companies you care about; their intern/new-grad roles get their own feed, with brand aliasing (`Amex` → American Express, `Twitter` → xAI, `Trello` → Atlassian) and sub-brand detection (Annapurna Labs inside Amazon, Slack inside Salesforce)
- **Deduplication** — `job_ids.json` persists across runs; every job appears exactly once, ever
- **Recency windows** — general roles age out after 14 days, intern/new-grad after 60 (programs post early)
- **Discord notifications** — summary embed per CI run; per-job alerts in watch mode
- **JSON export** — `jobs.json` published alongside the board for programmatic use
- **Diagnostic harness** — `go run ./cmd/baseline` tests every scraper in isolation (no DB/README writes) and prints a per-company health report

---

## Architecture

```
┌─────────────────────────────┐        ┌──────────────────────────────┐
│  vanshsinhaa/jobscanner     │        │  vanshsinhaa/jobs            │
│  (this repo)                │  CI    │  (public output)             │
│                             │ ─────► │                              │
│  · scraper code (Go)        │ hourly │  · README.md   ← job board   │
│  · CI workflow              │        │  · jobs.json   ← API-ish     │
│  · target watchlist config  │        │  · job_ids.json ← dedup      │
│  · target feed (README)     │        │                              │
└─────────────────────────────┘        └──────────────────────────────┘
```

**One CI run, start to finish:**

1. GitHub spins up a fresh Ubuntu runner at the top of every hour
2. Checks out this repo (`scraper/`) and the data repo (`jobs/`)
3. `go run main.go` — every scraper engine runs; results land in a per-run SQLite DB
4. Titles are classified (`intern` / `new_grad` / `general`) at insert time
5. Dedup against `job_ids.json`; sub-brands retagged; target aliases resolved
6. Public board written to `jobs/README.md` (two tables, newest first, row caps)
7. Target feed written to this repo's README (the section at the bottom of this page)
8. Both repos committed and pushed; Discord summary fired

**Scraper engine dispatch:**

```
process.ScrapeAllJobs()
├── workday_main/    generic Workday CXS engine ── 28 companies (payload registry)
└── sites_main/      per-company dispatch
    ├── greenhouse   generic ─ Stripe, Anthropic, Coinbase, Figma, xAI, +9 more
    ├── ashby        generic ─ OpenAI, Notion
    ├── lever        generic ─ Palantir
    └── custom       Google, Apple, Amazon, Meta, Microsoft (Eightfold PCSX),
                     Amex (Oracle HCM), Snowflake (Phenom), Atlassian (iCIMS),
                     Shopify (sitemap), Visa (SmartRecruiters), …
```

---

## Quick start — fork your own job board

### 🤖 Option A: make Claude do all of it

Have an agentic coding tool ([Claude Code](https://claude.com/claude-code), Cursor, etc.) with the `gh` CLI authenticated? Paste this prompt into a fresh session and it will run the entire setup below — fork, data repo, CI rewiring, secrets, first run, verification:

````text
Set up my own self-hosting job board from https://github.com/vanshsinhaa/jobscanner.
Assume `gh` CLI (authenticated as me), git, and Go are installed; check all three
first and stop to tell me if one is missing.

1. Ask me two questions before touching anything: (a) a name for my public data
   repo (default "jobs"), (b) which companies I want on my personal watchlist
   (default: keep the list already in local_data/target_companies.json).
2. Fork vanshsinhaa/jobscanner to my account with gh and clone my fork.
3. Create the public data repo under my account. Commit to it:
   - local_data/job_ids.json containing exactly: []
   - README.md containing the two-table skeleton copied VERBATIM from my fork's
     CONTRIBUTING.md, section "Data repo skeleton" (headers, separator rows, and
     HTML anchor comments must match exactly — the table writers pattern-match them).
4. In my fork, edit .github/workflows/scrape.yml: change
   "repository: vanshsinhaa/jobs" to my data repo. Update
   local_data/target_companies.json to my watchlist from step 1.
5. Secrets — you cannot create GitHub PATs, so walk me through it: I create a
   fine-grained PAT at https://github.com/settings/personal-access-tokens/new
   (repository access: ONLY my data repo; permissions: Contents = Read and write).
   When I paste it to you, run: gh secret set JOBS_REPO_TOKEN --repo <me>/jobscanner
   Then ask if I want Discord notifications; if yes, set DISCORD_WEBHOOK_URL the
   same way, otherwise skip.
6. Sanity-check the scrapers locally: go build ./... then go run ./cmd/baseline.
   A few permanently blocked companies (Tesla, LinkedIn) are expected — anything
   else in the ZERO OR ERROR section, investigate per plan/tracker.md before continuing.
7. Commit and push my fork's changes. Enable workflows (gh workflow enable, or tell
   me to click "Enable workflows" in the Actions tab if the CLI can't), then trigger
   the first run: gh workflow run "Scrape Jobs" and watch it with gh run watch.
8. When the run succeeds, verify: my data repo README has job tables with rows, and
   my fork's README has a populated "My Target Companies" section. Send me both links.

Debug failures yourself before asking me. Never push to vanshsinhaa's repos.
````

The agent will stop twice for input: your repo/watchlist choices at the start, and the PAT paste in step 5 (agents can't create tokens for you — that's the one thing that stays manual).

### ✍️ Option B: manual setup

### 1. Fork and create your data repo

Fork this repo, then create a public repo for the board (e.g. `yourname/jobs`) containing:

- `local_data/job_ids.json` with content `[]`
- `README.md` with the two-table skeleton — copy it from **[CONTRIBUTING.md → Data repo skeleton](CONTRIBUTING.md#data-repo-skeleton)** (it lives there because the table writer pattern-matches this README, so an inline example would get job rows injected into it)

### 2. Point CI at your repos

In `.github/workflows/scrape.yml`, change `repository: vanshsinhaa/jobs` to your data repo. Then in your fork's **Settings → Secrets and variables → Actions** add:

| Secret | What it is |
|---|---|
| `JOBS_REPO_TOKEN` | Fine-grained PAT with **Contents: Read and Write** on your data repo (fine-grained PATs default to read-only — set write explicitly) |
| `DISCORD_WEBHOOK_URL` | Discord channel webhook. Optional — omit to disable notifications |

### 3. Pick your target companies

Edit `local_data/target_companies.json`:

```json
["Google", "Meta", "Stripe", "Anthropic", "Notion", "Figma"]
```

Names are matched case-insensitively against scraped company names, with aliases for renamed brands (see `database/target_companies.go`). The target feed at the bottom of this README regenerates every run.

### 4. Enable workflows

Actions tab → **Enable workflows** (forks disable them by default). The cron fires at the top of every UTC hour. Done — your board now maintains itself.

---

## Running locally

```bash
go run main.go                       # full run: scrape → classify → README + jobs.json
go run main.go --watch --interval=15m  # daemon mode with per-job Discord alerts
go run main.go --target-report       # 7-day coverage per target company (spots broken scrapers)
go run ./cmd/baseline                # test every scraper in isolation, no side effects
```

Local runs read and write `local_data/` and this repo's `README.md` — CI state lives in the data repo, so local experiments can't corrupt production dedup state.

### Configuration

| Env var | Default | Purpose |
|---|---|---|
| `DATA_DIR` | `local_data` | Directory for `job_ids.json`, `jobs.db`, `jobs.json` |
| `README_PATH` | `README.md` | Where the public intern/SWE tables are written |
| `TARGET_README_PATH` | `README.md` | Where the target-companies feed is written |
| `TARGET_COMPANIES_FILE` | `local_data/target_companies.json` | Watchlist location |
| `DISCORD_WEBHOOK_URL` | _(empty = disabled)_ | Webhook for run summaries / watch alerts |

Copy `.env.example` to `.env` for local overrides.

---

## Adding a company

**On Greenhouse, Ashby, or Lever?** It's a three-step, ~5-line change — find the board token, add a company code, add a one-line wrapper + dispatch case. Full walkthrough with copy-paste snippets in **[CONTRIBUTING.md](CONTRIBUTING.md)**.

**On Workday?** Copy any file in `workday/`, point it at the tenant's CXS endpoint, and register the payload. Facet IDs come from the careers site's network tab.

**Something custom?** See the endpoint-discovery playbook in [CONTRIBUTING.md](CONTRIBUTING.md) — most "custom" career sites are one of six ATS platforms wearing a costume.

Then verify with `go run ./cmd/baseline` — your company should appear with a non-zero count.

---

## Project structure

```
jobscanner/
├── main.go                      entry point; --watch and --target-report flags
├── cmd/baseline/                isolated per-scraper diagnostic harness
├── common/                      JobPosting struct, company codes, HTTP client (retries + backoff)
├── common_const/                env-var-backed path config
├── database/                    SQLite layer (modernc.org/sqlite — pure Go, no CGO)
│   ├── classify.go              intern / new_grad / general title classification
│   ├── insert_data.go           single-transaction batch insert
│   └── target_companies.go      watchlist queries + brand alias table
├── process/                     orchestration, dedup, sub-brand retagging
├── readme/                      README table writers (public board + target feed)
├── notify/                      Discord webhook embeds
├── sites/                       custom scrapers + generic Greenhouse/Ashby/Lever engines
├── sites_main/                  company → scraper dispatch
├── workday/                     one payload config per Workday company
├── workday_main/                generic Workday CXS engine (pagination, retries)
├── plan/tracker.md              per-company scraper status ledger + debugging playbook
└── local_data/                  local state (CI uses the data repo instead)
```

---

## Scraper health

Current per-company status — including root-cause history for every outage and the re-diagnosis playbook — lives in **[plan/tracker.md](plan/tracker.md)**.

Known-blocked (documented, not bugs): **Tesla** (Akamai bot wall — needs a real browser), **LinkedIn** (no public job API). Everything else returns live postings as of the last audit.

---

## Contributing

Broken scraper? Company you want tracked? PRs are very welcome — most additions are a handful of lines thanks to the generic engines. See **[CONTRIBUTING.md](CONTRIBUTING.md)** for the scraper-writing guide, debugging playbook, and PR checklist. Bug reports: [open an issue](../../issues/new/choose).

## License

[MIT](LICENSE). Job posting data belongs to the respective companies; this tool only republishes titles, locations, and links to official application pages. Scrapers are rate-limited and hit public, unauthenticated endpoints.

---

## 🎯 My Target Companies

_Updated automatically every hour by CI._

| Company | Role | Location | Apply | Posted |
| --- | --- | --- | :---: | :---: |
| **Adobe** | 2026 AI/ML Intern - Machine Learning Engineer/Researcher  Intern | 3 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-AI-ML-Intern---Machine-Learning-Engineer-Intern_R160706" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Adobe** | 2026 Intern - Research Scientist/Engineer | 7 Locations | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/San-Jose/XMLNAME-2026-Intern---Research-Scientist-Engineer_R160317" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Stripe** | Software Engineer, Intern | Bengaluru | <a href="https://stripe.com/jobs/search?gh_jid=8031833" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 10 |
| **Notion** | Software Engineer Intern (Fall 2026) | San Francisco, California | <a href="https://jobs.ashbyhq.com/notion/5b15697c-fa91-4511-9482-c98a6ff29f90" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 06 |
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
| **Microsoft** | Security Research Intern | Israel, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556768751" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 05 |
| **Microsoft** | IMDA-CLT Program: Solution Engineering INTERN | Singapore, Singapore, Singapore | <a href="https://apply.careers.microsoft.com/careers/job/1970393556869379" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 06 |
| **Microsoft** | Data Center Technicians Intern | Spain, Madrid, Madrid | <a href="https://apply.careers.microsoft.com/careers/job/1970393556873445" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 06 |
| **Microsoft** | Critical Environment Technician Intern | Ireland, Dublin, Dublin | <a href="https://apply.careers.microsoft.com/careers/job/1970393556869574" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Microsoft** | Applied Sciences INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556917519" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 30 |
| **Microsoft** | Data Science INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556917520" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 30 |
| **Microsoft** | Critical Environment Ops INTERN | Canada, Ontario, Greater Toronto | <a href="https://apply.careers.microsoft.com/careers/job/1970393556866710" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Microsoft** | Software Engineering INTERN | Brazil, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556875247" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 02 |
| **Microsoft** | Software Engineering INTERN | India, Multiple Locations, Multiple Locations | <a href="https://apply.careers.microsoft.com/careers/job/1970393556911730" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 01 |
| **Pinterest** | PhD Fall Machine Learning Intern (ATG — Visual, Multimodal, and Recommender Systems) | San Francisco, CA, US; Palo Alto, CA, US; Seattle, WA, US; New York, NY, US | <a href="https://www.pinterestcareers.com/jobs/?gh_jid=7255640" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 04 |
| **Airbnb** | Sales Operations Intern, Italy | Milan, Italy | <a href="https://careers.airbnb.com/positions/7979270?gh_jid=7979270" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 05 |
| **Instacart** | Machine Learning Engineer, PhD Intern (Fall) | United States - Remote | <a href="https://instacart.careers/job/?gh_jid=5917202" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 24 |
| **Instacart** | Machine Learning PhD Intern, Economics (Fall) | United States - Remote | <a href="https://instacart.careers/job/?gh_jid=7532267" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 24 |
| **Meta** | Research Scientist Intern, AI Alignment | Bellevue, WA, Menlo Park, CA, Seattle, WA, Boston, MA, New York, NY, San Francisco, CA | <a href="https://www.metacareers.com/jobs/1782902493113620" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, 3D Vision & World Simulation (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/2839011673109571" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern PhD,  Applied Research | Menlo Park, CA | <a href="https://www.metacareers.com/jobs/2633206137040139" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, NMR Analysis Automation | Redmond, WA | <a href="https://www.metacareers.com/jobs/1418337243438665" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Photorealistic Telepresence (PhD) | Sausalito, CA, Pittsburgh, PA, Redmond, WA | <a href="https://www.metacareers.com/jobs/2022109075207025" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Embedded Contextual AI on Wearables (PhD) | Sunnyvale, CA, Redmond, WA | <a href="https://www.metacareers.com/jobs/2160167211413098" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Monetization Generative AI - LLM (PhD) | Bellevue, WA, Menlo Park, CA, Seattle, WA | <a href="https://www.metacareers.com/jobs/2916726525182155" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Optical System Design (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/1710381673750348" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, State Estimation for Dexterous Manipulation (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/2774289902955470" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, FAIR - Language & Multimodal Foundations (PhD) | Menlo Park, CA, Seattle, WA, New York, NY | <a href="https://www.metacareers.com/jobs/24536664159369645" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Machine Perception for Input and Interaction (PhD) | Redmond, WA, Seattle, WA | <a href="https://www.metacareers.com/jobs/779670167783218" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Meta** | Research Scientist Intern, Applied Vision and Image Quality (PhD) | Redmond, WA | <a href="https://www.metacareers.com/jobs/1422892385992613" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Square** | Applied Research Intern, Proactive Intelligence & Customer World Models (PhD / Graduate Co-op) | Bay Area, CA, United States of America | <a href="http://block.xyz/careers/jobs/5108007008?gh_jid=5108007008" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 03 |
| **Square** | Applied Research Intern, Proactive Intelligence & Customer World Models (PhD / Graduate Co-op) | Toronto, Ontario , Canada | <a href="http://block.xyz/careers/jobs/5108009008?gh_jid=5108009008" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 03 |
| **Adobe** | 2026 University Graduate - Machine Learning Engineer | Seattle | <a href="https://adobe.wd5.myworkdayjobs.com/en-US/external_experienced/job/Seattle/XMLNAME-2026-University-Graduate---Machine-Learning-Engineer_R160133" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Formal Verification Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Formal-Verification-Engineer---New-College-Grad-2026_JR2020837" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Yesterday |
| **Nvidia** | ASIC Design Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/ASIC-Design-Engineer---New-College-Grad-2026_JR2020309" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Yesterday |
| **Nvidia** | ASIC Verification Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/ASIC-Verification-Engineer---New-College-Grad-2026_JR2020640" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 3 Days Ago |
| **Nvidia** | Compiler Engineer - Smart Network Devices- New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Compiler-Engineer---Smart-Network-Devices--New-College-Grad-2026_JR2020535" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 11 Days Ago |
| **Nvidia** | ASIC Floorplan Design Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/ASIC-Floorplan-Design-Engineer---New-College-Grad-2026_JR2012971" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 11 Days Ago |
| **Nvidia** | ASIC Physical Design and Timing Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/ASIC-Physical-Design-and-Timing-Engineer---New-College-Grad-2026_JR2019810" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 14 Days Ago |
| **Nvidia** | System Design Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/System-Design-Engineer---New-College-Grad-2026_JR2011879" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 16 Days Ago |
| **Nvidia** | Circuit Design Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Circuit-Design-Engineer---New-College-Grad-2026_JR2019567" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 29 Days Ago |
| **Nvidia** | GPU Architect - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/GPU-Architect---New-College-Grad-2026_JR2019445" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | ASIC Verification Engineer - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-TX-Austin/ASIC-Verification-Engineer---New-College-Grad-2026_JR2012573" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Formal Verification Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Formal-Verification-Engineer---New-College-Grad-2026_JR2019447" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Software R&D Engineer, VLSI Physical Design - New College Grad 2026 | US, TX, Austin | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-TX-Austin/Software-R-D-Engineer--VLSI-Physical-Design---New-College-Grad-2026_JR2019330" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | ASIC Clocks Design Engineer - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/ASIC-Clocks-Design-Engineer---New-College-Grad-2026_JR2019229-1" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Software Engineer, Hardware Tools and Methodology - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Software-Engineer--Hardware-Tools-and-Methodology---New-College-Grad-2026_JR2018659" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | AI Inference Performance Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/AI-Inference-Performance-Engineer---New-College-Grad-2026_JR2014441" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Cell Modeling and Verification Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Cell-Modelling-and-Verification-Engineer---New-College-Grad-2026_JR2011631" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Software R&D Engineer, Digital Logic Synthesis - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Software-R-D-Engineer--Digital-Logic-Synthesis---New-College-Grad-2026_JR2018263" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Low Power ASIC Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Low-Power-ASIC-Engineer---New-College-Grad-2026_JR2017005" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Signoff Methodology Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Signoff-Methodology-Engineer---New-College-Grad-2026_JR2014110" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Graphics Architect, Hardware - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Graphics-Architect--Hardware---New-College-Grad-2026_JR2013161" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Formal Verification Engineer - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-TX-Austin/Formal-Verification-Engineer---New-College-Grad-2026_JR2013065" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | ASIC Design Verification Engineer - New College Grad 2026 | US, TX, Austin | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-TX-Austin/ASIC-Design-Verification-Engineer---New-College-Grad-2026_JR2010391" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Circuit Design Engineer - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Circuit-Design-Engineer---New-College-Grad-2026_JR2018635" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Signal and Power Integrity Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Signal-and-Power-Integrity-Engineer---New-College-Grad-2026_JR2017741" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Systems Software Engineer - New College Grad 2026 | US, OR, Hillsboro | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-OR-Hillsboro/Systems-Software-Engineer---New-College-Grad-2026_JR2017083" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Power Architect - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Power-Architect---New-College-Grad-2026_JR2017842" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | GPU System and Scheduling Architect - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/GPU-System-and-Scheduling-Architect---New-College-Grad-2026_JR2016691-1" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | ASIC Physical Design Engineer, Netlisting - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/ASIC-Physical-Design-Engineer--Netlisting---New-College-Grad-2026_JR2017681" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Verification Engineer - New College Grad 2026 | 3 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Verification-Engineer---New-College-Grad-2026_JR2017633" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Power Methodology and Modeling Engineer - New College Grad 2026 | US, TX, Austin | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-TX-Austin/Power-Methodology-and-Modeling-Engineer---New-College-Grad-2026_JR2017486-1" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | GPU Power Architect - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/GPU-Power-Architect---New-College-Grad-2026_JR2017169" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | ASIC Clocks Verification Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/ASIC-Clocks-Verification-Engineer---New-College-Grad-2026_JR2013336" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Applied Machine Learning Engineer, Circuit Design - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Applied-Machine-Learning-Engineer--Circuit-Design---New-College-Grad-2026_JR2011517" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Low Power ASIC Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Low-Power-ASIC-Engineer---New-College-Grad-2026_JR2017001" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | ASIC Verification Engineer - New College Grad 2026 | 3 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-NC-Durham/ASIC-Verification-Engineer---New-College-Grad-2026_JR2016248" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | ASIC Hardware Design Engineer - New College Grad 2026 | US, TX, Austin | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-TX-Austin/ASIC-Hardware-Design-Engineer---New-College-Grad-2026_JR2011787" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | AI and FSI Developer Technology Engineer - New College Grad 2026 | 3 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/AI-Developer-Technology-Engineer--Financial-Sector---New-College-Grad-2026_JR2013803" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | AI Chip Design Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/AI-Chip-Design-Engineer---New-College-Grad-2026_JR2015487" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | SoC ASIC Verification Engineer – New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/SoC-ASIC-Verification-Engineer---New-College-Grad-2026_JR2015202-1" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Signal and Power Integrity Engineer - New College Grad 2026 | 2 Locations | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-TX-Austin/Signal-and-Power-Integrity-Engineer---New-College-Grad-2026_JR2015067" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Nvidia** | Circuit Design Engineer - New College Grad 2026 | US, CA, Santa Clara | <a href="https://nvidia.wd5.myworkdayjobs.com/en-US/NVIDIAExternalCareerSite/job/US-CA-Santa-Clara/Circuit-Design-Engineer---New-College-Grad-2026_JR2014331" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | 30+ Days Ago |
| **Stripe** | Operations Associate, New Grad (Mexico) | Mexico City, Mexico | <a href="https://stripe.com/jobs/search?gh_jid=7544547" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Stripe** | Software Engineer, New Grad, Developer & End User Experience Platform | Toronto | <a href="https://stripe.com/jobs/search?gh_jid=7991718" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Stripe** | Tech Operations Associate, New Grad (Mexico) | Mexico City, Mexico | <a href="https://stripe.com/jobs/search?gh_jid=7718947" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Notion** | Software Engineer, New Grad | San Francisco, California | <a href="https://jobs.ashbyhq.com/notion/a6311f97-4850-4674-a5f3-d9fe5f6f2555" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 23 |
| **Notion** | Software Engineer, New Grad (AI) | San Francisco, California | <a href="https://jobs.ashbyhq.com/notion/7e6dc7fe-7ddd-42c1-8928-13f7bddb9ec9" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Apr 27 |
| **Apple** | Machine Learning Engineer, Proactive | United States of America | <a href="https://jobs.apple.com/en-us/details/200671602-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 08 |
| **Apple** | Software Engineer - Universal Media | United States of America | <a href="https://jobs.apple.com/en-us/details/200670706-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 01 |
| **Apple** | Backend Software Engineer - Universal Media | United States of America | <a href="https://jobs.apple.com/en-us/details/200669674-3543" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 26 |
| **Apple** | Computer Vision Software Engineer — Camera Technologies & Systems | United States of America | <a href="https://jobs.apple.com/en-us/details/200658419-0836" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 18 |
| **Apple** | Machine Learning Engineer, Web Indexing Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200668682-3337" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Apple** | Machine Learning Engineer, Web Indexing Team | United States of America | <a href="https://jobs.apple.com/en-us/details/200668682-3760" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jun 16 |
| **Robinhood** | Customer Experience Associate (New Grad) | Westlake, TX | <a href="https://boards.greenhouse.io/robinhood/jobs/8024530?t=gh_src=&gh_jid=8024530" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Jul 09 |
| **Google** | Network Operations Engineer, University Graduate |  | <a href="https://www.google.com/about/careers/applications/jobs/results/124862995078488774-network-operations-engineer-university-graduate?location=United+States&target_level=EARLY&target_level=INTERN_AND_APPRENTICE&sort_by=date&page=4" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
| **Google** | Network Operations Residency Program, University Graduate, August 2026 Start |  | <a href="https://www.google.com/about/careers/applications/jobs/results/118981017938600646-network-operations-residency-program-university-graduate-august-2026-start?location=United+States&target_level=EARLY&target_level=INTERN_AND_APPRENTICE&sort_by=date&page=5" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
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
| **Snowflake** | AI Research Scientist, New Grad – Agents & Reinforcement Learning | Bellevue, Washington, United States | <a href="https://jobs.ashbyhq.com/snowflake/1bad12df-f443-426f-9d09-e96fc780d698" target="_blank"><img src="https://i.imgur.com/u1KNU8z.png" width="118" alt="Apply"></a> | Unknown |
<!-- target-rows-end -->
