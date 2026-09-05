# Eval pack — M08 Shadow ActionIntent + Deterministic Policy

BR-03c.5 thêm [boundary file](../../docs/architecture/M08-JSON-BOUNDARY.md) và tests input/output schema. Typed eval cũ vẫn seal fixture; CLI mới không seal lại input. Không gọi typed eval là raw conformance.

Eval này chứng minh capability (năng lực) và authority boundary (ranh giới quyền hạn) của M08 bằng dữ liệu offline.

Các nhóm ca kiểm thử:

- `ALLOW` ở RISK0 vẫn chỉ là shadow và `execution_authorized=false`;
- RISK1/RISK2 đi `HUMAN_REVIEW` nhưng không thực thi;
- decision/evidence linkage bị thiếu → `GET_MORE_DATA`;
- intent hết hạn, bị sửa sau khi hash, target ngoài allowlist hoặc live flag → fail closed;
- duplicate cùng idempotency key/hash → `WAIT`;
- idempotency collision → `DENY`;
- policy không khả dụng → `DENY`.

Fixture PASS không thay Reality/Operated PASS. Learner phải chạy shadow policy trên artifact thật từ chuỗi E4 phù hợp.
