"""Compare the execution decoder with flatted from the pinned n8n install."""
import argparse
import json
from pathlib import Path
import subprocess

from run_n8n_m06_schedule_regression import decode_flatted_execution


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--node-modules", type=Path, required=True)
    parser.add_argument("--node", default="node")
    args = parser.parse_args()
    # Generate genuine flatted output, not a guessed table. The duplicated
    # object below is shared but acyclic; cycles have a documented marker in
    # the Python evidence decoder and are tested separately.
    script = """
const { stringify, parse } = require(require.resolve('flatted', {paths: [process.argv[1]]}));
const shared = {digits: '007', nested: ['1', '2', 'text', 5, null, true]};
const cases = [{value: '1'}, {value: '2'}, {a: shared, b: shared}];
console.log(JSON.stringify(cases.map(value => ({encoded: stringify(value), decoded: parse(stringify(value))}))));
"""
    result = subprocess.run([args.node, "-e", script, str(args.node_modules.resolve())], text=True, capture_output=True, check=True)
    cases = json.loads(result.stdout)
    for case in cases:
        actual = decode_flatted_execution(case["encoded"])
        if actual != case["decoded"]:
            raise AssertionError(f"flatted parity mismatch: {actual!r} != {case['decoded']!r}")
    print(f"FLATTED PARITY PASS: {len(cases)} cases from pinned n8n dependency")


if __name__ == "__main__":
    main()
