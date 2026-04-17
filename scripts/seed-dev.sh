#!/usr/bin/env sh

set -eu

CONTAINER_NAME="${POSTGRES_CONTAINER_NAME:-oportunidades-postgres}"
DATABASE_NAME="${POSTGRES_DB_NAME:-oportunidades}"
DATABASE_USER="${POSTGRES_DB_USER:-postgres}"

docker cp "db/seeds/dev.sql" "$CONTAINER_NAME:/tmp/dev.sql"
docker exec -i "$CONTAINER_NAME" psql -U "$DATABASE_USER" -d "$DATABASE_NAME" -v ON_ERROR_STOP=1 -f /tmp/dev.sql
