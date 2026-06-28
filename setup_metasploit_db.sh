#!/usr/bin/env bash
set -euo pipefail

worker="${1:-pentagi-terminal-9}"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

docker exec pgvector sh -lc '
set -eu
pw="$(tr -dc A-Za-z0-9 </dev/urandom | head -c 40)"

if [ "$(psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "select 1 from pg_roles where rolname = '\''metasploit'\''")" != "1" ]; then
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "CREATE ROLE metasploit LOGIN PASSWORD '\''$pw'\''" >/dev/null
else
  psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "ALTER ROLE metasploit LOGIN PASSWORD '\''$pw'\''" >/dev/null
fi

if [ "$(psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "select 1 from pg_database where datname = '\''metasploit'\''")" != "1" ]; then
  createdb -U "$POSTGRES_USER" -O metasploit metasploit
fi

psql -U "$POSTGRES_USER" -d metasploit -c "ALTER DATABASE metasploit OWNER TO metasploit" >/dev/null
psql -U "$POSTGRES_USER" -d metasploit -c "GRANT ALL PRIVILEGES ON DATABASE metasploit TO metasploit" >/dev/null

umask 077
cat > /tmp/metasploit_database.yml <<EOF
production:
  adapter: postgresql
  database: metasploit
  username: metasploit
  password: $pw
  host: pgvector
  port: 5432
  pool: 75
  timeout: 5
EOF
'

docker cp pgvector:/tmp/metasploit_database.yml "$tmp"
docker exec "$worker" sh -lc 'mkdir -p /root/.msf4 && chmod 700 /root/.msf4'
docker cp "$tmp" "$worker":/root/.msf4/database.yml
docker exec "$worker" sh -lc 'chmod 600 /root/.msf4/database.yml'
docker exec pgvector rm -f /tmp/metasploit_database.yml

docker exec "$worker" sh -lc 'msfconsole -q -x "db_status; exit"'
