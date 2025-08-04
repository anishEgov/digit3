-- Migration 007: Remove is_active column from escalation_configs table
-- Since we use DELETE API for removing configs, isActive field is redundant

ALTER TABLE escalation_configs DROP COLUMN IF EXISTS is_active;

-- Remove the index on is_active since the column is being dropped
DROP INDEX IF EXISTS idx_escalation_configs_active;

-- Update the constraint to remove the default value dependency on is_active
-- (No constraint changes needed since we're not touching SLA constraint) 