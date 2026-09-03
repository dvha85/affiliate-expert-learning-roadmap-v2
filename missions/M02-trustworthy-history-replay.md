---
mission_id: M02
title: Trustworthy History + Replay v0.2
status: ready
requires_missions: [M01]
minimum_evidence: E1
readiness_target: E3
authority: A0 deterministic
external_side_effects: false
runtime: lab/affiliate-bot/
starter: starter-kits/M02-history-replay/
eval_pack: evals/M02-history-replay/
---

# Mission M02 — Trustworthy History + Replay v0.2 (lịch sử đáng tin + phát lại)

## Vì sao Mission này tồn tại

M01 có thể trả một deterministic decision (quyết định tất định) ở hiện tại. Nhưng nếu Bot không giữ được **input (đầu vào) nào**, **version công thức (phiên bản công thức) nào** và **decision (quyết định) nào** đã tạo ra output (đầu ra) trước đây, ta không thể audit (kiểm toán), replay (phát lại) hay phân biệt market change (thay đổi thị trường) với code drift (độ lệch do mã nguồn).

M02 biến một lần chạy thành lịch sử có thể kiểm chứng:

```text
M01 canonical input (đầu vào chuẩn hóa)
+ explicit as_of (thời điểm đánh giá được ghi rõ)
+ explicit ingested_at (thời điểm Bot nhận dữ liệu được ghi rõ)
+ formula_version (phiên bản công thức)
→ immutable HistoryRecord (bản ghi lịch sử bất biến)
→ append-only JSONL (JSONL chỉ nối thêm)
→ deterministic query (truy vấn tất định)
→ replay bằng đúng formula version (phát lại bằng đúng phiên bản công thức)
→ MATCH | DRIFT | UNREPLAYABLE
→ KHÔNG có hành động bên ngoài
```

## Evidence semantics (ngữ nghĩa bằng chứng)

`minimum_evidence: E1` là mức bằng chứng tối thiểu để M02 nối với reality của M00/M01. `readiness_target: E3` chỉ nói history model phải **sẵn sàng tiếp nhận outcome thật ở Mission sau**; nó không có nghĩa M02 đã có hoặc được claim E3.

```text
readiness_target: E3
!= E3 evidence achieved
!= M03 PASS
```

## Contract (hợp đồng) cốt lõi

### Identity (định danh)

```text
product_id     = stable subject identity (định danh ổn định của đối tượng) trong learner runtime
observation_id = một observation (lần quan sát) cụ thể của đối tượng đó
record_id      = một decision/history snapshot (ảnh chụp quyết định/lịch sử) cụ thể
```

Không dùng product name (tên sản phẩm) làm stable identity và không tái sử dụng một `observation_id` cho nội dung khác. Cùng `observation_id` nhưng canonical content (nội dung chuẩn hóa) khác là conflict (xung đột), không phải một observation mới.

### Time (thời gian)

```text
observed_at = khi source/world (nguồn/thế giới) được quan sát
ingested_at = khi Bot nhận/lưu record (bản ghi)
as_of       = thời điểm decision context (ngữ cảnh quyết định) được đánh giá
```

Ba timestamp (mốc thời gian) này có vai trò khác nhau. Arrival order (thứ tự dữ liệu đến) không được giả làm world-time order (thứ tự thời gian thực tế). Một decision record cũng không được có `as_of` sớm hơn observation mà nó tuyên bố đã sử dụng.

### Append-only history (lịch sử chỉ nối thêm)

History (lịch sử) chuẩn của learner là JSONL append-only. Record cũ không được silently overwrite (âm thầm ghi đè).

```text
same record_id + same canonical content
(cùng record_id + cùng nội dung chuẩn hóa)
→ EXACT_DUPLICATE
→ idempotent (lặp lại an toàn), không append lần hai

same record_id + different canonical content
(cùng record_id + nội dung chuẩn hóa khác)
→ CONFLICT
→ reject / human review (từ chối / người kiểm tra)
```

Valid late/out-of-order record (bản ghi hợp lệ đến muộn/không đúng thứ tự) vẫn phải được giữ lại; query/report (truy vấn/báo cáo) sort theo `as_of` đã parse (phân tích), không theo thứ tự dòng trong file hay `ingested_at`.

### Replay (phát lại)

Replay không có nghĩa “chạy code hiện tại lên dữ liệu cũ” một cách mơ hồ. Mỗi record phải lưu:

```text
formula_version: phiên bản công thức
input_hash: hash của canonical input snapshot
recorded_result: decision đã lưu
```

Replay dùng đúng implementation tương ứng với `formula_version` nếu runtime còn hỗ trợ:

```text
same version + same canonical input integrity + same result
→ MATCH

same version + same input nhưng result khác
→ DRIFT

version không còn implementation tương ứng
→ UNREPLAYABLE
```

`UNREPLAYABLE` trung thực hơn việc lén dùng công thức mới để giả vờ tái tạo decision cũ.

### Integrity (toàn vẹn) trước replay

`input_hash` phải được kiểm trước khi gọi record là replayable. Nếu observation bị sửa sau khi capture, record là integrity failure (lỗi toàn vẹn), không phải một replay bình thường.

Hash canonical phải ổn định trước ordering (thứ tự) không có ý nghĩa. Cùng tập input semantic nhưng order khác không được tạo hash khác chỉ vì serializer (bộ tuần tự hóa) nhận mảng theo thứ tự khác.

## Query/restart

M02 phải chứng minh data không chỉ tồn tại trong RAM:

```text
capture
→ stop process
→ start process mới
→ load append-only history
→ query theo as_of
→ replay
```

Nếu restart làm mất history, M02 chưa đạt Operated.

## Các failure case bắt buộc

- duplicate `record_id` cùng nội dung → idempotent, không append lại;
- duplicate `record_id` khác nội dung → conflict;
- cùng `observation_id` bị tái sử dụng với nội dung khác → conflict;
- record đến muộn/out-of-order vẫn query đúng theo `as_of`;
- `as_of < observed_at` → reject;
- timestamp invalid → reject;
- `input_hash` mismatch → integrity failure;
- corrupt JSONL → fail closed;
- unsupported `formula_version` → `UNREPLAYABLE`, không silent fallback;
- cùng canonical input nhưng observation order khác → cùng hash;
- replay result khác recorded result cùng version → `DRIFT`.

## PASS

### Capability (năng lực)

- executable eval M02 PASS;
- Go tests/vet PASS;
- capture/list/replay hoạt động qua restart;
- history fail closed với tamper/corruption/version unknown.

### Reality (thực tế)

- nối ít nhất một HistoryRecord với E1 context thật từ Mission trước;
- không gọi fixture synthetic là market history thật.

### Operated (đã tự vận hành chứng minh)

- learner tự capture, restart, query, replay;
- tự gây một failure case và giải thích được vì sao fail;
- lưu operated evidence + limitation.

## Corrective authoring gate (cổng hoàn thiện nội dung)

Mission M02 chỉ được gọi `ready` khi lesson, starter, operated template, executable eval và runtime tests cùng tồn tại. `ready` là trạng thái authoring, không phải learner PASS.
