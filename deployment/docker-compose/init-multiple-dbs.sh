#!/bin/bash
set -e

# Default superuser from POSTGRES_USER
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
  -- Create databases
  CREATE DATABASE account;
  CREATE DATABASE registry;
  CREATE DATABASE workflow;

  -- Create users
  CREATE USER account_user WITH PASSWORD 'account_pass';
  CREATE USER registry_user WITH PASSWORD 'registry_pass';
  CREATE USER workflow_user WITH PASSWORD 'workflow_pass';

  -- Grant privileges on databases
  GRANT ALL PRIVILEGES ON DATABASE account TO account_user;
  GRANT ALL PRIVILEGES ON DATABASE registry TO registry_user;
  GRANT ALL PRIVILEGES ON DATABASE workflow TO workflow_user;
EOSQL

# Assign schema ownership inside each DB
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname account <<-EOSQL
  ALTER SCHEMA public OWNER TO account_user;
  GRANT ALL ON SCHEMA public TO account_user;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname registry <<-EOSQL
  ALTER SCHEMA public OWNER TO registry_user;
  GRANT ALL ON SCHEMA public TO registry_user;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname workflow <<-EOSQL
  ALTER SCHEMA public OWNER TO workflow_user;
  GRANT ALL ON SCHEMA public TO workflow_user;
EOSQL