#!/bin/sh

set -e

echo "run db migration"
/bin/migrate -path /app/db/migration -database "$DB_SOURCE" up

cd  /app

echo "start the app"
exec "$@"