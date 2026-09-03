# Starter kit — M02 Trustworthy History + Replay

M02 tiếp tục dùng **một runtime duy nhất** tại `lab/affiliate-bot`; starter kit không duplicate implementation.

## Dùng theo thứ tự

1. Học `curriculum/M02/M02.1` → `M02.4`.
2. Chạy tests/runtime tại `lab/affiliate-bot`.
3. Dùng `CHECKPOINTS.md` để kiểm Mission gate.
4. Copy `M02-OPERATED-EVIDENCE-TEMPLATE.md` vào `learner/M02/` cho evidence cá nhân.
5. Executable eval pack nằm ở `evals/M02-history-replay/`.

## Scope

```text
immutable local history
+ deterministic query
+ versioned replay
+ restart proof
+ NO external action
```

Không thêm database, scheduler, n8n hay Agent chỉ để hoàn thành M02.
