-- The shared snapshot store (sqlitedb) does not write the variant column; the variant
-- lives in the JSON state.
ALTER TABLE games DROP COLUMN variant;
