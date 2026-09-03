# ADR-001 — Mission-first authority model

**Status:** Accepted  
**Date:** 2026-09-03

## Decision

Learner progression dùng một canonical Mission spine:

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

`CURRICULUM.md` là authority duy nhất cho sequence, evidence ladder, autonomy ceiling và PASS. Không duy trì song song numeric lesson roadmap hoặc migration mapping như một authority thứ hai.

## Consequences

- lesson dùng Mission ID;
- Mission planned chưa có lesson/starter/eval thì không tạo file giả-complete;
- technology profile không được đổi learner sequence;
- historical curriculum tồn tại ở repo cũ, không migrate vào repo v2;
- conflict được sửa tại source thấp hơn, không thêm mapping layer.
