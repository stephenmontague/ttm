# TTM Tracker - Project Plan

## Overview

**TTM Tracker** (Time to Meeting Tracker) is a public-facing web application that showcases Temporal's durable execution capabilities by tracking how long it takes a BDR (Business Development Rep) to book a meeting with target companies. Each company gets its own long-running Temporal Workflow that persists indefinitely - surviving restarts, deployments, and failures - until a meeting is booked.

The app serves dual purposes:

1. **A real, useful tool** for tracking outreach cadence and leveraging an AI agent for prospecting help.
2. **A living demo** of Temporal's power - durable timers, human-in-the-loop signals, queries, agentic activities, and long-running workflows - that prospects can see running in production.

---

## Tech Stack

| Layer | Technology | Notes |
|-------|-----------|-------|
| **Temporal Worker** | Go | Long-running workflows, activities, signal/query handlers |
| **Backend API** | Go (net/http or Chi router) | REST API that bridges the Next.js frontend to Temporal (starts workflows, sends signals, proxies cached state) |
| **Frontend** | Next.js 16+ (App Router) | Public-facing UI + admin dashboard |
| **UI Components** | shadcn/ui + Tailwind CSS | Clean, modern component library |
| **AI Agent** | Claude API (Anthropic SDK for Go) | Called from Temporal activities for outreach assistance |
| **Database** | PostgreSQL (Railway-managed) | Stores periodically-snapshotted workflow state for the public UI to read from (avoids exposing Temporal queries publicly) |
| **Temporal** | Temporal Cloud | Managed, production-grade |
| **Hosting** | Railway (Hobby tier) | Go backend + Next.js frontend + Postgres as services in one project |
| **Containers** | Docker | Multi-stage Dockerfiles for Go and Next.js; docker-compose for local dev |

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     PUBLIC UI (Next.js)                  │
│  - Company cards with live elapsed time                 │
│  - Curated activity feed (sanitized)                    │
│  - "Workflow survived X restarts" counter               │
│  - Wall of wins (completed meetings)                    │
└──────────────────────┬──────────────────────────────────┘
                       │ reads from
                       ▼
┌──────────────────────────────────────────────────────────┐
│                   PostgreSQL State Cache                      │
│  - Periodically updated by the workflow                  │
│  - Public-safe fields only (no PII, no contact details)  │
└──────────────────────▲───────────────────────────────────┘
                       │ writes to (via activity)
                       │
┌──────────────────────┴───────────────────────────────────┐
│              TEMPORAL WORKFLOW (Go Worker)                │
│                                                          │
│  ┌─────────────────────────────────────────────────┐     │
│  │  CompanyOutreachWorkflow (one per company)       │     │
│  │                                                  │     │
│  │  State:                                          │     │
│  │   - CompanyName, StartedAt, ElapsedDays          │     │
│  │   - OutreachAttempts []Attempt                    │     │
│  │   - CurrentContact                               │     │
│  │   - AgentSuggestions []Suggestion                 │     │
│  │   - WorkerRestartCount                           │     │
│  │   - Status (active | meeting_booked | paused)    │     │
│  │                                                  │     │
│  │  Signals (HITL):                                 │     │
│  │   - LogOutreach(channel, notes)                  │     │
│  │   - UpdateContact(name, role)                    │     │
│  │   - RequestAgentHelp(taskType)                   │     │
│  │   - MeetingBooked(date, notes)                   │     │
│  │   - PauseWorkflow / ResumeWorkflow               │     │
│  │                                                  │     │
│  │  Queries:                                        │     │
│  │   - GetCurrentState -> full workflow state        │     │
│  │                                                  │     │
│  │  Internal Loop:                                  │     │
│  │   1. Wait for signal OR timer (daily tick)       │     │
│  │   2. On timer tick:                              │     │
│  │      - Update elapsed time                       │     │
│  │      - Snapshot state to PostgreSQL (activity)       │     │
│  │   3. On signal:                                  │     │
│  │      - Process signal                            │     │
│  │      - If RequestAgentHelp -> run agent activity  │     │
│  │      - Snapshot state to PostgreSQL                  │     │
│  │   4. If MeetingBooked -> complete workflow        │     │
│  │   5. Periodically continue-as-new to manage      │     │
│  │      event history size                          │     │
│  └─────────────────────────────────────────────────┘     │
│                                                          │
│  Activities:                                             │
│   - SnapshotStateToCache(state) -> writes to PostgreSQL   │
│   - RunAgent(agentRequest) -> calls Claude API           │
│   - LogRestartEvent() -> increments restart counter      │
│                                                          │
└──────────────────────────────────────────────────────────┘
                       │
                       │ signals sent from
                       ▼
┌──────────────────────────────────────────────────────────┐
│              ADMIN DASHBOARD (Next.js, auth-gated)       │
│                                                          │
│  - Send signals to workflows (log outreach, request      │
│    agent help, update contact, book meeting)             │
│  - View full workflow state (including PII/contacts)     │
│  - View agent suggestions and draft messages             │
│  - Start new company workflows                           │
│  - Pause/resume workflows                                │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## Workflow Definition (Detailed)

### State Struct

```go
type WorkflowState struct {
    CompanyName        string
    StartedAt          time.Time
    Status             string // "active", "paused", "meeting_booked"
    CurrentContact     *Contact
    OutreachAttempts   []OutreachAttempt
    AgentSuggestions   []AgentSuggestion
    WorkerRestartCount int
    LastSnapshotAt     time.Time
    MeetingBookedAt    *time.Time
    MeetingNotes       string
}

type Contact struct {
    Name  string
    Role  string
    LinkedIn string
}

type OutreachAttempt struct {
    Timestamp time.Time
    Channel   string // "email", "linkedin", "slack", "phone", "other"
    Notes     string
    Contact   string // name of who was contacted
}

type AgentSuggestion struct {
    Timestamp   time.Time
    TaskType    string // "draft_message", "suggest_contact", "next_action"
    Request     string
    Response    string
    DraftMessage string // if task was draft_message
}
```

### Signal Types

```go
// LogOutreachSignal - BDR logs that they reached out
type LogOutreachSignal struct {
    Channel string
    Notes   string
}

// UpdateContactSignal - switch to a different person at the company
type UpdateContactSignal struct {
    Name     string
    Role     string
    LinkedIn string
}

// RequestAgentHelpSignal - triggers the AI agent activity
type RequestAgentHelpSignal struct {
    TaskType string // "draft_message" | "suggest_contact" | "next_action"
    Context  string // optional additional context from the BDR
}

// MeetingBookedSignal - terminal signal, workflow completes
type MeetingBookedSignal struct {
    Date  time.Time
    Notes string
}

// PauseSignal / ResumeSignal - pause or resume outreach tracking
type PauseSignal struct{}
type ResumeSignal struct{}
```

### Workflow Logic (Pseudocode)

```
func CompanyOutreachWorkflow(ctx, params):
    state = initState(params)
    
    // Register signal channels
    logOutreachCh = RegisterSignal("log_outreach")
    updateContactCh = RegisterSignal("update_contact")
    agentHelpCh = RegisterSignal("request_agent_help")
    meetingBookedCh = RegisterSignal("meeting_booked")
    pauseCh = RegisterSignal("pause")
    resumeCh = RegisterSignal("resume")
    
    // Register query handler
    SetQueryHandler("get_state", func() -> state)
    
    // Main loop
    eventCount = 0
    for state.Status != "meeting_booked":
        // Selector: wait for any signal OR a daily timer
        selector = NewSelector()
        
        selector.AddReceive(logOutreachCh, func(signal):
            state.OutreachAttempts = append(state.OutreachAttempts, ...)
            ExecuteActivity(SnapshotStateToCache, state)
        )
        
        selector.AddReceive(updateContactCh, func(signal):
            state.CurrentContact = signal
            ExecuteActivity(SnapshotStateToCache, state)
        )
        
        selector.AddReceive(agentHelpCh, func(signal):
            // This is the agentic portion!
            agentResult = ExecuteActivity(RunAgent, AgentRequest{
                CompanyName:  state.CompanyName,
                ElapsedDays:  daysSince(state.StartedAt),
                Attempts:     state.OutreachAttempts,
                CurrentContact: state.CurrentContact,
                TaskType:     signal.TaskType,
                ExtraContext: signal.Context,
            })
            state.AgentSuggestions = append(state.AgentSuggestions, agentResult)
            ExecuteActivity(SnapshotStateToCache, state)
        )
        
        selector.AddReceive(meetingBookedCh, func(signal):
            state.Status = "meeting_booked"
            state.MeetingBookedAt = signal.Date
            state.MeetingNotes = signal.Notes
            ExecuteActivity(SnapshotStateToCache, state)
        )
        
        selector.AddReceive(pauseCh, func(signal):
            state.Status = "paused"
            // Block until resume signal
            resumeCh.Receive(ctx)
            state.Status = "active"
            ExecuteActivity(SnapshotStateToCache, state)
        )
        
        // Daily timer tick - snapshot state even if no signals
        selector.AddTimeout(24 * time.Hour, func():
            ExecuteActivity(SnapshotStateToCache, state)
        )
        
        selector.Select(ctx)
        eventCount++
        
        // Continue-as-new every 1000 events to manage history size
        if eventCount > 1000:
            return ContinueAsNew(ctx, state)
    
    // Workflow complete - final snapshot
    ExecuteActivity(SnapshotStateToCache, state)
    return state
```

---

## AI Agent Activity (Detailed)

The `RunAgent` activity implements a real agentic loop using Claude's tool-use capability. Instead of a single prompt-and-response, Claude autonomously decides which tools to call, evaluates results, and iterates until it has enough information to complete the task. This runs as a Temporal activity (or child workflow for longer research tasks), so if any step fails, Temporal retries it automatically.

### Agent Architecture

```
Signal: RequestAgentHelp("find_contact", "Need someone on the backend team")
        │
        ▼
┌─────────────────────────────────────────────────────┐
│  RunAgent Activity (Go)                             │
│                                                     │
│  1. Build initial prompt with workflow state         │
│     (company, days elapsed, past contacts, attempts)│
│                                                     │
│  2. Call Claude Messages API with tools:             │
│     - web_search(query)                             │
│     - get_company_profile(domain) [Apollo/PDL]      │
│     - get_workflow_state()                          │
│     - draft_message(contact, channel, context)      │
│                                                     │
│  3. AGENTIC LOOP:                                   │
│     while Claude returns tool_use blocks:            │
│       a. Execute the requested tool(s)              │
│       b. Return tool results to Claude              │
│       c. Claude reasons about results               │
│       d. Claude decides: call another tool OR       │
│          return final answer                        │
│                                                     │
│  4. Parse Claude's final response                   │
│  5. Return structured AgentSuggestion               │
└─────────────────────────────────────────────────────┘
```

### Tools Available to the Agent

**`web_search(query string)`**
- Executes a web search via Brave Search API or Serper
- Used by the agent to research companies, find engineers, discover tech stack details, find blog posts, etc.
- Example queries the agent might generate:
  - "Whoop engineering team site:linkedin.com"
  - "Whoop backend infrastructure blog"
  - "Whoop hiring distributed systems engineer"
  - "Whoop tech stack microservices"

**`get_company_profile(domain string)`** *(optional, Phase 3+)*
- Calls Apollo.io or People Data Labs API for structured company/people data
- Returns employee count, industry, recent hires, tech stack signals
- Apollo free tier: 1,200 credits/year (enough for this project)
- Can be skipped in v1 - web search alone covers most use cases

**`get_workflow_state()`**
- Returns the current workflow state (contacts, attempts, elapsed days)
- Lets the agent reason about what's already been tried

**`draft_message(contact, channel, context)`**
- A "self-tool" where Claude drafts the actual outreach message
- Takes the research it gathered and produces a ready-to-send message
- The agent calls this as a final step after gathering enough context

### Agent Task Types

#### `find_contact`
The agent's goal is to find a new person to reach out to at the target company.

Example agent loop:
1. Claude calls `get_workflow_state()` to see who's already been contacted
2. Claude calls `web_search("Whoop engineering team backend")` to find engineers
3. Claude evaluates results, decides it needs more specificity
4. Claude calls `web_search("Whoop distributed systems infrastructure engineer site:linkedin.com")`
5. Claude identifies a promising candidate from search results
6. Claude calls `web_search("[candidate name] Whoop engineering blog talk")` to learn more
7. Claude returns: suggested contact, reasoning, and a LinkedIn search query

#### `draft_message`
The agent's goal is to craft a personalized outreach message.

Example agent loop:
1. Claude calls `get_workflow_state()` to understand outreach history and elapsed time
2. Claude calls `web_search("Whoop engineering challenges 2026")` for relevant pain points
3. Claude calls `web_search("Whoop [current contact name] conference talk open source")` for personalization hooks
4. Claude synthesizes findings and calls `draft_message(...)` with full context
5. Claude returns: the draft, recommended channel, and reasoning

#### `next_action`
The agent's goal is to recommend what to do next based on the full outreach history.

Example agent loop:
1. Claude calls `get_workflow_state()` to review all attempts
2. Claude calls `web_search("Whoop recent news funding hiring")` for timing signals
3. Claude reasons about cadence, channels tried, and response patterns
4. Claude returns: recommended action, timing, and reasoning

---

## Frontend Pages

### Public Pages

#### `/` - Home / Dashboard
- Hero section: "Temporal-Powered Outreach Tracker" with brief explanation
- **Active Workflow Cards** - one per company, showing:
  - Company name (or logo if available)
  - Elapsed time (days, hours, minutes - live-updating counter)
  - Number of outreach attempts
  - Current status badge (Active, Paused)
  - "Workflow has survived X worker restarts" badge
  - Last activity (sanitized, e.g., "LinkedIn outreach sent 3 days ago")
- **Wall of Wins** section - completed workflows showing:
  - Company name
  - Total TTM (time to meeting)
  - Number of outreach attempts it took
  - "Meeting booked!" celebration state
- **"How It Works"** expandable section explaining:
  - This is a real Temporal workflow running in production
  - Durable execution means it survives crashes and restarts
  - Human-in-the-loop signals drive the workflow forward
  - An AI agent assists with outreach strategy
- Footer with link to Temporal, your LinkedIn, etc.

#### `/company/:slug` - Company Detail (public)
- Larger elapsed time display
- Sanitized activity timeline (curated, no PII):
  - "Outreach attempt #3 via LinkedIn"
  - "AI agent suggested next action"
  - "Worker restarted - workflow resumed"
  - "New contact identified"
- Temporal feature callouts (what's being demonstrated at each step)

### Admin Pages (auth-gated, simple password or env-based auth)

#### `/admin` - Admin Dashboard
- List of all workflows (active + completed)
- "Start New Workflow" button (company name input)
- Quick-action buttons per workflow

#### `/admin/company/:slug` - Admin Company Detail
- Full workflow state (including contacts, agent suggestions, raw notes)
- **Signal Panel:**
  - "Log Outreach" form (channel dropdown, notes textarea)
  - "Update Contact" form (name, role, LinkedIn URL)
  - "Request Agent Help" (task type dropdown + optional context)
  - "Meeting Booked" (date picker, notes)
  - "Pause" / "Resume" toggle
- **Agent Suggestions** section - shows all AI responses with ability to copy draft messages
- **Full Activity Timeline** (unfiltered)

---

## API Endpoints (Go Backend)

### Public Endpoints (no auth)

```
GET  /api/companies              - list all tracked companies (cached state from PostgreSQL)
GET  /api/companies/:slug        - single company public state
GET  /api/companies/:slug/feed   - sanitized activity feed for a company
GET  /api/stats                  - aggregate stats (total workflows, avg TTM, etc.)
```

### Admin Endpoints (auth required)

```
POST /api/admin/companies                          - start a new workflow
GET  /api/admin/companies/:slug                    - full state (including PII)
POST /api/admin/companies/:slug/signal/outreach    - send LogOutreach signal
POST /api/admin/companies/:slug/signal/contact     - send UpdateContact signal
POST /api/admin/companies/:slug/signal/agent       - send RequestAgentHelp signal
POST /api/admin/companies/:slug/signal/booked      - send MeetingBooked signal
POST /api/admin/companies/:slug/signal/pause       - send Pause signal
POST /api/admin/companies/:slug/signal/resume      - send Resume signal
```

---

## PostgreSQL Schema

```sql
-- Cached workflow state for public consumption
CREATE TABLE company_workflows (
    id              TEXT PRIMARY KEY,  -- workflow ID (slug)
    company_name    TEXT NOT NULL,
    slug            TEXT UNIQUE NOT NULL,
    started_at      TIMESTAMPTZ NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active',  -- active, paused, meeting_booked
    elapsed_days    INTEGER DEFAULT 0,
    outreach_count  INTEGER DEFAULT 0,
    restart_count   INTEGER DEFAULT 0,
    current_contact_role TEXT,  -- role only, no name (for public display)
    meeting_booked_at    TIMESTAMPTZ,
    last_snapshot_at     TIMESTAMPTZ,
    updated_at           TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Sanitized activity feed for public display
CREATE TABLE activity_feed (
    id          SERIAL PRIMARY KEY,
    workflow_id TEXT NOT NULL REFERENCES company_workflows(id),
    timestamp   TIMESTAMPTZ NOT NULL,
    event_type  TEXT NOT NULL,  -- outreach, contact_change, agent_action, restart, meeting_booked
    description TEXT NOT NULL,  -- sanitized, human-readable description
    channel     TEXT,           -- email, linkedin, slack, phone (for outreach events)
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Full agent suggestions (admin only, not exposed publicly)
CREATE TABLE agent_suggestions (
    id          SERIAL PRIMARY KEY,
    workflow_id TEXT NOT NULL REFERENCES company_workflows(id),
    task_type   TEXT NOT NULL,
    request     TEXT,
    response    TEXT,
    draft_message TEXT,
    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

---

## Project Structure

> **Note:** The Go backend lives in `server/` using standard Go layout (`cmd/` + `internal/`). The API server and worker are separate binaries.

```
ttm-tracker/
├── .env                            # Environment variables (gitignored)
├── .env.example                    # Template for environment variables
├── .gitignore
├── docker-compose.yml              # Local dev: Postgres (worker + API run via go run)
│
├── server/                         # Go backend
│   ├── go.mod                      # Go module (github.com/.../ttm-tracker/server)
│   ├── go.sum
│   ├── cmd/                        # Application entry points
│   │   ├── api/
│   │   │   └── main.go             # REST API server (Temporal client + PostgreSQL)
│   │   └── worker/
│   │       └── main.go             # Temporal worker (executes workflows + activities)
│   └── internal/                   # Private Go packages
│       ├── activities/
│       │   └── snapshot.go         # SnapshotStateToCache activity (writes to PostgreSQL)
│       ├── api/
│       │   └── handler.go          # HTTP handlers (public + admin endpoints)
│       ├── config/
│       │   └── config.go           # Constants, env helpers, signal/query names
│       ├── database/
│       │   ├── database.go         # PostgreSQL connection pool (pgx)
│       │   └── migrations.go       # Schema creation on startup
│       ├── models/
│       │   └── outreach.go         # Shared types (WorkflowState, signals, etc.)
│       ├── repository/
│       │   └── company.go          # Database query functions
│       ├── temporal/
│       │   └── client.go           # Temporal Cloud mTLS client
│       └── workflow/
│           └── outreach/
│               └── workflow.go     # CompanyOutreachWorkflow definition
│
├── frontend/                       # Next.js 16 application
│   ├── package.json
│   ├── next.config.ts
│   ├── components.json             # shadcn/ui config
│   ├── .env.local                  # API_URL (server-side only, no NEXT_PUBLIC_)
│   ├── app/
│   │   ├── globals.css             # Tailwind + shadcn theme (light + dark)
│   │   ├── layout.tsx              # Root layout (ThemeProvider, Geist fonts, Toaster)
│   │   ├── page.tsx                # Public home / dashboard (SSR)
│   │   ├── dashboard-content.tsx   # Client wrapper for polling
│   │   ├── company/
│   │   │   └── [slug]/
│   │   │       └── page.tsx        # Public company detail (SSR)
│   │   ├── admin/
│   │   │   ├── layout.tsx          # Admin layout wrapper
│   │   │   ├── page.tsx            # Admin dashboard
│   │   │   └── company/
│   │   │       └── [slug]/
│   │   │           └── page.tsx    # Admin company detail + signal panel
│   │   └── api/                    # Next.js API routes (proxy to Go backend)
│   │       ├── companies/
│   │       │   ├── route.ts        # GET  /api/companies
│   │       │   └── [slug]/
│   │       │       ├── route.ts    # GET  /api/companies/[slug]
│   │       │       └── feed/
│   │       │           └── route.ts # GET /api/companies/[slug]/feed
│   │       └── admin/
│   │           └── companies/
│   │               ├── route.ts    # POST /api/admin/companies
│   │               └── [slug]/
│   │                   ├── route.ts # GET /api/admin/companies/[slug]
│   │                   └── signal/
│   │                       └── [action]/
│   │                           └── route.ts # POST signal dispatcher
│   ├── components/
│   │   ├── ui/                     # shadcn/ui primitives (badge, button, card, etc.)
│   │   ├── theme-provider.tsx      # next-themes wrapper
│   │   ├── theme-toggle.tsx        # Dark/light toggle button
│   │   ├── site-header.tsx         # Sticky header with branding + nav
│   │   ├── site-footer.tsx         # Footer ("Powered by Temporal")
│   │   ├── workflow-card.tsx       # Public workflow card component
│   │   ├── elapsed-timer.tsx       # Live-updating elapsed time display
│   │   ├── status-badge.tsx        # Reusable status indicator
│   │   ├── activity-feed.tsx       # Activity feed timeline
│   │   ├── signal-panel.tsx        # Admin signal forms (outreach, contact, meeting, pause)
│   │   ├── create-workflow-form.tsx # Admin new workflow form
│   │   └── company-stats.tsx       # Stats display row
│   ├── hooks/
│   │   ├── use-polling.ts          # Generic data polling hook
│   │   └── use-signal.ts           # Signal submission + toast hook
│   └── lib/
│       ├── types.ts                # TypeScript type definitions
│       ├── utils.ts                # Utilities (shadcn/ui cn helper)
│       ├── backend.ts              # Server-side Go backend proxy
│       └── constants.ts            # Status config, channel options
```

---

## Environment Variables

```bash
# Temporal Cloud
TEMPORAL_ADDRESS=your-namespace.tmprl.cloud:7233
TEMPORAL_NAMESPACE=your-namespace
TEMPORAL_TLS_CERT_PATH=./certs/client.pem
TEMPORAL_TLS_KEY_PATH=./certs/client.key
TEMPORAL_TASK_QUEUE=ttm-tracker

# Claude API (for agent activities)
ANTHROPIC_API_KEY=sk-ant-...

# Web Search (for agent tools - pick one)
BRAVE_SEARCH_API_KEY=...          # Brave Search API (free tier: 2,000 queries/month)
# SERPER_API_KEY=...              # Alternative: Serper.dev

# Apollo.io (optional, for structured people/company data)
# APOLLO_API_KEY=...              # Free tier: 1,200 credits/year

# PostgreSQL (Railway provides DATABASE_URL automatically)
DATABASE_URL=postgresql://user:password@host:5432/ttm_tracker

# Admin Auth (Phase 3 — seed user + session-based login)
ADMIN_SEED_EMAIL=you@example.com
ADMIN_SEED_PASSWORD=your-secure-password
SESSION_COOKIE_NAME=ttm_session
SESSION_MAX_AGE=604800  # 7 days in seconds

# Frontend (server-side only — used by Next.js API routes to proxy to Go backend)
API_URL=http://localhost:9090/api

# App
PORT=8080
ENVIRONMENT=development
```

---

## Phased Build Plan

### Phase 1: Foundation (Core Workflow + Basic UI) ✅ COMPLETE

**Goal:** Get a workflow running that tracks elapsed time and accepts signals, with a basic UI showing the timer.

**Status:** Complete. All tasks done. Workflow runs on Temporal Cloud, API serves cached state from PostgreSQL, frontend displays live elapsed timers and allows signal sending from admin panel.

**Deviations from plan:**
- Used PostgreSQL instead of SQLite (matches the deployment target and other projects)
- Used standard Go project layout (`cmd/` + `internal/`) instead of flat `worker/` directory
- Implemented ALL signal handlers (not just LogOutreach and MeetingBooked) since the types were already defined
- Admin UI includes full signal panel (outreach, contact, pause/resume, meeting booked)

**Tasks:**

1. ~~**Go module setup** - initialize with Go module, install Temporal Go SDK, set up `main.go`~~ ✅
2. ~~**Workflow definition** - implement `CompanyOutreachWorkflow` with all signal handlers, query handler, daily timer, continue-as-new~~ ✅
3. ~~**PostgreSQL setup** - schema creation (company_workflows, activity_feed, agent_suggestions tables), migrations on startup~~ ✅
4. ~~**Snapshot activity** - `SnapshotStateToCache` writes sanitized state to PostgreSQL~~ ✅
5. ~~**API server** - Chi router with public endpoints (GET /api/companies, /api/companies/:slug, /api/companies/:slug/feed) and admin endpoints (POST for all signal types)~~ ✅
6. ~~**Next.js setup** - initialized with shadcn/ui + Tailwind CSS~~ ✅
7. ~~**Public dashboard page** - company cards with live elapsed timer, Wall of Wins section~~ ✅
8. ~~**Admin page** - start workflow, send all signal types, view full workflow state via Temporal query~~ ✅
9. ~~**Connect to Temporal Cloud** - mTLS configured, workflow verified running~~ ✅

**Known issues to address:**
- Orphaned timer bug exists in the workflow (each loop iteration creates a new timer without cancelling the old one). Fix is written but not deployed to the running workflow. New workflows will need versioning or terminate-and-restart.
- Activity feed table exists but the snapshot activity doesn't write sanitized events to it yet (only writes to company_workflows table)
- No auth middleware on admin routes yet (Phase 2)

**Deliverable:** ✅ Running app where you can start a workflow for "Whoop", log outreach attempts via admin panel, and see the elapsed timer on a public page.

---

### Phase 2: Full HITL Signals + Admin Dashboard ✅ COMPLETE

**Goal:** Complete signal handling and build out the admin experience.

**Status:** Complete. All tasks done. Auth middleware moved to dedicated Phase 3.

**Frontend rebuilt:** The Next.js frontend was rebuilt from scratch with a new architecture. API routes (`app/api/*/route.ts`) now contain full business logic (validation, Go backend proxy, error handling). The browser-side `lib/api.ts` client was removed — pages call Next.js API routes directly via `fetch`. Dark/light theme toggle added via `next-themes`. No `NEXT_PUBLIC_` env vars — the Go backend URL is server-side only.

**Tasks:**

1. ~~**Remaining signals** - implement `UpdateContact`, `Pause`, `Resume` in workflow~~ ✅ (done in Phase 1)
2. ~~**Admin signal panel** - full UI for sending all signal types~~ ✅ (done in Phase 1)
3. ~~**Activity feed** - write sanitized events to `activity_feed` table on each signal~~ ✅ (`PersistWorkflowState` activity writes sanitized events to `activity_feed` table via `sanitize.go`)
4. ~~**Public activity timeline** - show curated feed on company detail page~~ ✅ (UI built and connected to activity feed data)
5. ~~**Admin company detail** - full state view with outreach history, contacts~~ ✅ (done in Phase 1)
6. ~~**Restart counter** - track and display worker restart count~~ ✅ (`restart_count` persisted to DB, displayed on workflow cards and company stats)
7. ~~**Wall of Wins** - completed workflows section on homepage~~ ✅ (done in Phase 1)
8. ~~**Auth middleware** - simple password-based auth for admin routes~~ ⏭️ (moved to dedicated Phase 3)

**Deliverable:** ✅ Full HITL demo - you can manage the entire outreach lifecycle through the admin panel, and the public page shows a rich, sanitized view.

---

### Phase 3: Authentication & Admin Login ✅ COMPLETE

**Goal:** Lock down all admin functionality behind a real login page. You are the only user — the database is seeded with your account. Everyone else sees the public pages only.

**Status:** Complete. All tasks done. Admin dashboard fully locked down behind session-based login.

**Design decisions:**
- **Single-user model** — the `admin_users` table is seeded with one row (you). There is no registration flow.
- **Hidden login page** — `/admin/login` exists but is not linked from any public navigation. You navigate to it directly.
- **Session-based auth** — login sets an HTTP-only cookie containing a 256-bit random session token. No JWTs, no client-side token storage.
- **Server-side enforcement** — the Next.js admin layout uses a `(protected)` route group with a server-side cookie check that redirects unauthenticated requests to `/admin/login`. The Go backend admin endpoints also validate the session via `RequireSession` middleware.
- **Password hashing** — bcrypt-hashed password stored in the database. Plaintext password never persisted.
- **Sliding session expiry** — each authenticated request extends the session by 7 days. Expired sessions are lazily cleaned up.

**Architecture:** Browser → Next.js API routes (forward cookie) → Go backend (validates session in DB). Next.js never reads the session token — it just proxies the `Cookie` header to Go.

**Tasks:**

1. ~~**Database schema + seed** — `admin_users` and `admin_sessions` tables added to migrations. Admin account seeded from `ADMIN_SEED_EMAIL` / `ADMIN_SEED_PASSWORD` env vars on startup via `UpsertAdminUser` (idempotent).~~ ✅
2. ~~**Go auth endpoints** — `POST /api/auth/login` (validates credentials, creates session, sets cookie) and `POST /api/auth/logout` (deletes session, clears cookie). `GET /api/admin/auth/status` returns 200 if session valid (behind middleware).~~ ✅
3. ~~**Go admin middleware** — `RequireSession` middleware on all `/api/admin/*` routes reads session cookie, validates against DB, returns 401 if invalid/missing.~~ ✅
4. ~~**Next.js login page** — `/admin/login` with email + password form. On success, redirects to `/admin`. Clean UI using shadcn Card/Input/Button. Not linked from public nav.~~ ✅
5. ~~**Next.js admin layout guard** — `app/admin/(protected)/layout.tsx` checks session cookie server-side and redirects to `/admin/login` if absent. Login page sits outside the route group so it's never guarded.~~ ✅
6. ~~**Next.js API route proxying** — admin API routes forward the session cookie to the Go backend via `cookieHeader` parameter added to `backendGet`/`backendPost`.~~ ✅
7. ~~**Logout button** — logout button in site header (visible only when logged in). Admin nav link also conditionally rendered. Calls `POST /api/auth/logout`, clears cookie, redirects to `/admin/login`.~~ ✅
8. ~~**Session expiry + cleanup** — sliding 7-day expiry: each valid request extends `expires_at`. Expired sessions lazily deleted in background goroutine on validation.~~ ✅

**Deviations from original plan:**
- Used `(protected)` route group instead of checking pathname in a single layout — cleaner separation, login page can never accidentally be guarded
- `GET /api/auth/me` became `GET /api/admin/auth/status` — placed inside the admin group so the middleware does the validation, handler just returns 200
- Added sliding session expiry (each request extends the 7-day window) instead of fixed expiry
- Site header conditionally renders Admin link and logout button based on cookie presence (server component reads cookies)

**Security notes:**
- Cookie: `HttpOnly`, `SameSite: Lax`, `Secure: false` (toggle to `true` for production HTTPS in Phase 5)
- Session tokens: 64-char hex (256 bits entropy), not guessable
- Login rate limiting planned for Phase 5 (low risk now — hidden login page, single user)

**Deliverable:** ✅ The admin dashboard is fully locked down. You log in at `/admin/login`, get a session cookie, and access all admin features. Anyone without the cookie sees only public pages. No registration, no invite flow — just you.

---

### Phase 4: AI Agent Integration ✅ COMPLETE

**Goal:** Add a real agentic activity that uses Claude's tool-use to autonomously research and assist with outreach.

**Status:** Complete. All core tasks done plus significant UX enhancements beyond the original plan.

**Architecture:** The agent runs as a series of Temporal activities (`CallClaude`, `ExecuteAgentTool`, `SaveAgentSuggestion`) orchestrated by an agentic loop in the workflow. Each Claude API call and tool execution is a separate activity for independent retry and visibility. The workflow sets `AgentTaskInProgress` flag on the live state so queries reflect the agent's working status immediately.

**Activity parameter best practice:** Refactored `PersistWorkflowState` from bare parameters `(state, event)` to a single `PersistWorkflowStateRequest` struct, following Temporal's recommended pattern for forward compatibility.

**Tasks:**

1. ~~**Claude API client with tool-use loop** - Go client (`server/internal/agent/client.go`) calls Claude Messages API, detects `tool_use` response blocks, executes tools via activities, loops until `end_turn` (max 10 iterations)~~ ✅
2. ~~**Web search tool** - Brave Search API integration not yet implemented~~ ⏭️ (deferred — Lusha contact/company search covers the primary use case)
3. ~~**Tool definitions** - `get_workflow_state`, `draft_message`, `lusha_contact_search`, `lusha_company_search` defined in `server/internal/agent/tools.go`~~ ✅
4. ~~**System prompts** - task-specific system prompts for `draft_message`, `suggest_contact`, `next_action` with contact-aware context injection (`server/internal/agent/prompts.go`)~~ ✅
5. ~~**RunAgent activity** - Agentic loop runs in workflow (`runAgentLoop` in `workflow.go`) orchestrating `CallClaude` and `ExecuteAgentTool` activities with timeouts and retry policies~~ ✅
6. ~~**Agent signal handler** - `RequestAgentHelp` signal triggers `runAgentLoop`, sets/clears `AgentTaskInProgress` flag on workflow state~~ ✅
7. ~~**Agent suggestions storage** - `SaveAgentSuggestion` activity persists suggestions; also stored in workflow state for query access~~ ✅
8. ~~**Admin agent UI** - Dedicated agent page (`/admin/company/[slug]/agent`) with contact picker, task type selector, adaptive polling (2s while agent working, 10s otherwise), suggestion history with contact filtering, copy-to-clipboard for drafts, markdown rendering via `react-markdown`~~ ✅
9. ~~**Public agent callout** - sanitized mention on public feed ("AI agent researched company and suggested next action")~~ ✅ (activity feed events generated by `PersistWorkflowState`)
10. ~~**Lusha integration** - `lusha_contact_search` and `lusha_company_search` tools with plug-and-play activation via `LUSHA_API_KEY` env var. `CheckLushaEnabled` activity respects workflow determinism.~~ ✅ (replaces Apollo.io from original plan)

**Enhancements beyond original plan:**
- **Dedicated agent page** — moved from a tab in SignalPanel to a full-screen agent experience at `/admin/company/[slug]/agent`
- **Contact-centric agent requests** — `ContactName` field on `RequestAgentHelpSignal` and `AgentSuggestion`; contact name/role injected into Claude prompts
- **Loading state** — `AgentTaskInProgress` flag on `WorkflowState` enables real-time feedback via adaptive polling
- **Markdown rendering** — agent responses rendered with proper styling (headings, lists, bold, code blocks) via `react-markdown` + custom CSS
- **Reconcile endpoint** — admin can manually sync DB status with Temporal via refresh button (`POST /admin/companies/:slug/reconcile`)
- **Workflow status resilience** — `reconcileStatuses` now distinguishes `serviceerror.NotFound` from transient connection errors, preventing false "terminated" status

**Deliverable:** ✅ A genuinely agentic workflow - you trigger "draft a message for Stephen Montague at Whoop" from the dedicated agent page, and Claude autonomously inspects workflow state, researches the contact, and returns a personalized draft message with markdown formatting. All orchestrated by Temporal with real-time loading feedback.

---

### Phase 5: Polish + Deploy

**Goal:** Production-ready deployment with polish.

**Tasks:**

1. **Automated restart logging** - worker startup detection and counter increment
2. **Rate limiting** - protect public API endpoints
3. **Error handling** - graceful failures in activities with retries
4. **Responsive design** - mobile-friendly public UI
5. **SEO / Open Graph** - meta tags for sharing (great for LinkedIn posts)
6. **Railway deployment** - configure two services (Go backend + Next.js frontend) in one Railway project
7. **CI/CD** - GitHub Actions for build + deploy
8. **Custom domain** - e.g., `ttm.yourdomain.com`
9. **Loom-ready** - ensure the UI is visually clean for screen recordings

**Deliverable:** Live, public app at a custom domain that you can share with prospects and use for Loom videos.

---

## Key Temporal Features Demonstrated

| Feature | How It's Used | Why It Matters |
|---------|--------------|----------------|
| **Durable Execution** | Workflow runs for days/weeks/months without losing state | Core value prop - "set it and forget it" reliability |
| **Human-in-the-Loop (Signals)** | BDR sends signals to log outreach, update contacts, request help | Shows async human interaction with running workflows |
| **Queries** | Admin panel reads live workflow state | Non-intrusive state inspection |
| **Timers** | Daily tick for state snapshots, cadence between outreach | Long-running timers that survive restarts |
| **Activities** | AI agent calls, database writes | Reliable execution of external service calls with retry |
| **Continue-as-New** | Resets event history every 1000 events | Production pattern for infinite-duration workflows |
| **Worker Restart Resilience** | Restart counter shows workflow survives crashes | The "money shot" - workflow is decoupled from workers |

---

## Security Considerations

- **No PII on public pages** - contact names, emails, LinkedIn URLs are admin-only
- **Admin auth** - session-based login with bcrypt-hashed password. Single seeded user, no registration. HTTP-only secure cookie. All admin endpoints validated server-side
- **Rate limiting** - public API endpoints rate-limited to prevent abuse
- **No public query exposure** - Temporal query API is internal only; public UI reads from PostgreSQL cache
- **Environment variables** - all secrets (Temporal certs, API keys, admin password) via env vars, never committed
- **CORS** - configured to allow only the frontend origin

---

## Open Source

**License:** MIT

**Repository:** Public GitHub repo (e.g., `github.com/yourusername/ttm-tracker`)

### What's in the repo (public)
- All application code (Go worker, Next.js frontend, Dockerfiles, docker-compose)
- Database migrations
- Agent tool definitions and system prompts
- Full README with setup instructions, architecture diagram, and screenshots
- `.env.example` with all required variables documented (no actual secrets)
- Contributing guide (`CONTRIBUTING.md`)

### What stays out of the repo (private)
- `.env` files with real API keys, Temporal certs, admin password
- TLS certificates (`certs/` directory)
- Any runtime data (database contents, workflow state)
- Add to `.gitignore`: `.env`, `certs/`, `*.pem`, `*.key`, `data/`, `.next/`, `node_modules/`, `tmp/`

### README structure
1. **Hero** - what this is, screenshot/gif of the live app, link to the running instance
2. **Why this exists** - the story: a BDR built a real prospecting tool using Temporal to showcase durable execution
3. **Temporal features demonstrated** - table of features with explanations (reuse the "Key Temporal Features Demonstrated" table)
4. **Architecture** - the ASCII diagram from this plan
5. **Tech stack** - table of technologies used
6. **Getting started** - local dev setup with docker-compose
7. **Deployment** - Railway setup guide
8. **Configuration** - environment variables reference
9. **How the agent works** - explanation of the agentic tool-use loop
10. **Contributing** - how to contribute, code of conduct
11. **License** - MIT

### What makes this a good OSS project
- **Reference implementation:** Shows how to build long-running workflows + HITL + agentic AI with Temporal in Go
- **Forkable:** Anyone can clone this and adapt it to their own outreach workflow or use case
- **Well-documented:** Architecture decisions are explained, not just code
- **Actually runs in production:** This isn't a toy example - it's a live app tracking real outreach

---

## Notes for Claude Code

- Start with Phase 1 and get a working end-to-end flow before adding complexity
- Use the Go standard library (`net/http`) or a lightweight router like `chi` - no heavy frameworks
- **Docker setup:**
  - Go worker Dockerfile: multi-stage build (Go 1.23+ builder -> Alpine for the final image)
  - Next.js Dockerfile: multi-stage build (Node 22 builder -> Node 22 slim for the final image)
  - `docker-compose.yml` for local development with Go worker, Next.js, and Postgres services
  - Railway will detect and build from the Dockerfiles automatically
- **PostgreSQL:**
  - Use `pgx` (github.com/jackc/pgx/v5) as the Go Postgres driver - it's the modern standard
  - Connect via `DATABASE_URL` environment variable (Railway auto-injects this when you add a Postgres service)
  - Run migrations on startup in the Go worker's `main.go`
- **Next.js 16+ specific requirements:**
  - Use the App Router (not Pages Router)
  - `params` and `searchParams` are async - always `await` them in page/layout components (e.g., `const { slug } = await props.params`)
  - Use `proxy.ts` instead of `middleware.ts` for any request interception (renamed in Next.js 16). Export a `proxy` function, not `middleware`.
  - Turbopack is the default bundler - no webpack config needed
  - Caching is opt-in by default - use `"use cache"` directive explicitly where needed, don't rely on implicit caching
  - Use `npx create-next-app@latest` to scaffold - it will use Next.js 16+ and React 19.2 by default
  - Minimum Node.js version is 20.9.0
- Install shadcn/ui components as needed (Button, Card, Input, Select, Badge, Dialog, etc.)
- The elapsed timer on the frontend should be a client-side component that calculates from `started_at` - it doesn't need to poll the server for time updates
- The agent activity is pure Go - it just makes HTTP calls to the Claude Messages API. No need for Python or separate agent frameworks.
- The agent prompts should include the actual workflow state (days elapsed, attempt count) so Claude can reference real data in draft messages
- For continue-as-new, pass the full `WorkflowState` as the input to the new workflow execution