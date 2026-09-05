# Misc screens — live vs spec gap plan (2026-09-04)

Source: `docs/design/screens/reviews.md` (V1 scope) + `docs/design/screens/app-shell.md`
(route table) + `docs/design/frontend-implementation-plan.md` Stages 13–15
(Account/Auth/Export layouts) + `docs/requirements/v1.md` §1 (account/auth),
§11 (daily review, Q1 prompts), §12 (weekly, Q2), §14 (export, Q3 open),
Q4 (browser-detected IANA tz, fallback `UTC`), Q6 (password policy), Q9
(future dates allowed) vs live screenshots `/tmp/opencode/pg-reviews-daily.png`,
`pg-account.png`, `pg-export.png`, `pg-login.png`, `pg-register.png`.
Design system: `design-system.md` §4.2 header · §4.5 buttons · §4.7 card ·
§4.16 form (Field/Input/Textarea/Select) + §3 tokens, §5 rules.
No reference images exist for these screens (auth visual language:
`overall.png` panel 12 only — calm centered card).

V1 scope (fixed): daily review = 4 fixed Q1 prompts + live reference panel
(actual time per category + habits completed for that date); weekly review and
export stay `Placeholder` (backends missing — leave them); account = email
display + change password (Q6) + timezone select (Q4) + log out, NO profile
fields; auth = centered `Card`, no shell, Q4/Q6 wired.

## 1. Deviations from spec/refs KEPT (v1.md-justified, do not build)

| Screen | Kept as-is | Justification |
|---|---|---|
| Daily review | Review record on in-memory mock store (resets on reload) | No reviews HTTP endpoints exist (see §3); `docs/left.md` Phase 10 mandates the mock — keep |
| Daily review | Reference panel reads REAL `api.comparison(date)` + `api.habits(date)` | Already real; only the review record is mocked |
| Daily review | No ratings/scores, no generated summary, no history/versioning, no linking to tasks/goals/habits | §11 scope boundary |
| Daily review | No max-date on `DateStepper` (future dates allowed) | Q9 |
| Daily review | No rail (single-column form flow) | `reviews.md` layout (D6/VP3 call, same as every Phase 6+ screen) |
| Weekly review | `Placeholder` route (`phase={11}`) | Weekly backend + weekly reference reads missing (§3) — leave it |
| Export | `Placeholder` route (`phase={14}`) | Q3 format open + no export endpoint (§3) — leave it |
| Account | No profile fields beyond email/password/timezone; no email change, no self-service delete, no MFA | §1 scope boundary |
| Auth | No shell on `/login` / `/register`; calm centered card | `app-shell.md` route table + Stage 14 layout |

## 2. Gaps to build (this pass — frontend scope only)

### Daily review — code hygiene only (no functional gap)

1. **`tl2-form` timeline-layer class on the prompts `<form>`**
   (`DailyReviewScreen.tsx:174`) → new `.review-form` in `styles/reviews.css`
   with the identical token declaration (`display:flex; flex-direction:column;
   gap:var(--sp-4)`). Same rendered values; removes the cross-feature layer
   dependency (DS §5 rules 3–4). `GoalDialog.tsx` also uses `tl2-form` but is
   out of scope — left untouched.
2. Everything else verified present: `PageHeader` (eyebrow "Reviews" + H1 +
   factual subtitle), `DateStepper` (`label="Review date"`, no max-date),
   reference `Card` (chips with category-colour dots, zero-time categories
   filtered, habit checklist with `CheckIcon`), prompts `Card` (4 fixed
   `Field`+`Textarea`, 5000-char cap, "Save review"/"Save changes" + factual
   "Saved" note cleared on edit, `role="alert"` save error), loading/error
   (retry) states, responsive 2-col → 1-col, tokens only, one brand primary.

### Account (`web/src/pages/Account.tsx`) — restyle to primitives

3. **No `PageHeader`** (DS §4.2) — add eyebrow "Account" + H1 "Account" + one
   factual subtitle. (`App.tsx` already wraps the page in `ScreenLayout`; the
   page itself adds the header, not a nested layout.)
4. **Legacy markup** (`.stack` / `.card` / raw `label`+`input`+`button` /
   `.field-error` / `.error` / `.ok`) → `Card`, `Field`, `Input`, `Select`,
   `Button` primitives + tokens only. Sections: Account (email display),
   Change timezone, Change password, Log out.
5. **Timezone is a free-text input** → `Select` of IANA names (Stage 13
   layout; Q4), defaulting to the browser-detected zone. Shared
   `TimezoneSelect` lives in `web/src/features/reviews/TimezoneSelect.tsx`
   (feature-folder-local per scope constraints — **promote to
   `components/ui` later**) and is reused by Register.
6. **Password form has no confirm field** → add new-password confirm with
   match validation (Stage 13 layout: current / new / confirm). Keep Q6 hint
   text, `autocomplete` (`current-password` / `new-password`), field errors
   from the server, `role="alert"` errors, busy labels. Success still signs
   out (`setAccount(null)` + `/login`, server invalidates the session).
7. **No in-screen Log out** (Stage 13 layout lists log out as a section) →
   add a `Card` with a `Button` running the existing `api.logout()` flow
   (same as `UserMenu`: `api.logout()` → `setAccount(null)` → `/login`).
8. **`TimezoneForm` initial state goes stale** (`useState(current)` never
   re-syncs when the account loads) → sync on `current` change.
9. Tests: new `web/src/pages/Account.test.tsx` (email display, timezone
   select persists, confirm-mismatch blocks submit, logout clears session +
   redirects). No profile-field UI may appear (assert absence).

### Auth (`web/src/pages/Login.tsx`, `Register.tsx`) — restyle to primitives

10. **Legacy markup** (`.center` + `.card` + raw controls + `.brand`
    gradient tile — the `linear-gradient(… #d68bff)` second hue violates the
    one-brand-primary rule) → calm centered `Card` (no shell — keep),
    `Field`+`Input` (+`Select` for Register timezone), block primary
    `Button` with `loading`/busy labels, `role="alert"` errors, kept
    `autocomplete`, kept cross-links. Brand lockup = solid `--brand` tile +
    wordmark (tokens only, Grid/Flex, no new colours/spacing/radii).
11. **Q4/Q6 stay wired, no behaviour change**: Register keeps
    `browserTimezone()` default (fallback `UTC`) — now offered as a `Select`;
    Q6 hint text kept; server field errors mapped onto `Field error` +
    `Input invalid`; 429 → "Too many attempts…" copy kept.
12. Tests: new `web/src/pages/Login.test.tsx` (submit + redirect, wrong-password
    message, 429 message) and `web/src/pages/Register.test.tsx` (Q6 hint
    shown, server field error shown, timezone defaults to browser zone).

## 3. Backend gaps noticed (list only — not implemented)

- **Reviews HTTP endpoints not mounted**: `internal/reviews` service exists
  (daily + weekly get/upsert) but `cmd/server/main.go` mounts only
  account/categories/timeline/tasks/habits/goals — no `/api/reviews/*`
  routes. Daily stays on the `reviewData.ts` mock; weekly stays
  `Placeholder`. Swap point unchanged: `fetchDailyReview`/`saveDailyReview`
  → `api.dailyReview`/`api.saveDailyReview` (`docs/left.md` Phase 10).
- **Weekly reference reads missing**: week actuals per category, habit
  completion counts, tasks→`DONE`-in-week (Q10 interpretation) — nothing to
  call from the frontend.
- **Export missing entirely**: no endpoint + Q3 (JSON doc vs CSV archive)
  still open — route stays `Placeholder`.
- (Noticed, pre-existing: reports backend missing per `docs/left.md`
  Phase 9 — out of this pass's scope.)
