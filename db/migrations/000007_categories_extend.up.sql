-- MX1 Phase 1: categories become a shared, cross-cutting concept (ADR-0009). Each
-- category gains a colour and an icon. Both are a key from a fixed client-side set,
-- not a value the backend interprets — hence plain nullable text (NULL = unset).

ALTER TABLE categories ADD COLUMN colour text;
ALTER TABLE categories ADD COLUMN icon   text;
