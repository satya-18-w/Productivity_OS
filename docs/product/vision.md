# Productivity OS — Product Vision

> Status: stable. This document changes rarely. It describes why the product exists
> and the direction it grows in. It contains no feature list and no V1 scope — those
> live in `docs/requirements/`.

## Purpose

Productivity OS is a personal productivity application. It helps one person plan their
time, record what they actually did, and understand the difference between the two.
Every user works only with their own data; the product has no collaborative, shared,
or social dimension.

## The problem

People who direct their own time — students, self-taught developers, anyone holding
several serious commitments at once — form their intentions in one place (their head,
a note, a calendar) and live their day somewhere else entirely. At the end of a week
they cannot answer basic questions: where did the time actually go, did the habits
hold, which plans survived contact with reality.

The existing tools each see only part of this. Calendars capture intent but are blind
to what happened. Task managers capture lists but are blind to time. Time trackers
capture logs but are blind to planning. Nothing holds the plan and the actuals next to
each other so a person can learn from the gap.

## Target user

A single individual with a dense, self-directed schedule and multiple parallel
commitments in the same day — for example focused study, college, physical training,
and personal project work. They will spend a small amount of effort planning and
reviewing, but only if recording what happened is almost effortless. They want an
accurate picture of their time, not encouragement.

## What success looks like

- The user can answer "where did my week go?" in under a minute.
- For any past day, the user can see what they planned beside what they did.
- The user knows the current streak for every habit without doing any arithmetic.
- A daily review takes a couple of minutes; a weekly review takes under fifteen.
- The user trusts the numbers, including in a bad week.

## Product horizons

Productivity OS grows in three stages. Each stage is built on a working version of the
one before it. Horizons describe direction, not commitments — only the current version
is specified in `docs/requirements/`.

### Horizon 1 — Manual

The user plans, logs, tracks, and reviews entirely by hand. The system stores,
organizes, and reports. It does not suggest, predict, or automate. No AI.

### Horizon 2 — Assisted

The system reduces manual effort with reusable structure: day-plan templates,
recurring tasks, importing time from an external calendar, and a wider set of
deterministic analytics. Still rule-based; the system still does not make decisions
for the user.

### Horizon 3 — Intelligent

The system offers AI-assisted planning and productivity insight: proposed schedules,
detected patterns, and recommendations. The user accepts or rejects each one; the
system never changes a plan unattended.

## Permanent non-goals

Out of scope at every horizon, not just the first:

- **Collaboration and social features.** No sharing, teams, accountability partners,
  comments, activity feeds, or public profiles. Multi-user support exists only to give
  each person a private, isolated account.
- **Being a calendar client.** It will not replace Google Calendar or Outlook, nor
  manage meetings, invitations, or availability.
- **Being a note-taking or journaling app.** Review notes are structured and
  purpose-built, not a free-form notebook.
- **Being a team or organizational tool.** No org accounts, roles, seats, or admin
  hierarchies.
- **Gamification beyond streaks.** No points, badges, levels, leaderboards, or rewards.

## Glossary

The canonical name and meaning for each core concept. Every other document, and later
the code, uses these terms.

| Term | Meaning |
|---|---|
| **Account** | One person's private space. Every piece of data belongs to exactly one account. |
| **Category** | A user-defined label for a kind of activity (e.g. "DSA", "Gym"). A flat list, no hierarchy. |
| **Time block** | A span of time with a start and an end, optionally assigned to one category. |
| **Planned block** | A time block representing intent for a date. |
| **Actual block** | A time block representing what really happened. Recorded independently of the plan. |
| **Day plan** | All planned blocks for a single calendar date. |
| **Timeline** | A single date's blocks shown against the hours of that day. |
| **Task** | A unit of work to be done, with a lifecycle state. |
| **Board** | The Kanban view of an account's tasks, grouped by state. |
| **Habit** | A behaviour the user intends to perform every day. |
| **Streak** | A run of consecutive days on which a habit was completed. |
| **Goal** | A longer-term intention, tracked manually. |
| **Daily review** | A short structured reflection on one day. |
| **Weekly review** | A longer structured reflection on one week. |
| **Report** | A read-only, deterministic aggregation of an account's data over a date range. |
