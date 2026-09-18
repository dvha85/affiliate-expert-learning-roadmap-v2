# Local and CI regression after commit `598bb21`

## Scope

Commit `598bb21801d44d9f2a1bf5ed56fbcb522aa2c0e9` replaces the Windows
cooperative directory lock with an exclusive native file handle that the
kernel releases when the owning process exits. Windows stable append uses a
native handle with `FILE_FLAG_OPEN_REPARSE_POINT` and bounded non-symlink
ancestor preflight. This is a local single-host hardening boundary.

## Verification

- GitHub Curriculum CI and Mission Agent Path CI: 12/12 checks PASS.
- Learner Bot race/vet and full local learner tests: PASS.
- Readiness audit and 133 Python tests: PASS.
- `GOOS=windows GOARCH=amd64 go test -c ./cmd/bot`: PASS.
- POSIX/macOS stable append and parent-swap regressions: PASS.

## Boundary

No Windows runtime host was available in this workspace, so native Windows
execution and ancestor-race behavior remain unverified. This evidence does not
prove distributed locking, crash/power-loss durability, atomic multi-file
transactions, provider operation, live execution, business outcome,
clean-machine pilot or target-host deployment. Readiness remains
`NOT_READY_FOR_PRODUCTION`.
