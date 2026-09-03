# n8n compatibility ledger

Blueprint JSON validates structure in CI, but import/execution compatibility is
an engine property and must be demonstrated against a real n8n instance.

## Declared node baseline

| Blueprint | Node | Declared typeVersion |
|---|---|---|
| M06 | Schedule Trigger / Set / HTTP Request / Code | 1.2 / 3.4 / 4.2 / 2 |
| M07 | AI Agent / OpenAI Chat Model / HTTP Request Tool / Code | 3.1 / 1.2 / 1.2 / 2 |

## Release admission record

`tested_n8n_version` is intentionally **not claimed** until a maintainer runs
the two smoke tests below on the exact engine version and records version,
date, import result and execution IDs in the PR/release evidence. Do not turn
“JSON parses” into “n8n supports this workflow”.

```text
tested_n8n_version: UNVERIFIED
tested_node_versions: declared above; engine support unverified
upgrade_review_cadence: before each n8n engine upgrade and at least quarterly
```

## Required smoke tests

1. Import both blueprints into a clean, local n8n instance; inspect unknown or
   migrated node/type versions before saving.
2. M06: run twice with a permitted public fixture/source. Verify GET-only,
   `NEW → UNCHANGED`, `observation_id`, correlation and canonical-history
   handoff. Change fixture content and verify `CHANGED`.
3. M07: use a least-privilege read-only credential or mock. Verify normal
   evidence output, blocked write/prompt-injection request, and `HUMAN_REVIEW`
   boundary output.
4. On upgrade, repeat import and execution smoke tests. Any behavior change in
   HTTP Request Tool, AI Agent, credential scope, static-data behavior or node
   migration blocks activation until review updates this ledger.

The smoke result is integration evidence, not Reality/Operated evidence by
itself. Production credentials, write scopes and secrets must never be placed
in these blueprints or this repository.
