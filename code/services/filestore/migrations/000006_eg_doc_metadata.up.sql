CREATE TABLE eg_doc_metadata (
    id bigint NOT NULL,
    tenantId character varying(256) NOT NULL,
    type character varying(256) NOT NULL,
    code character varying(256) NOT NULL,
    allowedFormats Jsonb,
    minSize character varying(256),
    maxSize character varying(256),
    isSensitive boolean,
    description character varying(256) not null,
    isActive boolean,
    auditDetail Jsonb
);


ALTER TABLE eg_doc_metadata ADD CONSTRAINT pk_doc_metadata PRIMARY KEY (id); 
