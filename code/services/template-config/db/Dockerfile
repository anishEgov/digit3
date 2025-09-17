# Migration Dockerfile using Flyway
FROM flyway/flyway:11.12.0

# Set working directory
WORKDIR /flyway

# Copy SQL migration files to Flyway's default location
COPY migrations/ /flyway/sql/

# Copy Flyway configuration file
COPY config/flyway-docker.conf /flyway/conf/flyway.conf

# Create a simple migration script for dev environment
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
# Construct Flyway URL and parameters\n\
FLYWAY_URL="jdbc:postgresql://${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME:-postgres}"\n\
FLYWAY_PARAMS="-configFiles=/flyway/conf/flyway.conf -url=$FLYWAY_URL -user=${DB_USER:-postgres} -password=${DB_PASSWORD:-postgres}"\n\
\n\
# Simple dev-friendly migration approach\n\
echo "Running migrations (dev mode)..."\n\
\n\
# First try to repair any issues (safe for dev)\n\
echo "Repairing migration history if needed..."\n\
flyway $FLYWAY_PARAMS repair || true\n\
\n\
# Run migrations (idempotent - safe to run multiple times)\n\
echo "Applying migrations..."\n\
flyway $FLYWAY_PARAMS ${FLYWAY_COMMAND:-migrate}\n\
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

# Override Flyway's entrypoint to run our custom migration script
ENTRYPOINT ["/bin/bash"]
CMD ["/flyway/migrate.sh"]