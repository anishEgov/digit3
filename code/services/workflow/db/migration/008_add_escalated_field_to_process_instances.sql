-- Migration 008: Add escalated field to process_instances table
-- Following the Java service pattern for tracking escalated instances

ALTER TABLE process_instances ADD COLUMN escalated BOOLEAN DEFAULT false;

-- Index for better search performance on escalated instances
CREATE INDEX idx_process_instances_escalated_tenant ON process_instances(tenant_id, escalated);
CREATE INDEX idx_process_instances_escalated_business ON process_instances(tenant_id, escalated, process_id); 