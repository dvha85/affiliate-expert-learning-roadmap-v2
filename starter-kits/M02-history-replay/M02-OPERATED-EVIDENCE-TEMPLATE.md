# M02 Operated Evidence — learner copy

> Copy vào `learner/M02/M02-OPERATED-EVIDENCE.md`. Không commit raw personal/account evidence.

## Identity / time

```text
product_id:
observation_id t1:
observed_at t1:
ingested_at t1:
as_of t1:
observation_id t2:
observed_at t2:
ingested_at t2:
as_of t2:
```

## Append / restart

```text
History path:
Append commands/results:
History count before restart:
History count after restart:
Evidence that old record was not overwritten:
Evidence snapshot did not change after source object/file handling:
```

## Failure cases

```text
Exact duplicate result:
Same record_id conflict result:
Same observation_id + different content result:
Out-of-order result:
Corrupt-line result:
Hash-tamper result:
as_of-before-observed_at result:
```

## Replay

```text
Formula version:
Input hash verified? yes/no
Integrity gate: PASS | ERROR
Replay state if integrity PASS: MATCH | DRIFT | UNREPLAYABLE
Rerun same result? yes/no
```

## Explain-back

```text
What replay MATCH proves:
What replay MATCH does NOT prove:
Why formula_version must be preserved:
Why observed_at != ingested_at != as_of:
Why integrity failure is not DRIFT:
Why input_hash is integrity evidence, not market truth:
Why history/replay gives no execution permission:
Next measurement:
```
