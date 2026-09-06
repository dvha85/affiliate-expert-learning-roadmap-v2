# BR-09 — Quyết định adapter nhập M00 và nghiệm thu lab

Baseline: BR-08 đã review/merge #51 (`9fc6147`), closure scope trên main `a3bccc7`. BR-09 chưa DONE trước review/merge PR này.

PR triển khai: [#52](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/52), READY_FOR_REVIEW, chưa merge.

## Thiết kế

Chọn JSON transcription profile `m00-input/v1`, không parser Markdown. Core/m00 là converter thuần, learner internal/app chỉ file/stdout/stderr. History M02 giữ ownership, format và formula; chỉ bổ sung kiểm projection và ID nguồn của observation do importer tạo. Không đưa application/store vào core.

Chọn giữ provenance có version trong `transformation_or_method` hiện có thay vì thêm field JSONL: typed history không bỏ mất payload, input_hash bao gồm payload, không sửa lại record cũ hoặc reseal. Đánh đổi: provenance phải decode JSON bên trong string; đây là profile nội bộ, không coi mọi transformation string legacy là JSON. Marker `access_method=local_packet_conversion` kích hoạt kiểm tính lại projection. Bỏ/giả mạo toàn bộ marker và provenance không bị coi là xác thực nguồn; không có chữ ký hoặc trust store.

V1 chỉ price/commission_rate, hai field bắt buộc khai báo kể cả missing. Context product_name/currency được người dùng cung cấp rõ ràng, chưa là observation xác minh. Giá trị không biểu diễn giữ nguyên nghĩa thập phân qua float64 M02 bị từ chối, không âm thầm làm tròn hoặc underflow thành 0. Mỗi field giữ ID/source/time/state/claim/value/role/method/limitation; aggregate assumption không nâng source claim thành fact. Bảng map/lệnh đầy đủ: [bài thực hành](../../examples/m00-import/README.md).

## Ma trận review

| Điều kiện | Bằng chứng |
|---|---|
| Input strict, không đoán field | core/m00 tests: duplicate keys/IDs/field, wrong subject, missing slot, unknown field/case, invalid time/range/null, mixed origin, precision |
| 0 khác unknown/pending/missing | Tests mọi state và unknown; giữ state/value nguồn, projection null; t1 GET_MORE_DATA |
| Không ghi khi import | app tests input bytes không đổi, envelope VALID/USAGE_ERROR/IO_ERROR/INVALID_INPUT, lỗi không artifact |
| Cùng Bot, cùng history và adapter quyết định | scripts/smoke_br09.py: 2 binary process sequences import→M01→capture→DecisionPacket, exact aggregate IDs/action=null |
| Restart, duplicate/conflict | Smoke 2 MATCH, EXACT_DUPLICATE giữ bytes, đổi source dưới aggregate/record ID mới bị CONFLICT |
| Chống projection lệch provenance | Core tính lại projection khi capture/load; unit test và smoke sửa commission đều lỗi trước append |
| Tương thích | Tests/vet bốn module; smoke BR-08; không đổi schema, JSON tags, formula hoặc hash cũ |

Các kiểm thử local PASS ngày 2026-09-06: tests/vet bốn module, smoke BR-08/BR-09, 8 validators và 10 Python regressions. CI thêm smoke BR-09 ở deterministic-runtime. Bằng chứng này là synthetic lab, không là pilot học viên, E1 hoặc readiness vận hành.

Kiểm lại snapshot `c2266fc` bằng clone local độc lập `git clone --no-local`: cùng toàn bộ checks trên PASS; git status trống trước/sau. Dùng toolchain/cache sẵn có, không là cold install hoặc clone remote trên máy học viên. Sau snapshot chỉ đồng bộ README/evidence.

## Giới hạn còn lại

### Sửa findings review #52

- P1: reader Scanner mặc định 64 KiB không đọc lại dòng 80.103 byte dù capture đã APPENDED. Đã thống nhất `store.MaxHistoryRecordBytes=1<<20`, reader cấp chỗ cho payload + CRLF và kiểm payload; application/store chặn quá giới hạn trước ghi. Không nâng thành unbounded reader hoặc sửa history cũ.
- P2: timestamp cùng instant khác timezone được chọn theo thứ tự đầu vào, trong khi provenance sắp field; tính lại projection bị lệch. Đã thêm tie-break từ điển cho equal instant, giữ nguyên timestamp nguồn.
- Regression đã FAIL trước sửa: timezone self-validation; 40 sản phẩm load lỗi token too long; oversized append được chấp nhận. Sau sửa tests kiểm permutation, source timestamps, 40-product capture/load/replay/duplicate, không ghi khi oversize, giới hạn -1/đúng/+1 với LF/CRLF/no newline đều PASS. Giới hạn 1 MiB là chính sách lab tường minh, không thay schema/hash/formula.

Review/merge BR-09 trước khi đóng checklist. Không fetch account/program, không gom nguồn mâu thuẫn tự động, không migrate store, không trusted clock/đa writer/crash-safe persistence, không mở live execution. BR-10 mới liên kết/persist action/outcome. BR-06b vẫn chờ chương trình/kênh thật.
