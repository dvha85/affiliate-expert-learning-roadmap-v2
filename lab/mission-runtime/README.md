# Mission Runtime — O00 và M03–M07

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
```

Runtime này **không giả vờ thay thế Reality PASS**. M03 cần action thật do người thực hiện; M04/M05 cần evidence/evaluation thật; M06/M07 cần learner vận hành workflow/read-only Agent với nguồn được phép.
