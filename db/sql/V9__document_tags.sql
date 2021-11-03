DROP INDEX idx_fname;
ALTER TABLE consensus DROP COLUMN fname;

ALTER TABLE consensus ADD COLUMN tag TEXT;
