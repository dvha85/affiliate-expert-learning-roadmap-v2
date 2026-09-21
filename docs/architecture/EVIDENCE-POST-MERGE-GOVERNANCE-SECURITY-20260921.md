# Post-merge governance and security evidence — 2026-09-21

Merged-main baseline: `8f1c8aaf45247ea62c145e879a86d949a300e44b` (`main` and
`origin/main`). The product snapshot that this record extends is
`c155effbf6510d1aaa6587cfd8a85bfd009919e6`; the readiness matrix and evidence
graph now point at the merged-main commit. This record is bounded
repository/settings evidence and does not change the overall readiness
decision.

## Governance

GitHub ruleset **Protect main with required CI gates** (`23753234`) is active
for `main`. The observed rules require a pull request, an up-to-date branch,
`mission-gate`, `curriculum-gate`, and `codeql-go`; force-push and branch
deletion are blocked. No bypass actors or required approval count are configured
in this ruleset. Required-check behavior is still fail-closed in the checked-in
aggregate gate workflows.

This closes the repository-settings portion of the former governance blocker.
It does not substitute for independent human review, provider evidence, a
target-host drill, a learner pilot, or live-execution authority.

## Vulnerability scan

The security workflow now pins `golang.org/x/vuln/cmd/govulncheck@v1.8.0` and no
longer marks the scan `continue-on-error`. The previous `v1.1.4` scanner failed
on the Go 1.27 AST with `unexpected expr: *ast.KeyValueExpr`.

Local verification on Go `1.27.0` ran the new scanner against all four modules:

```text
contracts          PASS — no vulnerabilities affecting code
core               PASS — no vulnerabilities affecting code
lab/affiliate-bot  PASS — no vulnerabilities affecting code
lab/mission-runtime PASS — no vulnerabilities affecting code
```

The scanner also reported one vulnerability in a required module that is not
reachable by the scanned code; the command exited successfully. The hosted
post-merge Security Scans run **#70** on `8f1c8aaf45247ea62c145e879a86d949a300e44b`
passed both `codeql-go` and the blocking `govulncheck-go-modules` job. The
post-merge Curriculum CI **#1276** and Mission Agent Path CI **#1233** runs
also passed on the same merge commit.

## Remaining readiness boundary

The repository remains `NOT_READY_FOR_PRODUCTION`. Selected-source/provider
operation, target-host deployment and restore, clean-machine pilot, authorized
live executor/business outcomes, and distributed/power-loss durability still
need independent evidence and are not manufactured by this record.
