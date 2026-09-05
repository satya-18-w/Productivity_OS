-- MX-TL: a time block may optionally reference a task instead of carrying its own
-- category. ON DELETE SET NULL: deleting a task clears the link on its blocks
-- without deleting them. The CHECK constraint enforces category inheritance
-- structurally — a task-linked block never stores its own category (v1.md §3/§4/§10
-- of the analysis doc); its effective category is resolved by joining to the task.
ALTER TABLE time_blocks ADD COLUMN task_id uuid REFERENCES tasks (id) ON DELETE SET NULL;
ALTER TABLE time_blocks ADD CONSTRAINT time_blocks_category_xor_task
    CHECK (task_id IS NULL OR category_id IS NULL);
CREATE INDEX time_blocks_account_task_idx ON time_blocks (account_id, task_id);
