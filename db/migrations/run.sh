#!/bin/bash
# db/migrations/run.sh - Run goose migrations
# Usage: ./db/migrations/run.sh up   # Apply all pending migrations
#        ./db/migrations/run.sh down # Rollback last migration
#        ./db/migrations/run.sh status # Check migration status

set -e

MIGRATIONS_DIR="$(dirname "$0")"
DATABASE_URL="${DATABASE_URL:-postgres://cgap:cgap_dev_password@localhost:5432/cgap?sslmode=disable}"

if [ -z "$1" ]; then
  echo "Usage: $0 {up|down|status}"
  exit 1
fi

ACTION=$1

echo "📦 Running migrations: $ACTION"
echo "📍 Migrations dir: $MIGRATIONS_DIR"
echo "🔗 Database: $DATABASE_URL"

goose -dir "$MIGRATIONS_DIR" postgres "$DATABASE_URL" "$ACTION"

echo "✅ Migration complete!"
