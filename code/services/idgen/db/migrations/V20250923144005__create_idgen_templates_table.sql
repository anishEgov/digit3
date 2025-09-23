-- Create the idgen_templates table
CREATE TABLE IF NOT EXISTS idgen_templates_v2 (
    templateid VARCHAR(64) NOT NULL PRIMARY KEY,
    config JSONB NOT NULL,
    createdtime BIGINT,
    createdby VARCHAR(64)
)
