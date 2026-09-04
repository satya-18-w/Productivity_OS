Productivity OS — V1 & V2 Feature Roadmap

Product Goal

PLAN → DO → TRACK → REVIEW → IMPROVE

Productivity OS is a personal productivity system for planning time, tracking actual activity, managing tasks and habits, tracking goals, reviewing progress, and understanding productivity through deterministic analytics before introducing intelligent assistance.

V1 — Core Productivity OS

1. Authentication & Account

User registration

User login

User logout

Server-side sessions

Opaque HttpOnly authentication cookie

Authenticated request context

Account-scoped authorization

Strict per-account data isolation

Argon2id password hashing

Password validation/policy

Session expiration and invalid-session handling

Password change

Duplicate-account handling

Structured authentication errors

account_id must come from the authenticated session, never from trusted client input

2. Daily Timeline & Time Tracking

Planned time

Users can plan arbitrary time blocks, e.g.:

09:00–11:00 — DSA

11:00–13:00 — College

15:00–19:00 — Gym

21:00–02:00 — Backend Development

Actual time

Users can record what they actually did.

Requirements

Planned and actual blocks are separate records

Arbitrary [start,end) ranges

Blocks may cross midnight

Hour-by-hour display is a UI representation

Optional flat user-defined categories

Categories can be archived

No hard delete for categories

No meaningful category-color requirement

Planned-vs-actual comparison

Surface gaps/variance between plan and reality

Account timezone controls date/week interpretation

Store timestamps as timestamptz

Centralized timezone/date bucketing

3. Tasks

Create tasks

View tasks

Update tasks

Complete tasks

Reopen tasks where applicable

Task title/details

Task status

Task ordering where required

Kanban board

Fixed columns:

BACKLOG

TODO

IN_PROGRESS

DONE

Constraints:

Columns are fixed

No user-created/deleted columns

Kanban is a view over tasks

No collaboration/sharing

Recurring tasks are not V1

4. Habits & Streaks

Habits

Create habits

Daily cadence only

Mark a habit complete for a day

View completion history

Calculate completion percentage

Streaks

Consecutive completed days

Active streak ends today or yesterday

One missed day resets the streak

No grace days

No weekly/monthly cadence in V1

The system favors truthful tracking over artificial gamification.

5. Goals

Goal fields:

Title

Optional description

Optional target date

Manual four-state progress

Constraints:

Manual progress

No automatic roll-up

No automatic task/habit linkage

Flat goals in V1

Goal hierarchy is V2

6. Reviews

Daily review

Fixed, non-editable prompt set

Free-text answers

View previous answers

Display-only reference data

No score

Weekly review

Weekly review capability

Fixed prompt approach

Free-text reflection

No productivity score

The exact prompt sets and UI follow the approved V1 specification.

7. Deterministic Productivity Reports

V1 contains five deterministic report categories:

Time per category — day/week breakdown

Planned vs actual variance

Habit completion percentage

Task throughput

Productivity overview based on measurable metrics

Rules:

Calculations are deterministic

Metrics must be explainable

Do not pretend to infer psychology

No unsupported AI-generated productivity claims

8. Data Export

User-owned data export

Open format

Export V1-owned data, including applicable:

Timeline data

Planned time

Actual time

Tasks

Habits

Habit history

Goals

Reviews

Categories

Exact export structure is defined by the V1 export specification.

9. Frontend

Responsive web application

React

TypeScript

Vite

Authentication shell

Login UI

Registration UI

Authenticated application shell

Daily timeline

Task list

Kanban board

Habits

Goals

Reviews

Reports

Data export UI

Mobile and desktop responsive behavior

Acceptance testing at 375px and 1280px

10. Backend & Infrastructure Foundation

Go

Standard-library net/http

Hand-written middleware

Modular monolith

One Go application process

JSON HTTP API

/api prefix

No API version segment in V1

PostgreSQL

pgx/pgxpool

sqlc

Explicit SQL

No ORM

No sqlx

golang-migrate

Forward-only migrations

Docker Compose for local PostgreSQL

Makefile

Go unit/integration testing

Real PostgreSQL integration tests

Mandatory cross-account isolation tests

Frontend/E2E verification where required

CI verification for build, vet, lint, tests, PostgreSQL integration, sqlc verification, and frontend build

V1 Core Loop

PLAN
  ↓
DO
  ↓
TRACK
  ↓
REVIEW
  ↓
IMPROVE

Explicitly OUT of V1

Collaboration

Teams

Social features

Sharing

Native mobile apps

Offline-first

PWA requirements

Notifications

Calendar synchronization

Recurring tasks

AI planning

AI productivity insights

Intelligent scheduling

Goal hierarchy

Advanced gamification

Kubernetes

Kafka

RabbitMQ

Service mesh

Elasticsearch

Redis unless a demonstrated need arises

V2 — Assisted Productivity

V2 expands the deterministic V1 foundation with richer planning, focus, recurring work, learning, notifications, analysis, and initial intelligent assistance.

1. Calendar

Calendar-oriented view of planned activities

Better date navigation

Calendar-style visualization of time blocks

Potential external calendar synchronization

External calendar synchronization must receive an explicit V2 specification before implementation.

2. Focus Sessions

Start focus session

Stop/pause focus session

Record focus duration

Associate focus session with a task/activity

Focus-session history

Compare planned focus time with actual focus time

3. Pomodoro

Work intervals

Short breaks

Long breaks

Session counter

Focus history

Optional task association

Configurable durations where justified

4. Recurring Tasks

Recurring task definitions

Daily recurrence

Weekly recurrence

Additional recurrence rules if justified

Task occurrences

Completion tracking per occurrence

Historical completion preservation

Safe handling of edits, missed occurrences, and recurrence changes

5. Goal Hierarchy

Parent goals

Child goals

Goal decomposition

Goal milestones

Better goal-progress visualization

Optional relationships between goals and tasks/habits

Automatic progress roll-up must be explicitly specified before implementation.

6. Learning Tracker

Potential capabilities:

Subjects/topics

Learning sessions

Learning time

Learning history

Progress tracking

Relationship between learning and goals

7. Quick Capture

Reduce capture friction:

Rapid task capture

Rapid activity capture

Rapid note/input flow

Keyboard-friendly capture

Minimal interaction path

Principle:

Capture should be faster than avoiding capture.

8. Notifications

Potential notifications:

Task reminders

Habit reminders

Planned activity reminders

Review reminders

Focus-session reminders

Notification channels and delivery infrastructure must be specified before implementation.

9. Achievements / Lightweight Gamification

Potential achievements:

Habit consistency milestones

Task completion milestones

Focus milestones

Learning milestones

Constraints:

No manipulative engagement loops

No fake productivity scores

No meaningless activity farming

Achievements remain supportive rather than central

10. Richer Weekly Analysis

Expand V1 deterministic reports:

Weekly productivity summary

Planned-vs-actual analysis

Time allocation trends

Habit consistency trends

Task throughput trends

Goal progress trends

Previous-week comparisons

Recurring measurable patterns

Facts must remain distinguishable from interpretation.

11. Productivity Insights

Move from simple metrics toward evidence-based pattern detection:

Recurring plan-vs-actual gaps

Frequently delayed task categories

Time-allocation patterns

Habit consistency patterns

Repeated planning failures

Areas where plans systematically diverge from reality

Insights should be grounded in actual stored data and explainable.

12. Assisted Planning

Initial intelligent assistance:

Suggest a daily plan

Suggest where tasks could fit

Consider tasks, goals, habits, and available time

Use historical planned-vs-actual behavior

Suggest realistic time allocations

Help rebalance overloaded days

Explain planning suggestions

Principle:

AI suggests; the user remains in control.

AI must not silently modify the user's schedule.

V2 Feature Specification Gate

Every V2 feature must receive its own approved specification before implementation.

In particular:

Calendar synchronization

Focus-session model

Pomodoro rules

Recurrence semantics

Goal hierarchy

Learning tracker

Quick capture UX

Notifications

Achievements

Weekly analysis

Productivity insights

AI-assisted planning

AI provider/model integration

V1 → V2 Summary

Capability

V1

V2

Authentication

Core auth, sessions, isolation

Extend only if needed

Timeline

Planned + actual time

Calendar/focus workflows

Tasks

Basic tasks

Recurring tasks

Kanban

Fixed board

Extend only if justified

Habits

Daily habits + streaks

Richer analysis/reminders

Goals

Flat/manual

Goal hierarchy

Reviews

Daily/weekly reflection

Richer weekly analysis

Reports

Deterministic metrics

Trends and patterns

Planned vs Actual

Core V1 feature

Deeper analysis

Data Export

V1

Extend for new V2 data

Calendar

—

Yes

Focus Sessions

—

Yes

Pomodoro

—

Yes

Learning Tracker

—

Yes

Quick Capture

—

Yes

Notifications

—

Yes

Achievements

—

Yes, carefully

Productivity Insights

Basic deterministic metrics

Advanced pattern detection

AI Planning

—

Assisted planning

Goal Hierarchy

—

Yes

Recurring Tasks

—

Yes

Development Order

V1
│
├── Foundation
├── Authentication
├── Timeline
├── Tasks + Kanban
├── Habits + Streaks
├── Goals
├── Reviews
├── Reports
└── Data Export
        │
        ▼
V2
│
├── Calendar
├── Focus Sessions
├── Pomodoro
├── Recurring Tasks
├── Goal Hierarchy
├── Learning Tracker
├── Quick Capture
├── Notifications
├── Achievements
├── Richer Weekly Analysis
├── Productivity Insights
└── Assisted Planning

Feature Development Rule

No feature should go directly from this roadmap to implementation.

Use:

Roadmap Feature
      ↓
Detailed Requirement
      ↓
Milestone Specification
      ↓
Approved Specification
      ↓
Implementation Plan
      ↓
Human Approval
      ↓
Implementation
      ↓
Verification
      ↓
Code Review
      ↓
Security Review where applicable
      ↓
Acceptance
      ↓
Commit

Long-Term Direction

The product should become smarter only after it becomes trustworthy.

V1
Track reality
      ↓
V2
Understand patterns + assist planning
      ↓
V3+
Adaptive intelligence + deeper automation

The core philosophy is:

Deterministic foundation → evidence-based insights → assisted intelligence → careful automation