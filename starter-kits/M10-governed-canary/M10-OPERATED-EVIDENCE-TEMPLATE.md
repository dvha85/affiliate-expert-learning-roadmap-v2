# M10 Operated Evidence — mẫu E5 governed canary

## 1. Bối cảnh thật

- Decision/Evidence refs:
- ActionIntent ID/hash:
- Target/account:
- Action type + risk class:
- Vì sao impact thấp/reversible (có thể đảo ngược):

## 2. Human CanaryGrant

- Grant ID/version/hash:
- Trusted approval_ref + approver:
- Policy version:
- Validity window:
- Allowed risk/action/host/executor:
- Max executions total / per window:
- Max cost minor + currency:
- Max pending outcomes:

## 3. Preflight (kiểm trước)

- Kill switch test evidence:
- Atomic reservation/idempotency proof:
- Credential/compliance review:
- Recovery/reconciliation procedure:

## 4. Canary execution thật

- CanaryGateDecision ID + decision/reason:
- ExecutionAuthorization ID/mode:
- ExecutionRecord ID/status/side-effect state:
- External reference:
- Ledger before/after:

## 5. Outcome thật

- OutcomeRecord ID/source/observed_at:
- Outcome đã resolve đúng execution chưa?:
- Pending outcome trước/sau:

## 6. Failure evidence

- Rate/budget test:
- Duplicate/idempotency test:
- Kill/revoke test:
- Unknown-effect/reconciliation test:
- Bằng chứng RISK2 không auto:

## 7. Review sau canary

- KEEP / NARROW / REVOKE / NEW_GRANT_VERSION:
- Lý do:
- Measurement tiếp theo:
- Có tăng exposure không? Nếu có, evidence nào biện minh?:
