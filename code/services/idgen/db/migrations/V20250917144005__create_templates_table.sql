-- Migration: 001_create_templates_table.sql
-- Description: Create idgen_templates table for storing ID generation templates

CREATE TABLE IF NOT EXISTS idgen_templates (
    id          VARCHAR(64) PRIMARY KEY,
    config      JSONB NOT NULL,
    created_at  BIGINT,
    created_by  VARCHAR(64)
);
