-- MX3 Phase 2: nullable goal_id on tasks — many tasks per goal, one goal per task
-- (v1.md §10 amendment). ON DELETE SET NULL implements "deleting a goal clears the
-- link on its tasks without deleting them or blocking the delete" directly at the
-- DB layer.
ALTER TABLE tasks ADD COLUMN goal_id uuid REFERENCES goals (id) ON DELETE SET NULL;
CREATE INDEX tasks_account_goal_idx ON tasks (account_id, goal_id);
