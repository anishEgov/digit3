-- Create the notification_template table
CREATE TABLE IF NOT EXISTS notification_template (
    id UUID PRIMARY KEY,
    templateid VARCHAR(256) NOT NULL,
    version VARCHAR(256) NOT NULL,
    tenantid VARCHAR(256) NOT NULL,
    type VARCHAR NOT NULL,
    subject TEXT,
    content TEXT NOT NULL,
    ishtml BOOLEAN,
    createdby VARCHAR(64),
    lastmodifiedby VARCHAR(64),
    createdtime BIGINT,
    lastmodifiedtime BIGINT,
    UNIQUE (tenantid, templateid, version)
);

-- Create an index for queries filtering on (tenantid, type)
CREATE INDEX IF NOT EXISTS idx_notification_template_tenantid_type ON notification_template (tenantid, type);