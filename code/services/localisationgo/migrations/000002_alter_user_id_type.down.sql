ALTER TABLE localisation ALTER COLUMN created_by TYPE BIGINT USING created_by::bigint;
ALTER TABLE localisation ALTER COLUMN last_modified_by TYPE BIGINT USING last_modified_by::bigint; 