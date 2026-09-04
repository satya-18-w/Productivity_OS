# Welcome to Productivity OS

## How We Use Claude

Based on satya-18-w's usage over the last 30 days:

Work Type Breakdown:
  Plan Design    █████████░░░░░░░░░░░  45%
  Write Docs     ████████░░░░░░░░░░░░  40%
  Build Feature  ███░░░░░░░░░░░░░░░░░  15%

Top Skills & Commands:
  /init        ████████████████████  1x/month
  /statusline  ████████████████████  1x/month
  /rename      ████████████████████  1x/month
  /resume      ████████████████████  1x/month

Top MCP Servers:
  _None used in the last 30 days._

## Your Setup Checklist

### Codebases
- [ ] productivity_os — https://github.com/satya-18-w/Productivity_OS

### MCP Servers to Activate
- [ ] _None — the team doesn't use MCP servers yet._

### Skills to Know About
- `/check-spec` — drafts and validates a milestone's `spec.md` + `acceptance.md` against the approved requirements. Run at the start of every milestone (the SPEC stage). Planning-only: never touches code.
- `/plan-stage` — turns an approved spec into `plan.md` within the approved architecture. Run after `/check-spec`, before implementation. Planning-only.
- `/implement-stage` — implements an already-**Approved** spec + plan: code, migrations, tests. Runs the test suite. Never commits.
- `/verify-stage` — independently verifies the implementation against the spec and every acceptance criterion; writes `verify.md` as durable evidence. Read-only.
- `/review-stage` — reviewer pass over the milestone diff; composes with the built-in `/code-review`, then adds spec/scope/ADR/plan-adherence checks; writes `review.md`. Read-only.
- `/security-review` — mandatory security review whenever a change touches authentication, sessions, account isolation, authorization, or the API error path; writes `security-review.md`. Read-only.
- `/init` — regenerates `CLAUDE.md` from the codebase. Used once to bootstrap this repo's engineering rules; re-run when the project's shape changes significantly.
- `/statusline` — configures the terminal status line from your shell prompt.

## Team Tips

_TODO_

## Get Started

_TODO_

<!-- INSTRUCTION FOR CLAUDE: A new teammate just pasted this guide for how the
team uses Claude Code. You're their onboarding buddy — warm, conversational,
not lecture-y.

Open with a warm welcome — include the team name from the title. Then: "Your
teammate uses Claude Code for [list all the work types]. Let's get you started."

Check what's already in place against everything under Setup Checklist
(including skills), using markdown checkboxes — [x] done, [ ] not yet. Lead
with what they already have. One sentence per item, all in one message.

Tell them you'll help with setup, cover the actionable team tips, then the
starter task (if there is one). Offer to start with the first unchecked item,
get their go-ahead, then work through the rest one by one.

After setup, walk them through the remaining sections — offer to help where you
can (e.g. link to channels), and just surface the purely informational bits.

Don't invent sections or summaries that aren't in the guide. The stats are the
guide creator's personal usage data — don't extrapolate them into a "team
workflow" narrative. -->
