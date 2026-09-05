# Export format

`GET /api/export` returns a single JSON document containing every entity the
caller's account owns (`v1.md §14`, Q3: resolved as a single JSON document, not a
CSV archive). The response is also served with
`Content-Disposition: attachment; filename="productivity-os-export-<date>.json"`
so a browser downloads it directly.

The export is a point-in-time snapshot — nothing about it is paginated, filtered,
or partial. Re-requesting it later reflects the account's current data.

## Top-level shape

```json
{
  "exported_at": "2026-09-04T18:30:00Z",
  "categories": [ /* Category */ ],
  "planned_blocks": [ /* Block */ ],
  "actual_blocks": [ /* Block */ ],
  "tasks": [ /* Task */ ],
  "habits": [ /* Habit */ ],
  "habit_completions": [ /* HabitCompletion */ ],
  "goals": [ /* Goal */ ],
  "daily_reviews": [ /* DailyReview */ ],
  "weekly_reviews": [ /* WeeklyReview */ ],
  "notes": [ /* Note */ ]
}
```

`exported_at` is an RFC 3339 UTC instant — when this snapshot was produced.

Every array is present even when empty (`[]`, never `null`). Every array contains
**all** of the account's rows for that entity — active and archived alike, where
the entity has an archived state. Nothing is windowed by date.

## Category

```json
{ "id": "uuid", "name": "Deep Work", "colour": "blue", "icon": "brain", "archived_at": null }
```

`colour`/`icon` are opaque keys from a fixed client-side set (`""` when unset) —
the backend never interprets them (ADR-0009). `archived_at` is an RFC 3339 instant,
or `null` while active.

## Block (`planned_blocks` and `actual_blocks`)

```json
{
  "id": "uuid", "starts_at": "2025-06-15T09:00:00Z", "ends_at": "2025-06-15T10:00:00Z",
  "category_id": "uuid | null", "task_id": "uuid | null"
}
```

The two arrays share this shape; which array a block appears in is exactly its
`kind` (`v1.md §3`/`§4`) — `planned_blocks` never contains an actual block or vice
versa. `starts_at`/`ends_at` are RFC 3339 UTC instants. `category_id` joins to
`categories[].id` when set. `task_id` was added 2026-09-05 (task/time-block linkage)
and joins to `tasks[].id` when set — when `task_id` is set, `category_id` is always
`null` in this raw export (a task-linked block never stores its own category; the
consumer should resolve the effective category by looking up `tasks[].category_id`
for the linked task, the same inheritance rule the live API applies server-side).

## Task

```json
{
  "id": "uuid", "title": "Ship export", "description": "", "due_date": "2025-06-20 | null",
  "state": "BACKLOG | TODO | IN_PROGRESS | DONE", "category_id": "uuid | null",
  "goal_id": "uuid | null", "priority": "HIGH | MEDIUM | LOW | null",
  "created_at": "2025-06-15T09:00:00Z", "updated_at": "2025-06-15T09:00:00Z"
}
```

`due_date` is a plain calendar date (`YYYY-MM-DD`), or `null`. `goal_id` and
`priority` were added 2026-09-05 (MX3/MX3-follow-up); both `null` when unset. The
individual state-transition history (`task_transitions`) is **not** exported — it is
an internal audit trail, not a named export entity (`v1.md §14` lists "tasks", not
transition history).

## Habit

```json
{ "id": "uuid", "name": "Meditate", "category_id": "uuid | null", "target": "30 minutes | null", "archived_at": null }
```

`target` was added 2026-09-04 (MX3) — a free-text descriptor, display-only, `null`
when unset.

## HabitCompletion

```json
{ "habit_id": "uuid", "date": "2025-06-15" }
```

One row per completed date, joining to `habits[].id`. Absence of a row for a
(habit, date) pair means not completed — the same "presence = completion" model
the product uses everywhere (`v1.md §9`).

## Goal

```json
{
  "id": "uuid", "title": "Ship M8", "description": "", "target_date": "2025-12-31 | null",
  "progress": "NOT_STARTED | IN_PROGRESS | ACHIEVED | ABANDONED", "category_id": "uuid | null",
  "created_at": "2025-06-15T09:00:00Z", "updated_at": "2025-06-15T09:00:00Z",
  "done_tasks": 1, "total_tasks": 2
}
```

`done_tasks`/`total_tasks` were added 2026-09-05, closing a completeness gap flagged
during MX3-follow-up — they now match `GET /api/goals`'s live derived progress
exactly (`export.TasksReader.ProgressByGoal`, the same method `goals.Handler`'s list
endpoint already uses). Both `0` for a goal with no linked tasks.

## DailyReview

```json
{ "date": "2025-06-15", "answers": { "went_well": "Shipped M8" }, "updated_at": "2025-06-15T09:00:00Z" }
```

`answers` is the `{prompt_key: free_text}` map as saved — only the keys the
account actually answered are present; the fixed prompt text itself
(`v1.md §11`) is not repeated here since it is unrelated to the account's data.

## WeeklyReview

```json
{ "iso_year": 2025, "iso_week": 24, "answers": { "highlights": "M8" }, "updated_at": "2025-06-15T09:00:00Z" }
```

Same `answers` shape as `DailyReview`, keyed by ISO year and week instead of a
date (`v1.md §12`).

## Note

```json
{
  "id": "uuid", "title": "Idea", "body": "write it down",
  "created_at": "2026-09-05T09:00:00Z", "updated_at": "2026-09-05T09:00:00Z"
}
```

Added 2026-09-05 (MX4). Plain text only — no tags, no category, no pin/favourite/
archive state (`v1.md §15`).
