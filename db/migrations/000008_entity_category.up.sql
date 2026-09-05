-- MX1 Phase 3: tasks, habits, and goals may each carry an optional category
-- (ADR-0009). RESTRICT matches time_blocks — categories are never hard-deleted, so
-- this only guards the theoretical case.

ALTER TABLE tasks  ADD COLUMN category_id uuid REFERENCES categories (id) ON DELETE RESTRICT;
ALTER TABLE habits ADD COLUMN category_id uuid REFERENCES categories (id) ON DELETE RESTRICT;
ALTER TABLE goals  ADD COLUMN category_id uuid REFERENCES categories (id) ON DELETE RESTRICT;

CREATE INDEX tasks_account_category_idx  ON tasks  (account_id, category_id);
CREATE INDEX habits_account_category_idx ON habits (account_id, category_id);
CREATE INDEX goals_account_category_idx  ON goals  (account_id, category_id);
