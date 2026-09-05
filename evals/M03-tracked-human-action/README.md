# Eval pack — M03

`cases.json` là executable contract data cho `lab/mission-runtime`.

```bash
cd lab/mission-runtime
go test ./...
```

Eval PASS không tự tạo Reality/Operated PASS.

BR-02 kiểm `NO_OBSERVED_OUTCOME` trước window end phải trả `MEASUREMENT_WINDOW_OPEN`, đúng bằng/sau end vẫn nhận số 0 thực đo. `PENDING` và trạng thái giao dịch đã quan sát được phép ghi trước end nhưng không trước action. `mode=outcome` chỉ kiểm record riêng; `mode=link` mới kiểm điều kiện thời gian với ActionRecord.

Chạy riêng hồi quy biên thời gian (bao gồm độ chính xác nano giây và timezone): `go test ./cmd/demo -run 'TestM03(MeasurementWindow|EvalPack)' -v` từ `lab/mission-runtime`.
