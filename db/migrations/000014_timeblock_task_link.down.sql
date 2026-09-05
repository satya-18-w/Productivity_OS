DROP INDEX IF EXISTS time_blocks_account_task_idx;
ALTER TABLE time_blocks DROP CONSTRAINT time_blocks_category_xor_task;
ALTER TABLE time_blocks DROP COLUMN task_id;
