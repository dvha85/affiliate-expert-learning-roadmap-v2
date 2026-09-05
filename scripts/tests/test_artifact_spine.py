"""Run the real validator against isolated template mutations (standard library only)."""

from pathlib import Path
import json
import shutil
import subprocess
import sys
import tempfile
import unittest


ROOT = Path(__file__).resolve().parents[2]
FIELD = "- `action`: `null`"


class ArtifactSpineTemplateTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        shutil.copytree(ROOT / "contracts", self.root / "contracts")
        for relative in (
            "scripts/validate_artifact_spine.py",
            "templates/M00-EVIDENCE-PACKET.md",
            "lab/affiliate-bot/data/m02-sample-observations.json",
            "lab/affiliate-bot/cmd/bot/history.go",
            "lab/affiliate-bot/cmd/bot/history_test.go",
        ):
            target = self.root / relative
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(ROOT / relative, target)
        self.template_path = self.root / "templates/M00-EVIDENCE-PACKET.md"
        self.template = self.template_path.read_text(encoding="utf-8")
        self.assertEqual(self.template.count(FIELD), 1)

    def check_template(self, content, valid):
        self.template_path.write_text(content, encoding="utf-8")
        result = subprocess.run(
            [sys.executable, str(self.root / "scripts/validate_artifact_spine.py")],
            capture_output=True, text=True, check=False,
        )
        self.assertEqual(result.returncode, 0 if valid else 1, result.stdout + result.stderr)
        self.assertEqual(result.stderr, "")
        self.assertIn("VALIDATION PASS" if valid else "exactly one action field with value null", result.stdout)

    def test_canonical_template(self):
        self.check_template(self.template, True)

    def test_equivalent_markdown(self):
        for field in ("- action: null", "- action: `null`", "- `action`: null",
                      "* `action` :  `null`  ", "+ action:\t null"):
            with self.subTest(field=field):
                self.check_template(self.template.replace(FIELD, field), True)

    def test_non_null_despite_null_checklist(self):
        for value in ('{}', '{"type": "publish"}', '"null"', '`"null"`',
                      'false', '0', '[]', 'NULL', 'nullptr', '', '`null` or publish'):
            with self.subTest(value=value):
                self.check_template(self.template.replace(FIELD, "- `action`: " + value), False)

    def test_missing_field_despite_null_checklist(self):
        self.check_template(self.template.replace(FIELD, ""), False)

    def test_duplicate_fields(self):
        for extra in (FIELD, "- action: {}"):
            with self.subTest(extra=extra):
                self.check_template(self.template.replace(FIELD, FIELD + "\n" + extra), False)

    def test_null_in_other_section_does_not_count(self):
        self.check_template(self.template.replace(FIELD, "") + "\n## Other\n" + FIELD, False)

    def test_missing_packet_section(self):
        self.check_template(self.template.replace("## Human DecisionPacket", "## Other packet"), False)

    def test_duplicate_packet_sections(self):
        self.check_template(self.template + "\n## Human DecisionPacket\n" + FIELD, False)

    def test_prose_is_not_a_field(self):
        self.check_template(self.template.replace(FIELD, "The action: `null` is required."), False)

    def test_schema_null_invariant_still_enforced(self):
        schema = self.root / "contracts/decision-packet.schema.json"
        data = json.loads(schema.read_text(encoding="utf-8"))
        data["properties"]["action"]["type"] = "object"
        schema.write_text(json.dumps(data), encoding="utf-8")
        result = subprocess.run(
            [sys.executable, str(self.root / "scripts/validate_artifact_spine.py")],
            capture_output=True, text=True, check=False,
        )
        self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
        self.assertIn("DecisionPacket.action must remain null", result.stdout)


if __name__ == "__main__":
    unittest.main()
