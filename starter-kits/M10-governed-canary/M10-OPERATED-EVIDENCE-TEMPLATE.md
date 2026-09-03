# M10 Operated Evidence — mẫu E5 governed canary

## 1. Bối cảnh thật

- Tham chiếu Decision/Evidence:
- ID/hash của ActionIntent:
- Target/account (đích/tài khoản):
- Action type + risk class (loại hành động + lớp rủi ro):
- Vì sao impact thấp/reversible (tác động thấp/có thể đảo ngược):

## 2. Human CanaryGrant

- ID/version/hash của grant:
- `approval_ref` đáng tin + người phê duyệt:
- Policy version (phiên bản chính sách):
- Validity window (khoảng hiệu lực):
- Risk/action/host/executor được phép:
- Số execution tối đa tổng / mỗi window:
- Chi phí tối đa theo đơn vị nhỏ + currency (tiền tệ):
- Số outcome tối đa còn chờ:

## 3. Preflight (kiểm trước)

- Bằng chứng test kill switch (công tắc dừng):
- Bằng chứng atomic reservation/idempotency (giữ chỗ nguyên tử/chống lặp):
- Review credential/compliance (thông tin xác thực/tuân thủ):
- Quy trình recovery/reconciliation (phục hồi/đối soát):

## 4. Canary execution thật

- ID + decision/reason của CanaryGateDecision:
- ID/mode của ExecutionAuthorization:
- ID/status/side-effect state của ExecutionRecord:
- External reference (tham chiếu bên ngoài):
- Ledger trước/sau:

## 5. Outcome thật

- ID/source/observed_at của OutcomeRecord:
- Outcome đã resolve đúng execution chưa?:
- Pending outcome (kết quả còn chờ) trước/sau:

## 6. Failure evidence (bằng chứng ca lỗi)

- Test rate/budget (tần suất/ngân sách):
- Test duplicate/idempotency (lặp/chống lặp):
- Test kill/revoke (dừng/thu hồi):
- Test unknown-effect/reconciliation (tác động chưa xác định/đối soát):
- Bằng chứng RISK2 không auto (không tự động):

## 7. Review sau canary

- KEEP / NARROW / REVOKE / NEW_GRANT_VERSION:
- Lý do:
- Measurement tiếp theo (phép đo tiếp theo):
- Có tăng exposure (mức phơi nhiễm) không? Nếu có, evidence nào biện minh?:
