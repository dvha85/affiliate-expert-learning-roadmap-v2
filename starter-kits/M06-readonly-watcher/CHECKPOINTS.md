# M06 Checkpoints

- [ ] Đã kiểm output normalizer bằng m06-check; phân biệt synthetic/test với nguồn thật, thiếu body/HEAD với nội dung đã quan sát; schema PASS không phải history persistence.

- [ ] Source/host đã allowlist và review ToS/access method.
- [ ] Nguồn bên ngoài chỉ GET/HEAD; POST nội bộ loopback có Bearer chỉ gọi canonical adapter, không write credential nguồn.
- [ ] Workflow report có record_id/ACK/persisted; Observation resolve từ canonical history.
- [ ] Case normalizer riêng có previous_hash/fingerprint chứng minh change detection `NEW/UNCHANGED/CHANGED`.
- [ ] phân biệt persistence `APPENDED/EXACT_DUPLICATE` với change detection.
- [ ] sửa content giữ identity cũ bị reject và history bytes không đổi.
- [ ] correlation_id và limitation/provenance được giữ.
- [ ] canonical history có bản ghi từ run thật.
- [ ] retry/re-run không tạo semantics mơ hồ.
- [ ] Reality + Operated evidence đã lưu.
