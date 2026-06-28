---
name: pentagi-kali-tools
description: Verify and operate the local PentAGI Kali worker toolchain. Use when checking PentAGI/Kali tools, Metasploit DB setup, ProjectDiscovery paths, Graphiti/SearXNG/Langfuse/scraper reachability, Mullvad egress, or when looking for runbooks/examples for the Kali tools in /Users/rep/Documents/ROTMG/pentagi.
---

# PentAGI Kali Tools

## Quick Start

Default worker: `pentagi-terminal-9`.

Run the verifier:

```bash
scripts/verify_worker.sh pentagi-terminal-9
```

It runs the existing worker smoke test, the tooling inventory verifier, Metasploit DB status, service reachability, and flow count. Do not print provider keys or `.env` values.

## Workflow

1. Identify the active worker with `docker ps --format '{{.Names}}\t{{.Image}}\t{{.Status}}'`.
2. Verify the worker before changing anything:
   - `/tmp/pentagi_tool_smoke.py` should report `44 pass / 0 fail`.
   - `/usr/local/bin/verify_pentagi_tooling.sh` should report no missing tools.
   - `msfconsole -q -x "db_status; workspace; exit"` should show PostgreSQL connected.
3. Check path-sensitive tools resolve through `/usr/local/bin`: `nuclei`, `subfinder`, `dnsx`, `ffuf`, `httpx`, `naabu`, `katana`, `alterx`, `responder`.
4. If the active worker is patched, rebuild `local/kali-pentest:latest` from `/Users/rep/Documents/ROTMG/pentagi/Dockerfile.kali` so the next worker inherits the fix.
5. Keep only one PentAGI flow running unless the user explicitly asks otherwise.

## References

Read `references/tool-runbooks.md` when the user asks where runbooks, examples, or tool docs live.

Useful local artifacts:

- Worker smoke report: `/work/rotmg-reports/pentagi-tool-smoke-current.json`
- Inventory report: `/work/rotmg-reports/pentagi-tooling-verification-current.json`
- Custom image: `/Users/rep/Documents/ROTMG/pentagi/Dockerfile.kali`
- Metasploit setup helper: `/Users/rep/Documents/ROTMG/pentagi/setup_metasploit_db.sh`
