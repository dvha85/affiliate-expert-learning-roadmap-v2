"""Readiness audit: report evidence gaps without promoting status automatically."""
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
plan = (ROOT / "docs/plans/BEGINNER-READINESS-PLAN.md").read_text(encoding="utf-8")
required = {
    "affiliate_source": "chưa có chương trình/kênh",
    "n8n_execution": "UNVERIFIED",
    "pilot": "pilot người mới",
    "live_adapter": "live adapter",
    "runtime": "deployment 24/7",
}
missing = [name for name, marker in required.items() if marker.lower() not in plan.lower()]
if missing:
    raise SystemExit(f"readiness audit missing explicit gap markers: {missing}")
assert "production-ready" not in plan.lower() or "chưa" in plan.lower()
print("READINESS AUDIT: NOT_READY_FOR_PRODUCTION")
for name in required:
    print(f"- {name}: OPEN / evidence required")
print("Fixture, schema, validator and offline smoke PASS do not replace operated/live evidence.")
