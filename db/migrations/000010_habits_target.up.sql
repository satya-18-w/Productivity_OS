-- MX3 Phase 1: an optional target descriptor for a habit (v1.md §9, amended
-- 2026-09-04). Free text ("30 minutes", "8 glasses"), display-only — never
-- validated against a completion; the completion model itself is unchanged.
ALTER TABLE habits ADD COLUMN target text;
