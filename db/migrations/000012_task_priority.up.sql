-- MX3-follow-up: nullable priority on tasks — a plain label with no effect on
-- ordering or board layout (v1.md §7 amendment, approved 2026-09-05).
ALTER TABLE tasks ADD COLUMN priority text CHECK (priority IN ('HIGH', 'MEDIUM', 'LOW'));
