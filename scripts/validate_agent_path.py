from pathlib import Path
import json
import sys

ROOT = Path(__file__).resolve().parents[1]
errors = []
required = [
    "curriculum/O00/O00.1-safe-system-walkthrough.md",
    "missions/O00-safe-synthetic-walkthrough.md",
    "lab/mission-runtime/go.mod",
    "lab/mission-runtime/cmd/demo/main.go",
    "lab/mission-runtime/cmd/demo/m03_m05.go",
    "lab/mission-runtime/cmd/demo/m06_m07.go",
    "lab/mission-runtime/cmd/demo/mission_runtime_test.go",
    "lab/n8n/M06-readonly-watcher.blueprint.json",
    "lab/n8n/M07-readonly-evidence-agent.blueprint.json",
    "contracts/action-record.schema.json",
    "contracts/outcome-record.schema.json",
    "contracts/advisor-output.schema.json",
    "contracts/improvement-proposal.schema.json",
    "contracts/tool-registry.schema.json",
]
for mission in range(3, 8):
    mid = f"M{mission:02d}"
    if not (ROOT / "curriculum" / mid).is_dir(): errors.append(f"missing curriculum directory: {mid}")
    if not any((ROOT / "missions").glob(f"{mid}-*.md")): errors.append(f"missing mission contract: {mid}")
    if not any((ROOT / "starter-kits").glob(f"{mid}-*/README.md")): errors.append(f"missing starter: {mid}")
    if not any((ROOT / "evals").glob(f"{mid}-*/cases.json")): errors.append(f"missing executable eval pack: {mid}")
for rel in required:
    if not (ROOT / rel).exists(): errors.append(f"missing required file: {rel}")
for rel in ["contracts/action-record.schema.json","contracts/outcome-record.schema.json","contracts/advisor-output.schema.json","contracts/improvement-proposal.schema.json","contracts/tool-registry.schema.json","lab/n8n/M06-readonly-watcher.blueprint.json","lab/n8n/M07-readonly-evidence-agent.blueprint.json"]:
    try: json.loads((ROOT / rel).read_text(encoding="utf-8"))
    except Exception as exc: errors.append(f"invalid JSON {rel}: {exc}")
mission_index = (ROOT / "missions/README.md").read_text(encoding="utf-8")
for mid in ["M03","M04","M05","M06","M07"]:
    line = next((line for line in mission_index.splitlines() if line.startswith(f"| {mid} |")), "")
    if "| ready |" not in line: errors.append(f"{mid} must be ready in mission index")
runtime = "\n".join((ROOT / "lab/mission-runtime/cmd/demo" / name).read_text(encoding="utf-8") for name in ["m03_m05.go","m06_m07.go"])
for marker in ['REJECT_MACHINE_EXECUTION','REJECT_WRITE_REQUEST','REJECT_AUTO_APPLY','REJECT_WRITE_METHOD','REJECT_TOOL','REJECT_UNGROUNDED','DRY_RUN_ONLY']:
    if marker not in runtime: errors.append(f"runtime safety marker missing: {marker}")
if errors:
    print("AGENT PATH VALIDATION FAILED")
    for error in errors: print(f"- {error}")
    sys.exit(1)
print("AGENT PATH VALIDATION PASS: O00 and M03-M07 authoring assets are present and bounded")
