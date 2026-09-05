# Eval pack — M09 Approval-gated Execution

BR-03c.6 thêm [audit JSON](../../docs/architecture/M09-JSON-BOUNDARY.md) và tests raw schema, linked files, serialized runtime artifacts. Typed eval/durable replay/kill-switch cũ giữ nguyên. Chạy lại audit cùng file không thay test one-time execution hoặc trusted approval.

Eval offline + local sandbox kiểm các semantics (ngữ nghĩa): human approval binding, expiry, current policy recheck, executor allowlist, kill switch, idempotency, durable resume và audit linkage.

Fixture PASS chỉ chứng minh capability/boundary. Nó không thay Reality/Operated PASS và không tự tạo E5 governed canary.
