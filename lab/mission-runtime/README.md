# Mission Runtime — conformance harness cho O00 và M03–M11

Runtime offline để chứng minh boundary/semantics của các Mission mà không cần API key, n8n instance hay dịch vụ bên ngoài.

```bash
go test ./...
go vet ./...
go run ./cmd/demo O00
go run ./cmd/demo M03
go run ./cmd/demo M04
go run ./cmd/demo M05
go run ./cmd/demo M06
go run ./cmd/demo M07
go run ./cmd/demo M08
go run ./cmd/demo M09
go run ./cmd/demo M10
go run ./cmd/demo M11
```

Runtime này là **conformance oracle/harness (bộ đối chiếu chuẩn)**, không phải Affiliate Bot thứ hai. Từ M03 trở đi learner dùng behavior/failure cases ở đây để kiểm implementation được gắn vào cùng learner Bot/workspace đang tiến hóa từ `lab/affiliate-bot/`.

```text
run mission-runtime demo/test
!= learner integration
!= Reality PASS
!= Operated PASS
```

M03 cần action thật do người thực hiện; M04/M05 cần evidence/evaluation thật; M06/M07 cần learner vận hành workflow/read-only Agent với nguồn được phép; M08–M11 cần đúng authority/evidence gate của Mission. Continuity gate: `docs/architecture/LEARNER-BOT-CONTINUITY.md`.
