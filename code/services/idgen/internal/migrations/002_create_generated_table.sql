-- Migration: 002_create_generated_table.sql
-- Description: Create idgen_generated table for storing generated IDs

CREATE TABLE IF NOT EXISTS idgen_generated (
    id           SERIAL PRIMARY KEY,
    template_id  VARCHAR(64) NOT NULL,
    generated_id VARCHAR(128) NOT NULL,
    variables    JSONB,
    created_at   BIGINT
);
