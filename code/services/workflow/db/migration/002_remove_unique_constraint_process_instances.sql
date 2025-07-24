-- Remove unique constraint to allow multiple process instance records per entity
-- This enables audit trail by creating new records for each transition
ALTER TABLE process_instances DROP CONSTRAINT process_instances_tenant_id_entity_id_process_id_key;

-- Optional: Add an index to maintain query performance for latest instance lookups
CREATE INDEX idx_process_instances_entity_latest ON process_instances (tenant_id, entity_id, process_id, created_at DESC); 