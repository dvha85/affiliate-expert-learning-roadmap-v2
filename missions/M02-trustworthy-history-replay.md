---
mission_id: M02
title: Trustworthy History + Replay v0.2
status: ready
requires_missions: [M01]
minimum_evidence: E1 + replayable local history
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
canonical input snapshot: ảnh chụp đầu vào chuẩn hóa
input_hash: mã băm đầu vào
recorded decision: quyết định đã ghi nhận
```

Replay dùng version registry (sổ đăng ký phiên bản) của deterministic core (lõi tất định):

```text
known version + same result → MATCH
(phiên bản đã biết + cùng kết quả → MATCH)

known version + different result → DRIFT
(phiên bản đã biết + kết quả khác → DRIFT)

unknown/retired version → UNREPLAYABLE
(phiên bản không biết/đã ngừng hỗ trợ → UNREPLAYABLE)
```

Nếu `input_hash` không còn khớp hoặc record hỏng, đó là **integrity failure (lỗi toàn vẹn) trước replay**, không phải replay state (trạng thái phát lại) thứ tư. Record phải fail closed (đóng an toàn khi lỗi) và không được báo `MATCH`/`DRIFT` giả.

`UNREPLAYABLE` phải được báo rõ; không được tự động chạy formula mới rồi gọi đó là replay thành công.

## Bằng chứng để repo đạt trạng thái `ready`

M02 chỉ được chuyển từ `planned` sang `ready` sau khi repo có đủ:

- lesson cards (thẻ bài học) M02.1–M02.4;
- `HistoryRecord` machine contract (hợp đồng máy đọc được);
- append/read/query/replay path (luồng nối thêm/đọc/truy vấn/phát lại) trong `lab/affiliate-bot`;
- executable eval pack (bộ ca đánh giá có thể chạy) cho duplicate/conflict/out-of-order/corruption/integrity/replay drift;
- starter/checkpoint (bộ khởi đầu/điểm kiểm tra) + private operated-evidence template (mẫu bằng chứng vận hành riêng tư);
- restart proof (bằng chứng qua khởi động lại);
- CI bảo vệ M01 regression (hồi quy) + M02 replay semantics (ngữ nghĩa phát lại).

Corrective authoring gate (cổng hiệu chỉnh khi soạn nội dung) ngày 2026-09-03 đã chạy khi M02 còn `planned` và đạt cả bốn CI jobs trước khi file này được đổi sang `ready`. `ready` là trạng thái authoring (soạn nội dung) của repo, không phải learner PASS.

## Ranh giới Reality (thực tế)

Minimum reality (mức thực tế tối thiểu) của learner PASS:

```text
ít nhất 1 product_id thật
+ hai E1 observations ở observed_at khác nhau
+ cùng stable identity (định danh ổn định)
+ history vẫn đọc được sau process restart (khởi động lại tiến trình)
```

`UNCHANGED` là outcome (kết quả) hợp lệ. Không cần thị trường phải thay đổi để PASS.

Nếu t2 không quan sát được, ghi missing/access limitation (thiếu/giới hạn truy cập) trung thực; không copy last-known value (giá trị biết gần nhất) và gọi đó là observation mới.

Synthetic fixtures (dữ liệu kiểm thử mô phỏng) chỉ chứng minh failure/replay behavior (hành vi lỗi/phát lại), không thay E1 reality.

## Safety / authority ceiling (an toàn / trần quyền hạn)

M02 vẫn là local deterministic processing (xử lý tất định cục bộ):

- không tự scrape/login (thu thập web/đăng nhập);
- không publish/message/spend (đăng/gửi/chi tiền);
- không n8n/Agent tự thu thập dữ liệu;
- không external side effect (tác động bên ngoài);
- history/replay không tạo Approval (phê duyệt) hoặc Execution permission (quyền thực thi).

```text
replay MATCH != business truth
(MATCH khi phát lại != sự thật kinh doanh)

history exists != permission to act
(có lịch sử != có quyền hành động)

Decision (quyết định) != Approval (phê duyệt) != Execution (thực thi)
```

## PASS

### Capability (năng lực)
- append-only history không overwrite evidence (lịch sử chỉ nối thêm không ghi đè bằng chứng);
- immutable snapshot (ảnh chụp bất biến) không đổi khi caller (bên gọi) sửa object sau capture;
- exact duplicate idempotent (bản sao chính xác lặp lại an toàn), conflict về record/observation identity báo lỗi rõ;
- JSONL corrupt/truncated (hỏng/cụt) và hash tamper (sửa mã băm) phải fail closed;
- out-of-order record được giữ lại và query theo `as_of` đã parse;
- replay phân biệt `MATCH`, `DRIFT`, `UNREPLAYABLE`;
- M01 regression vẫn PASS.

### Reality (thực tế)
- có ít nhất hai E1 observations của cùng stable subject (đối tượng ổn định) ở hai thời điểm quan sát khác nhau, hoặc blocker (điểm chặn) được ghi trung thực.

### Operated (đã tự vận hành chứng minh)
- người học đã append, restart, query, replay và rerun (nối thêm, khởi động lại, truy vấn, phát lại và chạy lại) cùng history/version để chứng minh output deterministic;
- người học explain-back (tự giải thích lại) được `observed_at != ingested_at != as_of`, integrity failure và giới hạn của replay.

## Kết quả

Bot v0.2 có trustworthy local history + deterministic replay (lịch sử cục bộ đáng tin + phát lại tất định). Authority (quyền hạn) vẫn A0, không có external action.
