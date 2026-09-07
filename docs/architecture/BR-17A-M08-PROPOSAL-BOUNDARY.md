# BR-17a — M08 proposal-only boundary

M08 chỉ tạo `ActionIntent`/`PolicyDecision` dạng đề xuất. Các bất biến bắt buộc:

- `intent_mode=PROPOSAL_ONLY`;
- `policy_mode=NON_AUTHORIZING`;
- `execution_authorized=false` ở cả hai artifact;
- không có executor node/credential hoặc external side effect trong workflow.

Validator semantic hiện hành kiểm tra các bất biến này cùng các case tamper,
expired, missing evidence/proposal và duplicate/idempotency. Đây là lab/schema
evidence; không phải authorization hay live execution evidence.

```text
python scripts/validate_agent_semantics.py
```
