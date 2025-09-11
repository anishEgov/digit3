-- Create the template_config table
CREATE TABLE template_config (
    id UUID PRIMARY KEY,
    templateid VARCHAR(256) NOT NULL,
    tenantid VARCHAR(256) NOT NULL,
    version VARCHAR(256) NOT NULL,
    fieldmapping JSONB,
    apimapping JSONB,
    createdby VARCHAR(64),
    lastmodifiedby VARCHAR(64),
    createdtime BIGINT,
    lastmodifiedtime BIGINT,
    UNIQUE (templateid, tenantid, version)
);