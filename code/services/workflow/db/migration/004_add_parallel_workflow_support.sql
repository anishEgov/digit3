-- Migration 004: Add parallel workflow support

-- Add parallel workflow columns to process_instances table
ALTER TABLE process_instances 
ADD COLUMN parent_instance_id UUID REFERENCES process_instances(id),
ADD COLUMN branch_id VARCHAR(64),
ADD COLUMN is_parallel_branch BOOLEAN DEFAULT FALSE;

-- Create index for efficient parallel instance queries
CREATE INDEX idx_process_instances_parallel ON process_instances(tenant_id, entity_id, process_id, is_parallel_branch);
CREATE INDEX idx_process_instances_branch ON process_instances(tenant_id, entity_id, process_id, branch_id);

-- Table for tracking parallel execution coordination
CREATE TABLE parallel_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(64) NOT NULL,
    entity_id VARCHAR(64) NOT NULL,
    process_id UUID NOT NULL REFERENCES processes(id) ON DELETE CASCADE,
    parallel_state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,
    join_state_id UUID NOT NULL REFERENCES states(id) ON DELETE CASCADE,
    active_branches JSONB NOT NULL DEFAULT '[]',
    completed_branches JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(32) DEFAULT 'ACTIVE',
    created_by VARCHAR(64) NOT NULL,
    created_at BIGINT NOT NULL,
    modified_by VARCHAR(64) NOT NULL,
    modified_at BIGINT NOT NULL,
    UNIQUE(tenant_id, entity_id, process_id, parallel_state_id)
);

-- Create indexes for efficient parallel execution queries
CREATE INDEX idx_parallel_executions_entity ON parallel_executions(tenant_id, entity_id, process_id);
CREATE INDEX idx_parallel_executions_status ON parallel_executions(tenant_id, status); 