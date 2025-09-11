# ========= BUILD FRONTEND =========
FROM --platform=linux/arm64 node:24-alpine AS frontend-build
WORKDIR /frontend

ARG APP_VERSION=dev
ENV VITE_APP_VERSION=$APP_VERSION

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./

RUN if [ ! -f .env ] && [ -f .env.production.example ]; then \
      cp .env.production.example .env; \
    fi

RUN npm run build

# ========= BUILD BACKEND =========
FROM --platform=linux/arm64 golang:1.23.3 AS backend-build
WORKDIR /app

# Install Go tools for ARM64
RUN go install github.com/pressly/goose/v3/cmd/goose@latest \
    && go install github.com/swaggo/swag/cmd/swag@v1.16.4

COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy frontend build into backend
RUN mkdir -p /app/ui/build
COPY --from=frontend-build /frontend/dist /app/ui/build

COPY backend/ ./
RUN swag init -d . -g cmd/main.go -o swagger

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
    go build -o /app/main ./cmd/main.go

# ========= RUNTIME =========
FROM --platform=linux/arm64 debian:bookworm-slim

ARG APP_VERSION=dev
LABEL org.opencontainers.image.version=$APP_VERSION
ENV APP_VERSION=$APP_VERSION

# Install PostgreSQL clients (13–17) and runtime deps
RUN apt-get update \
 && apt-get install -y --no-install-recommends wget ca-certificates gnupg lsb-release sudo gosu \
 && echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" \
    > /etc/apt/sources.list.d/pgdg.list \
 && wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor \
    > /etc/apt/trusted.gpg.d/postgresql.gpg \
 && apt-get update \
 && apt-get install -y --no-install-recommends \
      postgresql-17 \
      postgresql-client-13 \
      postgresql-client-14 \
      postgresql-client-15 \
      postgresql-client-16 \
      postgresql-client-17 \
 && rm -rf /var/lib/apt/lists/*

# Create data dir and set ownership (postgres user already exists)
RUN mkdir -p /postgresus-data/pgdata \
    && chown -R postgres:postgres /postgresus-data

WORKDIR /app

# Copy goose and app binary from build stage
COPY --from=backend-build /go/bin/goose /usr/local/bin/goose
COPY --from=backend-build /app/main .
COPY --from=backend-build /app/ui/build ./ui/build
COPY --from=backend-build /app/swagger ./swagger
COPY backend/migrations ./migrations

# Copy env file if present
COPY backend/.env* /app/
RUN if [ ! -f /app/.env ] && [ -f /app/.env.production.example ]; then \
      cp /app/.env.production.example /app/.env; \
    fi

# Startup script (no rogue \q, DSN format for goose)
COPY <<'EOF' /app/start.sh
#!/bin/bash
set -e

PG_BIN="/usr/lib/postgresql/17/bin"

echo "Setting up data directory permissions..."
mkdir -p /postgresus-data/pgdata
chown -R postgres:postgres /postgresus-data

if [ ! -s "/postgresus-data/pgdata/PG_VERSION" ]; then
  echo "Initializing PostgreSQL database..."
  gosu postgres $PG_BIN/initdb -D /postgresus-data/pgdata --encoding=UTF8 --locale=C.UTF-8
  echo "host all all 127.0.0.1/32 md5" >> /postgresus-data/pgdata/pg_hba.conf
  echo "local all all trust" >> /postgresus-data/pgdata/pg_hba.conf
  echo "port = 5437" >> /postgresus-data/pgdata/postgresql.conf
  echo "listen_addresses = 'localhost'" >> /postgresus-data/pgdata/postgresql.conf
  echo "shared_buffers = 256MB" >> /postgresus-data/pgdata/postgresql.conf
  echo "max_connections = 100" >> /postgresus-data/pgdata/postgresql.conf
fi

echo "Starting PostgreSQL..."
gosu postgres $PG_BIN/postgres -D /postgresus-data/pgdata -p 5437 &
POSTGRES_PID=$!

echo "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
  if gosu postgres $PG_BIN/pg_isready -p 5437 -h localhost >/dev/null 2>&1; then
    echo "PostgreSQL is ready!"
    break
  fi
  if [ $i -eq 30 ]; then
    echo "PostgreSQL failed to start"
    exit 1
  fi
  sleep 1
done

echo "Setting up database and user..."
gosu postgres $PG_BIN/psql -p 5437 -h localhost -d postgres <<'SQL'
ALTER USER postgres WITH PASSWORD 'Q1234567';
SQL

# Create the database using createdb (avoids running CREATE DATABASE inside a function)
if ! gosu postgres $PG_BIN/psql -p 5437 -h localhost -tAc "SELECT 1 FROM pg_database WHERE datname='postgresus'" | grep -q 1; then
  echo "Creating database 'postgresus'..."
  gosu postgres $PG_BIN/createdb -p 5437 -h localhost -O postgres postgresus || {
    echo "Failed to create database 'postgresus'"
    exit 1
  }
else
  echo "Database 'postgresus' already exists"
fi

echo "Running database migrations..."
# Use goose with explicit driver and DB string; also export env vars so goose can pick them up if it expects
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgres://postgres:Q1234567@localhost:5437/postgresus?sslmode=disable"
# Run goose with env vars set (goose recognizes GOOSE_DRIVER and GOOSE_DBSTRING)
/usr/local/bin/goose -dir ./migrations up

echo "Starting Postgresus application..."
exec ./main
EOF

RUN chmod +x /app/start.sh

EXPOSE 4005
VOLUME ["/postgresus-data"]

CMD ["/app/start.sh"]
