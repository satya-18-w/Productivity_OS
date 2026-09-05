# Backend work left — endpoints the frontend needs

> The frontend is built screen-by-screen ahead of some backend endpoints. Where an
> endpoint doesn't exist yet, the screen is wired to a clearly-marked mock in the
> frontend and a checkpoint is added here. When the endpoint lands, swap the mock for
> the real call at the noted file.
>
> **Rule:** every item here is either (a) a real V1 requirement the frontend can't fully
> satisfy without it, or (b) a performance optimisation, or (c) a reference-only
> affordance the frontend deliberately does NOT build until a product requirement exists.
> Each item says which.

---

## Phase 6 — Habits

### ☐ `GET /api/habits/history?from=<ISO>&to=<ISO>`  — **(a) required for the "This Month" view**

The Habits screen's **This Month** heatmap needs per-habit completion dates over a date
range. The existing `GET /api/habits?date=` only returns one date's status plus a
`last_30_days` *count*.

**Response shape the frontend expects:**
```json
{
  "from": "2026-08-01",
  "to": "2026-09-04",
  "habits": [
    { "id": "h1", "name": "Workout", "completions": ["2026-08-01", "2026-08-03", ...] }
  ]
}
```
- `completions` = ISO dates (account timezone) the habit was marked complete, within
  `[from, to]`. Active habits only (or include archived with an `archived: true` flag —
  frontend currently only shows active).
- Range is resolved in the account timezone (N4). Bound the range server-side (e.g. ≤ 92
  days) and 400 on anything larger.

**Frontend swap point:** `web/src/features/habits/habitData.ts` →
`mockHabitHistory(...)` is the placeholder. Replace its body with a call to
`api.habitHistory(from, to)` and add that method to `web/src/api.ts`. The
`<HabitMonthHeatmap>` view shows a "sample data — history endpoint pending" note until
this is done.

### ☐ `GET /api/habits/week?date=<any-day-in-week>`  — **(b) optimisation, not blocking**

The **This Week** grid currently makes **7 parallel `GET /api/habits?date=` calls**
(`fetchWeek()` in `habitData.ts`). A single endpoint returning the ISO week's per-habit
per-day completion would replace them.

**Response shape:**
```json
{
  "week_start": "2026-08-31",
  "days": ["2026-08-31", ..., "2026-09-06"],
  "habits": [
    { "id": "h1", "name": "Workout", "current_streak": 12,
      "completed": ["2026-08-31", "2026-09-01", ...] }
  ],
  "archived": [{ "id": "h9", "name": "Old habit" }]
}
```
**Frontend swap point:** `web/src/features/habits/habitData.ts` → `fetchWeek()`. Works
today with 7 calls; drop them for one call when this exists.

### ☐ `PATCH /api/habits/:id` (rename) and `DELETE /api/habits/:id`  — **(c) not V1, confirm before building**

`v1.md §9` grants create / mark / unmark / archive / unarchive / view — **not** rename or
delete. The Habits screen's kebab menu therefore offers **Archive only** (see
`HabitActions` in `web/src/features/habits/HabitBits.tsx`). If product later wants
rename/delete, add the endpoints + `api.renameHabit` / `api.deleteHabit` and extend that
component's menu items.

---

## Phase 8 — Categories

### ☐ List archived categories + `POST /api/categories/:id/unarchive`  — **(c) not V1, confirm before building**

`v1.md §2` guarantees only **archive** ("the user can archive a category… without
changing blocks already assigned to it"). It does not require an archived-categories view
or unarchive. The existing endpoints (`listCategories`, `createCategory`,
`renameCategory`, `archiveCategory`) give no way to list archived categories or bring one
back — unlike Habits, where `GET /api/habits` already returns `archived:
ArchivedHabit[]` and `unarchive` exists.

**Frontend status:** the Categories screen (`web/src/features/categories/`) shows the
active list only, with Rename / Archive. No "Archived" tab was built — there's nothing to
show it from. This is also open design-system item **C1**
(`docs/design/design-system.md §6.2`) — whether categories store anything beyond a name.

**If product wants archived categories to be visible/reversible:** add
`GET /api/categories?state=archived` (or an `?include_archived=true` flag on the existing
list, mirroring `HabitList.archived`) and `POST /api/categories/:id/unarchive`. Frontend
swap point: `web/src/features/categories/CategoriesScreen.tsx` — add a fetch for the
archived set and an "Archived" `SegmentedControl` tab (the pattern already exists in
`features/habits/HabitAllList.tsx`).

---

## Phase 9 — Reports

### ☐ `GET /api/reports?from=<ISO>&to=<ISO>`  — **(a) required for the entire screen**

**No reports backend exists at all** — confirmed via `grep -n "report" web/src/api.ts`
(no matches) and a search of the Go `internal/` tree (no `report` module). The whole
`/reports` screen (`v1.md §13`'s five fixed reports) currently runs on deterministic
mock data. This is the largest single backend gap in the project so far.

`v1.md §13` fixes the report set — no ad-hoc report builder, no saved views, no
period-over-period comparison, no trend lines. One endpoint returning all five
sub-reports for a chosen `[from, to]` is enough; the frontend does not need them
split into five requests.

**Response shape the frontend expects:**
```json
{
  "from": "2026-08-01",
  "to": "2026-09-04",
  "time_by_category": [
    { "category_id": "c1", "category_name": "Deep Work", "seconds": 72000 },
    { "category_id": null, "category_name": "Uncategorized", "seconds": 5400 }
  ],
  "planned_vs_actual": [
    { "category_id": "c1", "category_name": "Deep Work", "planned_seconds": 90000, "actual_seconds": 72000 }
  ],
  "habit_completion": [
    { "habit_id": "h1", "habit_name": "Workout", "completed_days": 21, "range_days": 35 }
  ],
  "task_throughput": 14,
  "daily_actual_totals": [
    { "date": "2026-08-01", "seconds": 14400 }
  ]
}
```
Notes for the implementer:
- **Q7 (sum of durations):** `time_by_category[].seconds` and `daily_actual_totals[].seconds`
  must be the sum of actual (logged) time, computed the same way as the Timeline's actual
  blocks — reuse that aggregation rather than re-deriving it.
- **Q8 (explicit Uncategorized bucket):** a block/task with no category must appear under
  an explicit `category_id: null, category_name: "Uncategorized"` row, not be silently
  dropped from the total.
- **`planned_vs_actual`** excludes the Uncategorized bucket (planned time is only
  meaningful per named category) — the frontend computes `difference = actual - planned`
  itself, so the endpoint need not include it.
- **`habit_completion.range_days`** is the number of days in `[from, to]` the habit was
  *active* (i.e. existed and wasn't archived) — not necessarily the full range length if
  the habit was created partway through it.
- **`task_throughput`** = count of tasks whose status entered `DONE` within `[from, to]`
  (`v1.md §13`, "tasks completed in range" — not tasks merely due in range).
- **N4 (timezone/DST correctness):** resolve `from`/`to` in the account's timezone, same as
  the Habits history endpoint (`Phase 6` above). Bound the range server-side (e.g. ≤ 366
  days) and 400 on anything larger.
- Sort order is the endpoint's choice; the frontend does its own display sort (categories
  by `seconds` descending, days chronologically) so it doesn't matter functionally.

**Frontend swap point:** `web/src/features/reports/reportsData.ts` → `mockReportData(from,
to)` is the whole placeholder. Replace its call site in
`web/src/features/reports/ReportsScreen.tsx` (`useMemo(() => mockReportData(from, to), ...)`)
with `api.reports(from, to)` (new method + `Report`-shaped types in `web/src/api.ts`), and
delete `mockReportData`/`seededRandom` from `reportsData.ts` once it's unused. The
"⚠ Sample data" notice in `ReportsScreen.tsx` should be removed in the same change. Every
report sub-component (`TimeByCategoryReport.tsx`, `PlannedVsActualReport.tsx`,
`HabitCompletionReport.tsx`, `TaskThroughputReport.tsx`, `DailyActualTotalsReport.tsx`)
already consumes the exact shapes above via `ReportData`'s sub-interfaces, so no
presentation changes should be needed — only the data source.

---

## Phase 10 — Daily Review

### ☐ `GET /api/reviews/daily?date=<ISO>` and `PUT /api/reviews/daily?date=<ISO>`  — **(a) required for the entire screen**

**No reviews backend exists at all** — confirmed via `grep -n "review" web/src/api.ts`
(no matches) and a search of the Go `internal/` tree (no `review` module). The whole
`/reviews/daily` screen (`v1.md §11`) runs on an **in-memory mock store** (resets on page
reload — enough to demo create/edit/view within a session, unlike Reports' pure
random-generator mock, because a review is something the user actually writes and expects
to see again).

**`GET` response shape the frontend expects** (404 / `null` body if nothing saved yet for
that date):
```json
{
  "date": "2026-08-15",
  "answers": {
    "wentWell": "Shipped Phase 10",
    "notPlanned": "Underestimated the reference-panel layout",
    "differently": "Start with the data shape before the UI",
    "grateful": "A quiet afternoon"
  },
  "updated_at": "2026-08-15T21:40:00Z"
}
```
**`PUT` request body:** `{ "answers": { ...same four keys } }` → same response shape back.

Notes for the implementer:
- The **four prompt keys** (`wentWell`, `notPlanned`, `differently`, `grateful`) are the
  frontend's internal names for Q1's fixed wording — pick whatever field names the backend
  prefers, the frontend's `DAILY_REVIEW_PROMPTS` constant is the source of truth for the
  *display* text and can be remapped at the swap point (below) either way.
- Free text only — no ratings, scores, or structured fields (§11 scope boundary). No length
  limit is specified; the frontend caps each field at 5000 chars client-side only.
- **No "not after today" check** — a review may be created/edited for a future date (Q9).
  Do not add server-side date validation beyond a well-formed ISO date.
- A `PUT` should **upsert** (create if absent, overwrite if present) — the frontend always
  calls the same save path for both "Save review" (new) and "Save changes" (edit); it
  doesn't distinguish create vs. update at the API level.

**Not needed:** the reference-totals panel ("actual time by category" + "habits
completed") is **already real** — it calls the existing `api.comparison(date)` and
`api.habits(date)`, same as Timeline and Habits. Only the review record itself is mocked.

**Frontend swap point:** `web/src/features/reviews/reviewData.ts` — `fetchDailyReview` and
`saveDailyReview` are the whole placeholder (backed by a module-level `Map`). Replace both
with `api.dailyReview(date)` / `api.saveDailyReview(date, answers)` calls (new methods +
`DailyReview`-shaped types in `web/src/api.ts`), and delete the `Map` once nothing calls the
old functions. `DailyReviewScreen.tsx` imports only the two functions + `emptyAnswers` +
`DAILY_REVIEW_PROMPTS`, so no other file needs to change.

---

## Phase — Timeline Week/Month (G2, 2026-09-05)

### ✅ `GET /api/timeline/range?from=<ISO>&to=<ISO>` — done (2026-09-05, later)

Backend built exactly the shape specced below (its own code comment cites this doc).
Frontend swap done: `web/src/api.ts` gained `timelineRange(from, to)` +
`RangeTimeline`; `WeekView.tsx` and `MonthView.tsx` each now fire one
`api.timelineRange` call instead of 7 / up to 42 parallel `api.timeline(date)` calls,
indexing the returned `days` by date. Verified live — Network tab shows exactly one
`GET /api/timeline/range?from=2026-08-31&to=2026-09-06` for Week and one
`...&to=2026-10-11` for Month, both 200, both rendering real blocks including a
task-linked one. Tests updated (`WeekView.test.tsx`, `MonthView.test.tsx`,
`TimelineScreen.test.tsx`) to mock `timelineRange`; full suite green (289 tests).

### ☐ ~~`GET /api/timeline/range?from=<ISO>&to=<ISO>`~~  — **(b) optimisation, not blocking**

Week and Month (`design-system.md` G2, `v1.md §5` amendment) render **real data** today —
no mock, no fixture — but inefficiently: `WeekView` fires **7 parallel**
`GET /api/timeline?date=` calls (one per day of the ISO week) and `MonthView` fires **up
to 42** (every visible day, including the leading/trailing days from adjacent months
needed to fill the calendar grid). Both work correctly; this is the same trade-off
already accepted for Habits' week grid (`docs/left.md`, Phase 6).

**Response shape the frontend expects**, if this is built:
```json
{
  "from": "2026-08-31",
  "to": "2026-09-06",
  "days": [
    { "date": "2026-08-31", "planned": [ /* PositionedBlock[] */ ], "actual": [ /* … */ ] }
  ]
}
```
Same `PositionedBlock` shape `GET /api/timeline?date=` already returns per day, just
batched. Resolve in the account timezone (N4), same as the single-date endpoint. Bound
the range server-side (e.g. ≤ 62 days, enough for a month grid's 42 cells) and 400 on
anything larger.

**Frontend swap point:** `web/src/features/timeline/WeekView.tsx` and `MonthView.tsx` —
each has its own `useEffect` that does `Promise.all(dates.map((d) => api.timeline(d)))`.
Replace both with one `api.timelineRange(from, to)` call (new method on `web/src/api.ts`)
and index the returned `days` array by date instead. No other file changes — both
components already isolate their own fetch from the rest of `TimelineScreen`.

### ☐ Category `colour`/`icon` — adjacent finding, not this phase's responsibility

`v1.md §2` (ADR-0009) already gives categories a real `colour` and `icon`, and the
backend fully stores/returns both (`internal/categories`). **The frontend has never been
updated to consume either field anywhere** — `web/src/api.ts`'s `Category` type is still
`{id, name}`, and every category-coloured surface (Timeline blocks/chips, the Categories
screen's tile) uses `categoryColor(id)`, a deterministic **hash**, not the user's actual
chosen colour. This is a frontend gap, not a backend one — recorded here because it was
re-discovered while building Timeline Week/Month (which also colour blocks via the
hash). Properly fixing it means: extending `Category` in `api.ts`; deciding the fixed
colour-key → CSS-token mapping and icon set (a Categories-screen design decision); and
updating the Categories create/edit form to offer them. Out of scope for Timeline —
flagged for a dedicated Categories-screen pass. See
`docs/design/screens/timeline-implementation-plan.md` for where this was first noted.

---

## Phase — Task ↔ Time Block linking (2026-09-05)

### ☐ A time block referencing a task — **(c) not V1, requires a requirements amendment first**

**No backend support exists at all** — no `task_id` column on `time_blocks`, no Go
field, no API surface. `docs/architecture/task-timeblock-model-analysis.md` is the full
analysis (schema, API, service, validation, analytics-impact, migration, exact `v1.md`
sections to amend) — written because a frontend task assumed this relationship was
already "finalized" and buildable; it is not. Its bottom line: **the model is sound and
buildable, but is new product scope not yet in `v1.md`, and needs product-owner approval
of the 4 questions in that doc's §17 before any code (backend or frontend) is written.**

**Frontend status:** nothing built against this — no task-picker on `BlockDialog`, no
"linked task" display on a block, no "this task's blocks" list on a task. Building any
of these now would mean inventing a persistence layer that doesn't exist, which this
project's standing rule (and this task's own explicit instruction) forbids. Once
approved and the API in the analysis doc's §5 lands, the exact frontend integration
points are already enumerated in that doc's §2.4 — no re-discovery needed.

### Update (2026-09-05, later same day) — backend implementation has started

While the frontend Category-on-Task work in this same phase was being browser-QA'd,
a concurrent process began implementing the backend side described above, **without
the §17 approval this doc calls for**: migration `000014_timeblock_task_link` (adds
`time_blocks.task_id`, nullable, `ON DELETE SET NULL`, with a
`CHECK (task_id IS NULL OR category_id IS NULL)` — matching the doc's "mutually
exclusive" recommendation), `Block.TaskID` + a `TaskChecker` interface in
`internal/timeline/timeline.go`, and `task_id` on the block HTTP request/response in
`internal/timeline/http.go`. `go build ./...` and `go test ./internal/tasks/...
./internal/timeline/... ./internal/categories/...` all pass as of this note.

**Still not safe to build the frontend against:** the dev backend process running
during this session's QA (PID 186148, started 04:11) predates all of the above
(files last touched ~05:20–05:25) and was not restarted against it; it's unconfirmed
whether migration `000014` has been applied to the dev database; and the source was
observed mid-edit (a test file briefly failed to build, then was fixed a minute
later), so treat it as unstable until a session deliberately picks this up, confirms
the migration is applied, restarts the backend, and re-verifies the four §17
questions were actually answered rather than bypassed. The frontend
`Block`/`NewBlock` types in `web/src/api.ts` and `BlockDialog.tsx` still have no
`task_id` field — that part is genuinely unbuilt, not just untested.

### Update (2026-09-05, later still) — confirmed stable; frontend built

The product owner confirmed the backend session above is their own intentional,
ongoing work (not a rogue process) — the §17-bypass note above is superseded by that.
Re-verified before building: `go build ./...` clean, `go test
./internal/{tasks,timeline,categories,account}/...` all pass, the dev backend was
restarted (new PID) and `GET /api/timeline` now genuinely returns `"task_id":null`
on live blocks, confirming migration `000014` is applied to the dev DB.

Built against it: `web/src/api.ts` (`Block`/`NewBlock` gained `task_id: string |
null`); `BlockDialog` (Task `<Select>` — picking a task disables the Category field,
replacing it with the task's own category read-only, and clears `category_id` on
submit per the API's mutual-exclusivity rule; an edit-mode "Linked to `<task>` →"
link to `/tasks?openTask=<id>`); `TimelineGrid` + `AgendaList` (a task-linked block
shows "↳ `<task title>`" instead of its category name, plus a thicker left border as
a structural association cue, and Agenda additionally gets a sibling "Open task →"
link); `TasksScreen` (reads `?openTask=<id>`, opens that task's edit dialog once,
then clears the param — the deep-link target for the two links above).

**Not built — a real backend gap, not a frontend omission:** there is still no
`GET /api/tasks/{id}/blocks` (or equivalent) to list a task's scheduled blocks
across dates. `internal/timeline` only exposes per-date (`GET /api/timeline`) and
range (`GET /api/comparison`) reads, plus an unwired `CountBlocksByTask` query used
internally for category-total analytics. So "when viewing a Task, show its
scheduled/planned blocks" (the reverse direction of the two links just built) is not
implemented — scanning every date client-side to find a task's blocks isn't
practical. Needs a dedicated endpoint on the backend side before this can be built.

### ✅ Landed as `GET /api/tasks/{id}/blocks` (2026-09-05, later) — different shape than spec'd below

The backend session picked this up fast, but built a **different route and response
shape** than the spec below — the first of the three contracts handed off this
session where that happened (the task-link fields and the `/api/timeline/range`
batching endpoint both landed byte-for-byte as spec'd; this one didn't). For the
record, and so the next divergence gets caught faster:

- Route: `GET /api/tasks/{id}/blocks` (path param, in `internal/timeline/http.go`'s
  `Mount`), not `GET /api/blocks?task_id=` as proposed below. Reasonable either way —
  this reads fine as "a task's sub-resource" too.
- Response: `{"blocks": [...]}`, not `{"task_id": ..., "blocks": [...]}` — no
  top-level `task_id` echo. Harmless, the frontend already had the id from its own
  request.
- Errors: a genuinely-missing/foreign task_id → `404 Not Found` (via the existing
  `pathID`/`ErrTaskNotFound` helpers already used by every other `{id}`-path block
  route), not the `400 VALIDATION_ERROR` proposed below. Also a reasonable, already-
  established convention — just a different one than this doc guessed at.
- **The one substantive gap:** each block in `blocks` is the plain `blockBody` shape
  — raw UTC `starts_at`/`ends_at` only. It does **not** include the `local_date`/
  `local_start`/`local_end` convenience fields this doc asked for, unlike every other
  block-reading endpoint (`GET /api/timeline`, `GET /api/timeline/range`). Those
  three fields exist precisely so the frontend never has to convert a UTC instant to
  the account's wall-clock date/time itself (ADR-0005 — "the client never does tz
  math"). Category resolution for a task-linked block (verified in
  `BlocksForTask`/`readmodel.go`) is correct — this is only about the missing local
  wall-clock fields.

**Frontend workaround shipped anyway** (rather than block on another round-trip):
`web/src/components/date/dateUtils.ts` gained `utcInZone(isoUtc, timeZone)`, which
converts a UTC instant to an ISO date + 24h time in a *given* IANA zone via
`Intl.DateTimeFormat`, called with the account's own `timezone`
(`useAuth().account`) — not the browser's ambient zone. This is a narrow, documented
exception to "the client never does tz math": it's display-only (nothing computed
this way is ever sent back to the server — every write path still sends wall-clock
values the server resolves itself, unchanged), and it exists only because this one
endpoint's response leaves no other way to show a correct date. **Ideal fix:** add
`local_date`/`local_start`/`local_end` to this endpoint's response the same way
`toPositionedBody` already does for `GET /api/timeline` — trivial, since
`BlocksForTask` already has the account's timezone resolution available via the
same `AccountZone`/`timezone.DayWindow` machinery every other read path uses — and
then delete `utcInZone`'s one call site.

### ☐ `GET /api/blocks?task_id=<uuid>` — original spec (superseded by the above; kept for context)

Lives in `internal/timeline` (it owns `time_blocks`; `internal/tasks` never reaches
into that table directly, matching every existing cross-module boundary here) — a
query-param `GET` added to the same `/api/blocks` collection `POST`/`PUT
.../{id}`/`DELETE .../{id}` already live on, in `internal/timeline/http.go`'s
`Mount`. No new route prefix, no new migration — filters `time_blocks` by
`account_id` (from `reqctx`, never the client) and `task_id`, which is exactly what
index `time_blocks_account_task_idx` (migration `000014_timeblock_task_link`)
already exists for. Validate `task_id` belongs to the caller via the existing
`TaskChecker.AssignableToAccount`; on failure `400 VALIDATION_ERROR {"task_id":
"task not found"}` — the same shape `AddBlock`/`EditBlock` already return, not a new
convention.

**Response shape the frontend expects** (built against this already — see the
"Task shows its scheduled blocks" note further below):
```json
{
  "task_id": "9d98c186-...",
  "blocks": [
    {
      "id": "b1419576-...", "kind": "planned",
      "starts_at": "2026-09-05T08:30:00Z", "ends_at": "2026-09-05T09:30:00Z",
      "category_id": "f4d9f0cd-...", "category_name": "Study",
      "local_date": "2026-09-05", "local_start": "14:00", "local_end": "15:00"
    }
  ]
}
```
Reuses the existing `blockBody` fields plus the three "local wall clock"
convenience fields `positionedBlockBody` already computes for the day view
(`local_date`/`local_start`/`local_end`) — these blocks span arbitrary dates, not
one day, so the day-relative `start_minute`/`end_minute`/`from_prev_day` fields
don't apply and should be omitted. Sort by `starts_at` ascending. Empty `blocks: []`
(not 404) when the task has none linked. Both `kind`s included — the frontend
decides what to surface (it currently shows both).
