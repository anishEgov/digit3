CREATE TABLE notification_template (
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
    UNIQUE (templateid, tenantid, version)
);