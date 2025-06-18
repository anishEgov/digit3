package postgres

import "fmt"

// Schema defines the SQL for creating the localisation table
const Schema = `


-- Create the localisation table
CREATE TABLE IF NOT EXISTS localisation (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(256) NOT NULL,
    module VARCHAR(256) NOT NULL,
    locale VARCHAR(256) NOT NULL,
    code VARCHAR(256) NOT NULL,
    message TEXT NOT NULL,
    created_by BIGINT,
    created_date TIMESTAMP WITH TIME ZONE NOT NULL,
    last_modified_by BIGINT,
    last_modified_date TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(tenant_id, locale, module, code)
);

-- Add indexes for frequent queries
CREATE INDEX IF NOT EXISTS idx_localisation_tenant_module_locale ON localisation(tenant_id, module, locale);
CREATE INDEX IF NOT EXISTS idx_localisation_tenant_locale ON localisation(tenant_id, locale);
`

// GetSeedDataSQL returns SQL to insert initial sample data
func GetSeedDataSQL() string {
	return fmt.Sprintf(`
    INSERT INTO localisation (tenant_id, module, locale, code, message, created_by, created_date, last_modified_by, last_modified_date)
    VALUES 
    ('DEFAULT', 'rainmaker-dss', 'en_IN', ' DSS_PGR_COMPLETION_RATE', 'Completion Rate', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP),
    ('DEFAULT', 'rainmaker-common', 'en_IN', 'ABG_ABG_GENERATE_NEW_BILL', 'Generate New Bill', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP),
    ('DEFAULT', 'rainmaker-common', 'en_IN', 'ABG_ABG_PAY', 'Pay', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP),
    ('DEFAULT', 'rainmaker-common', 'en_IN', 'ABG_ADVT.Gas_Balloon_Advertisement_CONSUMER_CODE_LABEL', 'Consumer Code', 1, CURRENT_TIMESTAMP, 1, CURRENT_TIMESTAMP)
    ON CONFLICT (tenant_id, locale, module, code) DO NOTHING;
    `)
}
