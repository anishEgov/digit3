-- Drop the unique index on the uuid column
DROP INDEX IF EXISTS idx_localisation_uuid;

-- Drop the uuid column from the localisation table
ALTER TABLE localisation DROP COLUMN uuid; 