# Productivity OS — Feature Tracker

**Goal:** `PLAN → DO → TRACK → REVIEW → IMPROVE`

A personal productivity system for planning time, tracking actual activity, managing
tasks and habits, tracking goals, reviewing progress, and understanding productivity
through deterministic analytics — before any intelligent assistance.

> **How this file works.** Every feature below is a checkbox. Mark `[x]` only when the
> feature is built, tested, and reviewed; `[~]` for partial; `[ ]` for not started.
> Phase-level detail lives in `planning.md`. The authoritative scope is
> `docs/requirements/v1.md` — this file tracks progress, it does not define scope.
> Status legend: ✅ complete · 🔨 in progress · ⬜ not started.

---

## V1 — Core Productivity OS

**Progress: 4 / 10 areas built** (Authentication, Timeline, Tasks, Habits). All awaiting
a browser click-through and a first CI run.

| # | Area | Milestone | Status |
|---|------|-----------|--------|
| 1 | Authentication & Account | M1 | ✅ built (browser + CI run still to do) |
| 2 | Daily Timeline & Time Tracking | M2 | ✅ built (browser + CI run still to do) |
| 3 | Tasks + Kanban | M3 | ✅ built (browser + CI run still to do) |
| 4 | Habits & Streaks | M4 | ✅ built (browser + CI run still to do) |
| 5 | Goals | M5 | ⬜ |
| 6 | Reviews (daily + weekly) | M6 | ⬜ |
| 7 | Deterministic Productivity Reports | M7 | ⬜ |
| 8 | Data Export | M8 | ⬜ |
| 9 | Frontend | spans M1–M8 | 🔨 auth shell done |
| 10 | Backend & Infrastructure Foundation | M1 + ongoing | ✅ core done — deploy / ADR-0008 pending |

### 1. Authentication & Account — ✅ M1

- [x] User registration (auto-logs in)
- [x] User login
- [x] User logout
- [x] Server-side sessions in PostgreSQL (token stored hashed)
- [x] Opaque `HttpOnly` + `Secure` + `SameSite=Lax` session cookie
- [x] Authenticated request context (`reqctx.Identity`, from the session only)
- [x] Account-scoped authorization
- [x] Strict per-account data isolation (+ dedicated isolation test suite)
- [x] Argon2id password hashing (19 MiB / t=2 / p=1, PHC format)
- [x] Password policy (Q6: 6–128 chars; lowercase + uppercase + special required)
- [x] Session expiration + invalid-session handling
- [x] Password change (ends every session for the account)
- [x] Duplicate-account handling (`409`, case-insensitive email)
- [x] Structured authentication errors (shared error envelope)
- [x] CSRF protection (double-submit token)
- [x] Login rate limiting (per email + IP) + per-IP auth throttle
- [x] `account_id` comes only from the session, never from client input
- [x] Security review — `docs/security-review-m1.md`
- [ ] Browser click-through verification (needs a browser)
- [ ] CI green on first push (needs a git remote)

**Scope boundary:** no email verification, password reset, OAuth, MFA, or self-service
account deletion. No profile fields beyond email / password / timezone.

### 2. Daily Timeline & Time Tracking — ✅ M2

- [x] `internal/platform/timezone` — DST-correct day/week windows, midnight-spanning totals (N4)
- [x] Timestamps stored as `timestamptz`; server runs UTC; date/week bucketing centralized
- [x] Categories — create / rename / archive / list active (flat, no hard delete, no colour meaning)
- [x] Planned blocks — add / edit / delete; arbitrary `[start, end)`; optional category; may end next day
- [x] Actual blocks — add / edit / delete; may cross midnight; never tied to a planned block
- [x] Timeline view — API + UI (24h grid, planned/actual lanes, block editor, wall-clock times)
- [x] Planned-vs-actual comparison — API + UI (per-category planned/actual/difference table, Uncategorized bucket, total row)
- [x] Account timezone controls which date/week a block belongs to (server-side, no client tz math)

**Scope boundary:** no templates, "copy a day", or recurring plans; no live timer; no
"convert plan to actual"; overlaps allowed and not flagged; one date at a time.

### 3. Tasks + Kanban — ✅ M3

- [x] Create / view / update / complete / reopen tasks (state → any of the four, any direction)
- [x] Task title, optional description, optional due date, status
- [x] Ordering within a column — newest-first, no manual reorder (Q5 default)
- [x] Fixed Kanban board: `BACKLOG` · `TODO` · `IN_PROGRESS` · `DONE`
- [x] Move a task between any columns (`<select>` + drag) — sets its state
- [x] Every state change recorded in `task_transitions` (Q10 — all transitions logged)

**Scope boundary:** columns are fixed and uneditable; no recurring tasks, subtasks,
priorities, labels, estimates, attachments, comments, or dependencies.

### 4. Habits & Streaks — ✅ M4

- [x] Create a habit (daily cadence only)
- [x] Mark / unmark a habit complete for a specific date (mark is idempotent; future dates allowed — Q9)
- [x] Archive / unarchive a habit — completion history preserved (Q11)
- [x] Completion count for the last 30 days (display)
- [x] Current streak = consecutive completed days ending today or yesterday; one miss resets to zero

**Scope boundary:** daily only — no weekly/monthly cadence, weekday selection, grace
days, freezes, or vacation mode. A completion has no quantity or note.

### 5. Goals — ⬜ M5

- [ ] Create / edit / delete a goal (title, optional description, optional target date)
- [ ] Set manual four-state progress: not started / in progress / achieved / abandoned
- [ ] See the list of goals with progress state

**Scope boundary:** progress is manual only — no percentages, no roll-up from tasks /
habits / time, no goal hierarchy, no milestones or check-in history.

### 6. Reviews (daily + weekly) — ⬜ M6

- [ ] Daily review — fixed prompt set, free-text answers, edit + view past (pending Q1)
- [ ] Daily review shows that date's totals (time per category, habits completed) for reference
- [ ] Weekly review — fixed prompt set, free-text, per ISO week, edit + view past (pending Q2)
- [ ] Weekly review shows that week's totals (time, habit counts, tasks that entered `DONE`) for reference

**Scope boundary:** prompt sets are fixed and non-editable; free text only, no ratings or
scores; reference data is display-only; a review produces no follow-up items.

### 7. Deterministic Productivity Reports — ⬜ M7

- [ ] Time by category (over a chosen date range)
- [ ] Planned vs actual by category — planned, actual, difference
- [ ] Habit completion — completed days and rate per habit
- [ ] Task throughput — tasks that entered `DONE` within the range
- [ ] Daily actual totals — total actual time per day in the range
- [ ] All calculations deterministic and explainable; DST-crossing ranges correct

**Scope boundary:** this list is exhaustive for V1 — no custom report builder, arbitrary
group-by, saved reports, range-to-range comparison, trends, or forecasts.

### 8. Data Export — ⬜ M8

- [ ] One user-initiated export of all account-owned data (categories, planned + actual blocks, tasks, habits + completions, goals, daily + weekly reviews)
- [ ] Delivered as a download in a single open, documented format (pending Q3)
- [ ] Round-trips: the file contains every entity the user created
- [ ] Contains only the caller's own data

**Scope boundary:** full snapshot only — no scheduled, partial, or filtered export; no
import; no delivery to a third party.

### 9. Frontend — 🔨 (spans milestones)

- [x] React + TypeScript + Vite SPA, served single-origin by the Go process (`go:embed`)
- [x] Login UI · Registration UI
- [x] Authenticated shell — account, change timezone, change password, log out
- [x] Route guards; auth state resolved from `GET /api/account` on load
- [x] Categories UI · daily timeline UI · planned-vs-actual view (M2)
- [x] Kanban board — 4 columns, task cards, create/edit/delete, move via select + drag (M3)
- [x] Habits — date nav, streak, per-date toggle, archive/unarchive (M4)
- [ ] Habits (M4)
- [ ] Goals (M5)
- [ ] Reviews (M6)
- [ ] Reports (M7)
- [ ] Data export UI (M8)
- [~] Responsive — CSS targets 375px & 1280px; visual acceptance pass pending

**Scope boundary:** no native app, offline mode, or installable PWA in V1.

### 10. Backend & Infrastructure Foundation — ✅ M1 (core)

- [x] Go · standard-library `net/http` · hand-written middleware · modular monolith · one process
- [x] JSON HTTP API under `/api`, no version segment, shared error envelope
- [x] PostgreSQL · pgx/pgxpool · sqlc · explicit SQL · no ORM · no sqlx
- [x] `golang-migrate` · forward-only migrations (embedded in the binary)
- [x] Docker Compose for local PostgreSQL · Makefile
- [x] Go unit + real-PostgreSQL integration tests · mandatory cross-account isolation tests
- [x] CI: build, vet, lint, tests + Postgres service, `sqlc diff`, frontend typecheck + build
- [ ] Production hosting / database provider / backup mechanism → **ADR-0008** (after a free-tier spike)
- [ ] A database restore performed successfully once (V1 Definition of Done)

### V1 Core Loop

`PLAN → DO → TRACK → REVIEW → IMPROVE`

### Explicitly OUT of V1

Collaboration · teams · social · sharing · native mobile apps · offline-first · PWA ·
notifications · calendar sync · recurring tasks · AI planning · AI insights · intelligent
scheduling · goal hierarchy · advanced gamification · Kubernetes · Kafka · RabbitMQ ·
service mesh · Elasticsearch · Redis (unless a demonstrated need arises).

---

## V2 — Assisted Productivity

**Progress: 0 / 12 areas.** Still deterministic, still no social features. **Every V2
feature must receive its own approved specification before implementation.**

### 1. Calendar — ⬜

- [ ] Calendar-oriented view of planned activities
- [ ] Better date navigation
- [ ] Calendar-style visualization of time blocks
- [ ] External calendar synchronization (needs its own V2 spec)

### 2. Focus Sessions — ⬜

- [ ] Start / stop / pause a focus session
- [ ] Record focus duration; associate with a task/activity
- [ ] Focus-session history
- [ ] Compare planned focus time with actual focus time

### 3. Pomodoro — ⬜

- [ ] Work intervals · short breaks · long breaks
- [ ] Session counter · focus history
- [ ] Optional task association · configurable durations where justified

### 4. Recurring Tasks — ⬜

- [ ] Recurring task definitions (daily, weekly, more if justified)
- [ ] Task occurrences with per-occurrence completion tracking
- [ ] Historical completion preservation
- [ ] Safe handling of edits, missed occurrences, and recurrence changes

### 5. Goal Hierarchy — ⬜

- [ ] Parent / child goals · decomposition · milestones
- [ ] Better goal-progress visualization
- [ ] Optional relationships between goals and tasks/habits
- [ ] Automatic progress roll-up (needs its own spec)

### 6. Learning Tracker — ⬜

- [ ] Subjects / topics · learning sessions · learning time · history
- [ ] Progress tracking · relationship between learning and goals

### 7. Quick Capture — ⬜

- [ ] Rapid task / activity / note capture
- [ ] Keyboard-friendly, minimal interaction path ("capture faster than avoiding capture")

### 8. Notifications — ⬜

- [ ] Task / habit / planned-activity / review / focus-session reminders
- [ ] Notification channels + delivery infrastructure (needs its own spec)

### 9. Achievements / Lightweight Gamification — ⬜

- [ ] Habit / task / focus / learning consistency milestones
- [ ] No manipulative loops, fake scores, or activity farming — supportive, not central

### 10. Richer Weekly Analysis — ⬜

- [ ] Weekly productivity summary · planned-vs-actual analysis
- [ ] Time-allocation, habit-consistency, task-throughput, goal-progress trends
- [ ] Previous-week comparisons · recurring measurable patterns (facts vs interpretation kept distinct)

### 11. Productivity Insights — ⬜

- [ ] Recurring plan-vs-actual gaps · frequently delayed categories
- [ ] Time-allocation and habit-consistency patterns · repeated planning failures
- [ ] Grounded in stored data and explainable

### 12. Assisted Planning — ⬜

- [ ] Suggest a daily plan; suggest where tasks could fit
- [ ] Consider tasks, goals, habits, available time, and historical planned-vs-actual
- [ ] Suggest realistic allocations; help rebalance overloaded days; explain suggestions
- [ ] AI suggests; the user approves. AI never silently modifies the schedule.

---

## V3+ — Intelligent

Adaptive AI-assisted planning and insight, deeper automation. The product becomes
smarter only after it is trustworthy: `track reality → understand patterns → assist
planning → careful automation`.

---

## Reference

### V1 → V2 at a glance

| Capability | V1 | V2 |
|---|---|---|
| Authentication | core auth, sessions, isolation | extend only if needed |
| Timeline | planned + actual time | calendar / focus workflows |
| Tasks | basic tasks | recurring tasks |
| Kanban | fixed board | extend only if justified |
| Habits | daily habits + streaks | richer analysis / reminders |
| Goals | flat / manual | goal hierarchy |
| Reviews | daily / weekly reflection | richer weekly analysis |
| Reports | deterministic metrics | trends and patterns |
| Data Export | full snapshot | extended for new V2 data |
| Calendar · Focus · Pomodoro · Learning · Quick Capture · Notifications · Achievements | — | yes |
| Productivity Insights | basic deterministic metrics | advanced pattern detection |
| AI Planning | — | assisted planning |

### Development order

`Foundation → Authentication → Timeline → Tasks + Kanban → Habits + Streaks → Goals →
Reviews → Reports → Data Export` — then V2.

### Feature development rule

No feature goes straight from this roadmap to code:

`roadmap feature → detailed requirement (v1.md) → milestone plan (planning.md) →
implement → test → review → security review where applicable → mark complete here`.

### Long-term direction

Deterministic foundation → evidence-based insights → assisted intelligence → careful
automation. The product should become smarter only after it becomes trustworthy.
