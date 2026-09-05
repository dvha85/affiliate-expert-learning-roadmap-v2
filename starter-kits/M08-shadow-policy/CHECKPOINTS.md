# M08 Checkpoints

- [ ] m08-check từ file đã thử hash tamper, authority=true và missing links; không dùng seal để che lỗi input.

- [ ] ActionIntent liên kết tới Decision/Evidence có thật.
- [ ] Nếu intent bắt nguồn từ Agent thì có `proposal_ref`, và tham chiếu đó truy được về proposal đã biết; proposal của Agent không tự cấp authority.
- [ ] ActionIntent có `intent_mode=PROPOSAL_ONLY` và `execution_authorized=false`.
- [ ] Intent hash đổi khi candidate action bị sửa; lỗi tuần tự hóa JSON không được tạo hash giả.
- [ ] Thời hạn (expiry), target allowlist và policy version đều được kiểm.
- [ ] Loại hành động (action type) không có policy mapping phải `DENY` theo nguyên tắc fail closed.
- [ ] Nhóm rủi ro (risk class) của action type đã biết giữ ổn định kể cả khi provenance/linkage còn thiếu.
- [ ] Ngữ nghĩa idempotency cho duplicate/collision đã được thử.
- [ ] PolicyDecision luôn `execution_authorized=false`, kể cả khi kết quả là `ALLOW`.
- [ ] Policy unavailable/invalid phải fail closed.
- [ ] Không có write tool/executor trong operated path.
- [ ] Reality + Operated evidence đã lưu.
