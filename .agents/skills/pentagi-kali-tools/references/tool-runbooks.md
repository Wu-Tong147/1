# Tool Runbooks

There is no curated custom runbook set yet beyond this skill. Use these existing sources first.

## PentAGI Local Docs

- `/Users/rep/Documents/ROTMG/pentagi/examples/guides/`
- `/Users/rep/Documents/ROTMG/pentagi/examples/prompts/`
- `/Users/rep/Documents/ROTMG/pentagi/backend/docs/`

Useful files:

- `examples/guides/worker_node.md`
- `examples/guides/openvas-custom-image.md`
- `examples/prompts/base_web_pentest.md`
- `examples/prompts/offline_rotmg_static_review.md`
- `backend/docs/flow_execution.md`
- `backend/docs/docker.md`
- `backend/docs/observability.md`
- `backend/docs/prompt_engineering_pentagi.md`

## Kali Worker Vendor Docs

Inside the worker container:

- Metasploit guides: `/usr/share/metasploit-framework/docs/metasploit-framework.wiki/`
- Metasploit DB example: `/usr/share/metasploit-framework/config/database.yml.example`
- ZAP getting started: `/root/.ZAP/lang/ZAPGettingStartedGuide-2.17.pdf`
- sqlmap docs/examples: `/usr/share/doc/sqlmap/`
- hashcat docs/examples: `/usr/share/doc/hashcat/`, `/usr/share/doc/hashcat-data/examples/`
- john examples: `/usr/share/doc/john/EXAMPLES.gz`
- impacket examples: `/usr/share/doc/python3-impacket/examples/`
- socat examples: `/usr/share/doc/socat/EXAMPLES.gz`

## Verified Tool Groups

- ProjectDiscovery: `nuclei`, `subfinder`, `httpx`, `naabu`, `katana`, `dnsx`, `alterx`
- Metasploit: `msfconsole`, `msfvenom`, `msfdb`, `msfrpcd`
- Network/AD: `nmap`, `masscan`, `netexec`, `crackmapexec`, `enum4linux`, `impacket-scripts`, `bloodhound.py`, `ldapdomaindump`, `responder`
- Web: `nikto`, `gobuster`, `dirb`, `dirsearch`, `ffuf`, `zap.sh`, `wpscan`, `sqlmap`
- Credentials: `hydra`, `john`, `hashcat`
- Capture/pivot: `tcpdump`, `tshark`, `socat`, `nc`
- OSINT/utilities: `searchsploit`, `iptables`, `curl`, `jq`

## Safety Defaults

- Prefer `--help`, `-version`, config, and service health checks.
- Do not run target scans unless the user provides scope.
- Do not print `.env`, provider keys, tokens, cookies, hashes, or raw target secrets.
- Keep one flow active unless explicitly asked to start another.
