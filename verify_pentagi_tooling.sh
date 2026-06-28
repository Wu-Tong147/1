#!/usr/bin/env bash
set -euo pipefail

FIX_MODE=false
OUTPUT_FILE="/work/rotmg-reports/pentagi-tooling-verification.json"

for arg in "$@"; do
    case "$arg" in
        --fix) FIX_MODE=true ;;
        *) OUTPUT_FILE="$arg" ;;
    esac
done

mkdir -p "$(dirname "$OUTPUT_FILE")"

declare -a PROJECT_DISCOVERY=(nuclei subfinder httpx naabu katana dnsx alterx)
declare -a METASPLOIT=(msfconsole msfvenom msfdb msfrpcd)
declare -a NETWORK=(nmap masscan netexec crackmapexec enum4linux impacket-scripts bloodhound.py ldapdomaindump responder)
declare -a WEB=(nikto gobuster dirb dirsearch ffuf zaproxy wpscan sqlmap)
declare -a CREDS=(hydra john hashcat)
declare -a CAPTURE=(tcpdump tshark socat netcat-openbsd netcat-traditional)
declare -a OSINT=(searchsploit)

ensure_on_path() {
    local target="$1"
    local link_name="${2:-$1}"
    if [[ -x "$target" && ! -e "/usr/local/bin/$link_name" ]]; then
        ln -sf "$target" "/usr/local/bin/$link_name" 2>/dev/null || true
    fi
}

fix_tool() {
    local name="$1"
    case "$name" in
        impacket-scripts)
            cat > /usr/local/bin/impacket-scripts <<'EOF'
#!/usr/bin/env bash
if [ $# -eq 0 ]; then
  echo 'Usage: impacket-scripts <script-name> [args...]'
  echo 'Available scripts:'
  ls /usr/bin/impacket-* 2>/dev/null | sed 's|/usr/bin/impacket-||'
  exit 1
fi
script=$1
shift
exec /usr/bin/impacket-$script "$@"
EOF
            chmod +x /usr/local/bin/impacket-scripts
            ;;
        bloodhound.py)
            python3 -m pip install bloodhound 2>/dev/null || true
            rm -f /usr/local/bin/bloodhound.py
            cat > /usr/local/bin/bloodhound.py <<'EOF'
#!/usr/bin/env python3
import subprocess
import sys
sys.exit(subprocess.call([sys.executable, "-m", "bloodhound"] + sys.argv[1:]))
EOF
            chmod +x /usr/local/bin/bloodhound.py
            ;;
        netcat-openbsd)
            apt-get update -qq 2>/dev/null && apt-get install -y -qq netcat-openbsd 2>/dev/null || true
            ensure_on_path /usr/bin/nc netcat-openbsd
            ;;
        netcat-traditional)
            apt-get update -qq 2>/dev/null && apt-get install -y -qq netcat-traditional 2>/dev/null || true
            ensure_on_path /usr/bin/nc.traditional netcat-traditional 2>/dev/null || ensure_on_path /usr/bin/nc netcat-traditional
            ;;
    esac
}

check_tool() {
    local name="$1"
    local path
    path=$(command -v "$name" 2>/dev/null || true)
    if [[ -n "$path" && -x "$path" ]]; then
        printf '{"name":"%s","path":"%s","status":"available"}' "$name" "$path"
        return
    fi

    if [[ "$FIX_MODE" == true ]]; then
        fix_tool "$name" >/dev/null 2>&1 || true
        path=$(command -v "$name" 2>/dev/null || true)
        if [[ -n "$path" && -x "$path" ]]; then
            printf '{"name":"%s","path":"%s","status":"available"}' "$name" "$path"
            return
        fi
    fi

    if dpkg -s "$name" >/dev/null 2>&1; then
        printf '{"name":"%s","path":null,"status":"installed_package"}' "$name"
    else
        printf '{"name":"%s","path":null,"status":"missing"}' "$name"
    fi
}

check_category() {
    local category="$1"
    shift
    local -a tools=("$@")
    local results=()
    for t in "${tools[@]}"; do
        results+=("$(check_tool "$t")")
    done
    local joined
    joined=$(IFS=,; printf '%s' "${results[*]}")
    printf '{"category":"%s","tools":[%s]}' "$category" "$joined"
}

CATEGORIES=()
CATEGORIES+=("$(check_category "projectdiscovery" "${PROJECT_DISCOVERY[@]}")")
CATEGORIES+=("$(check_category "metasploit" "${METASPLOIT[@]}")")
CATEGORIES+=("$(check_category "network" "${NETWORK[@]}")")
CATEGORIES+=("$(check_category "web" "${WEB[@]}")")
CATEGORIES+=("$(check_category "credentials" "${CREDS[@]}")")
CATEGORIES+=("$(check_category "capture_pivot" "${CAPTURE[@]}")")
CATEGORIES+=("$(check_category "osint" "${OSINT[@]}")")
CATEGORIES_JOINED=$(IFS=,; printf '%s' "${CATEGORIES[*]}")

MULLVAD_STATUS=$(curl -sS --max-time 10 https://am.i.mullvad.net/json 2>/dev/null || echo '{}')

JSON=$(cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "hostname": "$(hostname)",
  "container_image": "${PENTAGI_DOCKER_IMAGE:-unknown}",
  "categories": [$CATEGORIES_JOINED],
  "mullvad": $MULLVAD_STATUS
}
EOF
)

printf '%s\n' "$JSON" | tee "$OUTPUT_FILE"
