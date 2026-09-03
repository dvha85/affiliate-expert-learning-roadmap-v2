# M02 Operated Evidence (bằng chứng đã vận hành) — bản sao cho người học

> Copy vào `learner/M02/M02-OPERATED-EVIDENCE.md`. Không commit raw personal/account evidence (bằng chứng thô cá nhân/tài khoản).

## Identity / time (định danh / thời gian)

```text
product_id:
observation_id t1:
observed_at t1:
ingested_at t1:
as_of t1:
observation_id t2:
observed_at t2:
ingested_at t2:
as_of t2:
```

## Append / restart (nối thêm / khởi động lại)

```text
History path: đường dẫn lịch sử
Append commands/results: lệnh/kết quả nối thêm
History count before restart: số bản ghi trước khi khởi động lại
History count after restart: số bản ghi sau khi khởi động lại
Evidence that old record was not overwritten: bằng chứng bản ghi cũ không bị ghi đè
Evidence snapshot did not change after source object/file handling: bằng chứng snapshot không đổi sau khi xử lý object/file nguồn
```

## Failure cases (các ca lỗi)

```text
Exact duplicate result: kết quả bản trùng chính xác
Same record_id conflict result: kết quả xung đột cùng record_id
Same observation_id + different content result: kết quả cùng observation_id nhưng nội dung khác
Out-of-order result: kết quả dữ liệu đến sai thứ tự
Corrupt-line result: kết quả dòng dữ liệu hỏng
Hash-tamper result: kết quả khi hash bị sửa
as_of-before-observed_at result: kết quả khi as_of sớm hơn observed_at
```

## Replay (phát lại)

```text
Formula version: phiên bản công thức
Input hash verified? yes/no: hash đầu vào đã xác minh? có/không
Integrity gate: PASS | ERROR: cổng toàn vẹn
Replay state if integrity PASS: MATCH | DRIFT | UNREPLAYABLE: trạng thái phát lại nếu toàn vẹn PASS
Rerun same result? yes/no: chạy lại có cùng kết quả? có/không
```

## Explain-back (tự giải thích lại)

```text
What replay MATCH proves: replay MATCH chứng minh gì
What replay MATCH does NOT prove: replay MATCH KHÔNG chứng minh gì
Why formula_version must be preserved: vì sao phải giữ formula_version
Why observed_at != ingested_at != as_of: vì sao ba mốc thời gian khác nhau
Why integrity failure is not DRIFT: vì sao lỗi toàn vẹn không phải DRIFT
Why input_hash is integrity evidence, not market truth: vì sao input_hash là bằng chứng toàn vẹn, không phải sự thật thị trường
Why history/replay gives no execution permission: vì sao history/replay không cấp quyền thực thi
Next measurement: phép đo tiếp theo
```
