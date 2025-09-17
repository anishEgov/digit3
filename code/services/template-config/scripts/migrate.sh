#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values (can be overridden by environment variables)
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-1234}
DB_NAME=${DB_NAME:-template}
FLYWAY_VERSION=${FLYWAY_VERSION:-11.12.0}

# Flyway installation directory
FLYWAY_DIR="./db/tools/flyway"
FLYWAY_BIN="$FLYWAY_DIR/flyway"

echo -e "${YELLOW}🚀 Database Migration Script${NC}"
echo "================================"

# Function to download and install Flyway if not present
install_flyway() {
    if [ ! -f "$FLYWAY_BIN" ]; then
        echo -e "${YELLOW}📥 Flyway not found. Installing Flyway v$FLYWAY_VERSION...${NC}"

        # Create directory
        mkdir -p "$FLYWAY_DIR"

        # Download Flyway
        FLYWAY_URL="https://repo1.maven.org/maven2/org/flywaydb/flyway-commandline/$FLYWAY_VERSION/flyway-commandline-$FLYWAY_VERSION-linux-x64.tar.gz"

        echo "Downloading from: $FLYWAY_URL"
        curl -L "$FLYWAY_URL" | tar xz --strip-components=1 -C "$FLYWAY_DIR"

        # Make executable
        chmod +x "$FLYWAY_BIN"

        echo -e "${GREEN}✅ Flyway installed successfully${NC}"
    else
        echo -e "${GREEN}✅ Flyway already installed${NC}"
    fi
}

# Function to wait for database
wait_for_database() {
    echo -e "${YELLOW}⏳ Waiting for database to be ready...${NC}"

    max_attempts=30
    attempt=1

    while [ $attempt -le $max_attempts ]; do
        if nc -z "$DB_HOST" "$DB_PORT" 2>/dev/null; then
            echo -e "${GREEN}✅ Database is ready!${NC}"
            return 0
        fi

        echo "Attempt $attempt/$max_attempts: Database not ready, waiting..."
        sleep 2
        attempt=$((attempt + 1))
    done

    echo -e "${RED}❌ Database is not ready after $max_attempts attempts${NC}"
    exit 1
}

# Function to run migrations
run_migrations() {
    echo -e "${YELLOW}🔄 Running database migrations...${NC}"

    # Construct database URL
    DB_URL="jdbc:postgresql://$DB_HOST:$DB_PORT/$DB_NAME"

    # Run Flyway migrate using config file
    "$FLYWAY_BIN" \
        -configFiles=db/config/flyway.conf \
        -url="$DB_URL" \
        -user="$DB_USER" \
        -password="$DB_PASSWORD" \
        migrate

    echo -e "${GREEN}✅ Migrations completed successfully!${NC}"
}

# Function to show migration info
show_info() {
    echo -e "${YELLOW}📊 Migration Status:${NC}"

    DB_URL="jdbc:postgresql://$DB_HOST:$DB_PORT/$DB_NAME"

    "$FLYWAY_BIN" \
        -configFiles=db/config/flyway.conf \
        -url="$DB_URL" \
        -user="$DB_USER" \
        -password="$DB_PASSWORD" \
        info
}

# Main execution
main() {
    echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"
    echo "User: $DB_USER"
    echo ""

    # Check if netcat is available for database check
    if ! command -v nc &> /dev/null; then
        echo -e "${YELLOW}⚠️  netcat not found. Skipping database readiness check.${NC}"
        echo "Make sure your database is running before proceeding."
    else
        wait_for_database
    fi

    install_flyway

    # Handle command line arguments
    case "${1:-migrate}" in
        "migrate")
            run_migrations
            ;;
        "info")
            show_info
            ;;
        "help")
            echo "Usage: $0 [migrate|info|help]"
            echo "  migrate  - Run database migrations (default)"
            echo "  info     - Show migration status"
            echo "  help     - Show this help message"
            ;;
        *)
            echo -e "${RED}❌ Unknown command: $1${NC}"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@" 