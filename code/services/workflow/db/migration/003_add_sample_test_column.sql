-- Test migration to verify the migration system is working
-- This adds a test column to workflow_migrations table itself
 
ALTER TABLE workflow_migrations ADD COLUMN IF NOT EXISTS test_column VARCHAR(50) DEFAULT 'test_value'; 