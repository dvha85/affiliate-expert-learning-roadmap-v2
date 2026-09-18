# Local regression after commit `3eb1a96`

## Scope

Commit `3eb1a9670a7e5ff005427d1af7d49c549953d926` fixes POSIX stable append
on macOS by resolving only the system `/var` alias before descriptor traversal.
Caller-controlled symlink ancestors remain rejected. The same commit makes the
process-termination test harness compile on Windows by using
`os.Process.Kill`; it does not claim native Windows lock/path parity.

## Verification

- GitHub Curriculum CI and Mission Agent Path CI: 12/12 checks PASS.
- macOS default `TMPDIR` parent-swap regression: PASS.
- Learner Bot `go test -race ./cmd/bot` and `go vet ./...`: PASS.
- Contracts, core and mission-runtime test/vet suites: PASS.
- 133 Python tests, readiness audit, BR-16a and BR-18b: PASS.
- `GOOS=windows GOARCH=amd64 go test -c ./cmd/bot`: PASS.

## Boundary

This is bounded local pathname/test portability evidence. It does not prove
Windows native locking, crash/power-loss durability, atomic multi-file
transactions, distributed locking, provider operation, live execution,
business outcome, clean-machine pilot or target-host deployment. Readiness
remains `NOT_READY_FOR_PRODUCTION`.
