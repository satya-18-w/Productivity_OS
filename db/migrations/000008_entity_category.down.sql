DROP INDEX IF EXISTS goals_account_category_idx;
DROP INDEX IF EXISTS habits_account_category_idx;
DROP INDEX IF EXISTS tasks_account_category_idx;

ALTER TABLE goals  DROP COLUMN category_id;
ALTER TABLE habits DROP COLUMN category_id;
ALTER TABLE tasks  DROP COLUMN category_id;
