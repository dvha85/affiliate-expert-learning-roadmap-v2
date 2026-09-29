# M06 Operated Evidence (bằng chứng vận hành)

- Kiểm offline m06-check (ghi riêng khỏi operated proof): input ref, output/schema và rejection; không dùng synthetic làm nguồn thật:

- subject_id + source_url (mã đối tượng + URL nguồn):
- access method / allowlist review (cách truy cập / rà soát danh sách cho phép):
- Run 1 correlation_id + change_state (lần chạy 1: mã liên kết + trạng thái thay đổi):
- Run 2 correlation_id + change_state (lần chạy 2: mã liên kết + trạng thái thay đổi):
- Trạng thái persistence + record_id (`APPENDED` / `EXACT_DUPLICATE` / reject):
- Change detection context và persistence context đã tách chưa:
- History path/record count/SHA-256 trước và sau retry hoặc rejection:
- canonical_history_ack, canonical_history_persisted, canonical_history_handoff và execution_permitted quan sát được:
- Restart dùng lại history nào, execution ID mới, record_id và replay output MATCH:
- Missing/wrong Bearer, adapter down, source/profile sai: HTTP/node error, không có ACK/report giả, history hash không đổi:
- Change detection chỉ điền nếu đã chạy normalizer riêng: previous_hash, content fingerprint và expected/observed; nếu chưa chạy ghi NOT_RUN:
- Canonical Observation sample (mẫu quan sát chuẩn):
- history proof (bằng chứng lịch sử):
- retry/failure case (ca retry/lỗi):
- provenance + limitation (nguồn gốc + giới hạn):
- observed result (kết quả quan sát):
- next measurement (phép đo tiếp theo):
