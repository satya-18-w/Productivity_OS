# Engineering Conventions

> Living document. It records idiomatic implementation guidance — how to write code
> that fits this codebase. Architectural decisions and their rationale live in
> `docs/decisions/` (ADRs); this file points to them rather than repeating them.

## Where the rules live

| Topic | Owned by |
|-------|----------|
| Deployment shape, single-origin, no distributed infrastructure | ADR-0001 |
| Go toolchain, module/package structure, the dependency rule, HTTP layer, API & error contract, validation stance | ADR-0002 |
| pgx driver, `sqlc`, migrations (forward-only), transaction boundaries | ADR-0003 |
| Sessions, password-hashing algorithm, account-scoping mechanism, no RLS in V1 | ADR-0004 |
| Timezone storage and all date / week bucketing | ADR-0005 |
| Frontend stack and single-origin asset serving | ADR-0006 |
| Local dev topology, testing approach, CI | ADR-0007 |
| Spec-before-code, the eight-stage workflow | `CLAUDE.md`; principle E4 |
| ₹0 filter, "boring tech", justify every dependency | `docs/product/principles.md` E1, E3 |

## Idioms to be established in Milestone M1

Filled in here when the M1 skeleton lands. These are style choices *within* the ADRs,
not new decisions:

- Error creation and wrapping idiom — how a `code` / `message` / `fields` error is
  built and propagated to the central writer defined in ADR-0002.
- Structured logging (`log/slog`): field names, levels, what every request logs.
- Naming: packages, exported module interfaces, `sqlc` query files and methods.
- Internal module layering in practice (`service` / `store` / handler) — the shape
  ADR-0002 leaves to convention.
- Test file layout, test naming, and how to run a single test (approach set by
  ADR-0007).
- `gofmt` / `golangci-lint` configuration and how it is run locally.
- Frontend: component structure, styling approach, and any libraries added (per
  ADR-0006, only on demonstrated need). — **established, see "Frontend" below.**

## Frontend

Established during the design-system foundation stage (first frontend implementation).

### Styling

- **One visual system.** Its source of truth is `docs/design/design-system.md`
  (tokens + component contracts) and `docs/design/visual-principles.md`.
- **Plain CSS + CSS custom properties.** No CSS-in-JS, no CSS Modules, no utility
  framework. Layers live in `web/src/styles/` — `tokens.css` (canonical tokens),
  `base.css` (reset), `primitives.css` (`ui-` component styles), and the legacy
  `web/src/styles.css` (feature classes still in use, migrated screen by screen).
  `web/src/styles/README.md` has the detail.
- Components consume **tokens**, never raw values. Adding a colour / spacing / radius /
  shadow / type step means adding a token and getting it approved (project `CLAUDE.md`
  → "Design System Changes").
- Ratified design decisions D1–D6, D8–D10 are applied; several token *values* remain
  `PROVISIONAL` pending the "T1" extraction pass (design-system.md §6.2).

### App shell (D3 — approved 2026-09-04)

- **Three-region shell**: left sidebar (primary nav + brand + user chip) · main content
  (`PageHeader` + body) · right contextual rail (per-screen; a screen may have none).
  CSS Grid. Replaces `web/src/AuthLayout.tsx`. Build spec: `docs/design/screens/app-shell.md`.
- Auth screens (`/login`, `/register`) do **not** use the shell — centered narrow layout.
- Top bar carries a theme toggle + user avatar only. **No global search, no notification
  bell** (not V1).
- Responsive shed order (D4): right rail → sidebar labels (icon-only) → sidebar drawer.
  Main content + the primary action survive at every width; the page never scrolls sideways.

### Routing (D10 — approved 2026-09-04)

React Router, client-side. Authenticated routes render inside the app shell:

| Route | Screen |
|---|---|
| `/` | Timeline (today) — **the landing screen; there is no dashboard** |
| `/timeline` | Timeline (Day / Agenda views via `?view=`) |
| `/tasks` | Tasks (list) |
| `/board` | Board (Kanban) — separate route, same task model as `/tasks` |
| `/habits` | Habits |
| `/goals` | Goals |
| `/categories` | Categories |
| `/reports` | Reports (the five fixed §13 reports) |
| `/reviews/daily` · `/reviews/weekly` | Daily / Weekly review |
| `/account` | Account |
| `/export` | Data export |
| `/login` · `/register` | Auth (no shell) |

Unauthenticated → `/login`; authenticated on an auth route → `/`. Unknown route → `/`.
Routes not listed here are **not** V1 (no `/dashboard`, `/notes`, `/calendar`,
`/timeline/week`, `/timeline/month`).

### Component structure

```
web/src/components/
  ui/            — presentation primitives (Button, Card, Input, Dialog, Tabs, …)
  layout/        — layout primitives (Stack, Inline, Container, Section, PageHeader)
  productivity/  — domain-shaped presentation (StatCard, ListRow, StatusBadge, …)
  date/          — MiniCalendar + local-date helpers
web/src/features/<screen>/  — feature screens: own data fetching + state, compose
                              primitives, rendered by a route in App.tsx
web/src/shell/             — the app shell (AppShell, Sidebar, ScreenLayout, …)
```

`components/**` is **presentation-only** — no data fetching, no business rules. A
**feature** (`features/<screen>/`) owns its API calls and state, wraps its content in
`<ScreenLayout>`, and is the element for its route. Each folder has a barrel `index.ts`.
`web/src/pages/` holds the pre-design-system pages still being migrated screen by screen.

### Testing

- **Vitest + @testing-library/react** (jsdom), added here as the ADR-0007 milestone
  decision ("the frontend grows enough logic to warrant its own unit tests → add a
  frontend test runner"). Config lives in `web/vite.config.ts`; setup in
  `web/src/test/setup.ts`.
- Scope: component **behaviour and accessibility** (roles, labels, keyboard, focus,
  disabled) — not visual appearance. Run with `pnpm test` (in `web/`).
- Test files sit next to the component as `*.test.tsx`.
- CI (ADR-0007) currently runs the frontend typecheck + build; adding `pnpm test` to
  that workflow is a follow-up.
- Real-browser verification stays per-milestone (ADR-0007) and is done ad hoc with
  Playwright, not committed as a suite.

## Commands

To be added when the M1 `Makefile` exists (see ADR-0007 for the intended entry points).
Frontend, from `web/`: `pnpm dev`, `pnpm build`, `pnpm typecheck`, `pnpm test`.
