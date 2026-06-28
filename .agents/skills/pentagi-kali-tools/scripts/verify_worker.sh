#!/usr/bin/env bash
set -euo pipefail

worker="${1:-pentagi-terminal-9}"

docker exec "$worker" sh -lc '
set -eu
python3 /tmp/pentagi_tool_smoke.py > /work/rotmg-reports/pentagi-tool-smoke-current.json
/usr/local/bin/verify_pentagi_tooling.sh > /work/rotmg-reports/pentagi-tooling-verification-current.json
python3 - <<PY
import json
smoke="/work/rotmg-reports/pentagi-tool-smoke-current.json"
verify="/work/rotmg-reports/pentagi-tooling-verification-current.json"
s=json.load(open(smoke))
v=json.load(open(verify))
missing=[c["category"]+":"+t["name"] for c in v.get("categories",[]) for t in c.get("tools",[]) if t.get("status")=="missing"]
print("smoke="+smoke)
print("pass="+str(s["summary"]["pass"]))
print("fail="+str(s["summary"]["fail"]))
print("inventory="+verify)
print("missing="+("none" if not missing else ",".join(missing)))
print("mullvad_exit_ip="+str(v.get("mullvad",{}).get("mullvad_exit_ip")))
PY
msfconsole -q -x "db_status; workspace; exit"
for url in http://graphiti:8000/healthcheck http://searxng:8080/ http://langfuse-web:3000/api/public/health https://scraper/; do
  printf "%s=" "$url"
  curl -ksS --max-time 8 "$url" | head -c 100
  echo
done
'

docker exec pgvector sh -lc 'PGPASSWORD="$POSTGRES_PASSWORD" psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Atc "select status,count(*) from flows where deleted_at is null group by status order by status; select id,title,status,model_provider_name,model from flows where deleted_at is null order by id desc limit 3;"'
