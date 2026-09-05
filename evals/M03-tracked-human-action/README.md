# Eval pack — M03

`cases.json` là executable contract data cho `lab/mission-runtime`.

```bash
cd lab/mission-runtime
go test ./...
```

Eval PASS không tự tạo Reality/Operated PASS.

BR-03c.1 thêm `TestM03RawEvalPack` dùng JSON gốc, kiểm schema trước semantic. E02 giữ expected typed `REJECT_MACHINE_EXECUTION` và thêm `raw_expected:INVALID_SCHEMA` vì performed_by=agent vi phạm const human. Các expected còn lại giữ nguyên. Test raw required/null/duplicate/unknown và CLI nằm trong `m03_boundary_test.go`; [boundary M03](../../docs/architecture/M03-JSON-BOUNDARY.md) mô tả phạm vi. Không bỏ test typed cũ.

BR-02 kiểm `NO_OBSERVED_OUTCOME` trước window end phải trả `MEASUREMENT_WINDOW_OPEN`, đúng bằng/sau end vẫn nhận số 0 thực đo. `PENDING` và trạng thái giao dịch đã quan sát được phép ghi trước end nhưng không trước action. `mode=outcome` chỉ kiểm record riêng; `mode=link` mới kiểm điều kiện thời gian với ActionRecord.

Chạy riêng hồi quy biên thời gian (bao gồm độ chính xác nano giây và timezone): `go test ./cmd/demo -run 'TestM03(MeasurementWindow|EvalPack)' -v` từ `lab/mission-runtime`.
