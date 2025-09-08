-- Migration 006: Create escalation_configs table
-- Auto-escalation configuration for workflow processes

CREATE TABLE escalation_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(64) NOT NULL,
    process_id UUID NOT NULL REFERENCES processes(id) ON DELETE CASCADE,
    state_code VARCHAR(64) NOT NULL,
    escalation_action VARCHAR(64) NOT NULL,
    state_sla_minutes INTEGER,
    process_sla_minutes INTEGER,
    is_active BOOLEAN DEFAULT true,
    created_by VARCHAR(64) NOT NULL,
    created_at BIGINT NOT NULL,
    modified_by VARCHAR(64) NOT NULL,
    modified_at BIGINT NOT NULL,
    CONSTRAINT escalation_sla_check CHECK (state_sla_minutes IS NOT NULL OR process_sla_minutes IS NOT NULL),
    UNIQUE(tenant_id, process_id, state_code)
);

-- Indexes for performance
CREATE INDEX idx_escalation_configs_tenant_process ON escalation_configs(tenant_id, process_id);
CREATE INDEX idx_escalation_configs_active ON escalation_configs(is_active);
CREATE INDEX idx_escalation_configs_state_code ON escalation_configs(state_code); 