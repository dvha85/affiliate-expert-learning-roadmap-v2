# M10 Operated Evidence — mẫu E5 governed canary

## 1. Bối cảnh thật

- Tham chiếu Decision/Evidence:
- ID/hash của ActionIntent:
- Target/account (đích/tài khoản):
- Action type + risk class (loại hành động + lớp rủi ro):
- Vì sao impact thấp/reversible (tác động thấp/có thể đảo ngược):

## 2. CanaryGrant do người phê duyệt + ràng buộc approval

- ID/version/hash của grant:
- ID của `CanaryGrantApproval` / `approval_ref`:
- `CanaryGrantApproval.grant_hash` có bằng exact grant hash không?:
- Người phê duyệt + thời điểm:
- Policy version (phiên bản chính sách):
- Validity window (khoảng hiệu lực):
- Risk/action/host/executor được phép:
- Số execution tối đa tổng / mỗi window:
- Chi phí tối đa theo đơn vị nhỏ + currency (tiền tệ):
- Số outcome tối đa còn chờ:

## 3. Trusted CanaryCostBound (giới hạn chi phí đáng tin)

- ID/hash của CostBound:
- Intent ID/hash được bind:
- Chi phí trần theo đơn vị nhỏ + currency (tiền tệ):
- Tham chiếu nguồn + `observed_at` + `expires_at`:
- Bằng chứng source nằm trong deterministic/control-plane registry:
- Có chứng minh estimate do Agent/ActionIntent tự khai báo không thể thay cost bound không?:

## 4. Preflight (kiểm trước)

- Bằng chứng test kill switch (công tắc dừng):
- Bằng chứng atomic reservation/idempotency (giữ chỗ nguyên tử/chống lặp):
- `CanaryLedger.grant_hash` có đúng exact grant không?:
- Test mất durable ledger sau non-zero state có fail closed không?:
- Review credential/compliance (thông tin xác thực/tuân thủ):
- Quy trình recovery/reconciliation (phục hồi/đối soát):

## 5. Canary execution thật

- ID + decision/reason của CanaryGateDecision:
- Grant hash + cost-bound hash trong gate:
- ID/mode của ExecutionAuthorization:
- Grant hash + cost-bound hash trong authorization:
- ID/status/side-effect state của ExecutionRecord:
- External reference (tham chiếu bên ngoài):
- Ledger trước/sau:

## 6. Outcome thật

- ID/source/observed_at của OutcomeRecord:
- Execution ID mà OutcomeRecord resolve tới:
- Outcome đã resolve đúng pending execution chưa?:
- Pending outcome (kết quả còn chờ) trước/sau:

## 7. Failure evidence (bằng chứng ca lỗi)

- Test `approval_ref` đúng nhưng `grant_hash` sai:
- Test trusted cost-bound (giới hạn chi phí đáng tin) bị thiếu/sửa/hết hạn:
- Test Agent cost hint (gợi ý chi phí) thấp hơn trusted cost bound:
- Test rate/budget (tần suất/ngân sách):
- Test mất/reset durable ledger (sổ bền vững):
- Test outcome liên kết sai execution:
- Test duplicate/idempotency (lặp/chống lặp):
- Test kill/revoke (dừng/thu hồi):
- Test unknown-effect/reconciliation (tác động chưa xác định/đối soát):
- Bằng chứng RISK2 không auto (không tự động):

## 8. Review sau canary

- KEEP / NARROW / REVOKE / NEW_GRANT_VERSION:
- Lý do:
- Measurement tiếp theo (phép đo tiếp theo):
- Có tăng exposure (mức phơi nhiễm) không? Nếu có, evidence nào biện minh?:
