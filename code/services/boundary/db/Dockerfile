# Migration Dockerfile using Flyway
FROM flyway/flyway:11.13.0

# Set working directory
WORKDIR /flyway

# Copy SQL migration files to Flyway's default location
COPY migrations/ /flyway/sql/

# Copy Flyway configuration file
COPY config/flyway-docker.conf /flyway/conf/flyway.conf

# Set default Flyway configuration (now handled by flyway.conf)
# ENV variables removed - using config file approach

# Create a script to wait for database and run migrations
RUN echo '#!/bin/bash\n\
set -e\n\
\n\
# Wait for database to be ready\n\
echo "Waiting for database to be ready..."\n\
while ! nc -z ${DB_HOST:-localhost} ${DB_PORT:-5432}; do\n\
  echo "Database not ready, waiting..."\n\
  sleep 2\n\
done\n\
echo "Database is ready!"\n\
\n\
# Construct Flyway URL\n\
FLYWAY_URL="jdbc:postgresql://${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-postgres}"\n\
\n\
# Run Flyway migrations\n\
echo "Running database migrations..."\n\
flyway \\\n\
  -configFiles="/flyway/conf/flyway.conf" \\\n\
  -url="$FLYWAY_URL" \\\n\
  -user="${DB_USER:-postgres}" \\\n\
  -password="${DB_PASSWORD:-postgres}" \\\n\
  migrate\n\
\n\
echo "Migrations completed successfully!"\n\
' > /flyway/migrate.sh && chmod +x /flyway/migrate.sh

# Install netcat for database readiness check and ensure proper permissions
USER root
RUN apt-get update && apt-get install -y netcat-openbsd && rm -rf /var/lib/apt/lists/*

# Ensure flyway directory has proper permissions and run as root to avoid user issues
RUN chmod +x /flyway/migrate.sh
RUN chown -R root:root /flyway

# Run as root to avoid user permission issues
USER root

# Default command
CMD ["/flyway/migrate.sh"] 