# M00 Evidence Workspace (không gian lưu bằng chứng M00)

Thư mục này là nơi lưu **evidence (bằng chứng) thực tế của learner** cho Mission M00.

## File cần dùng

- Template chuẩn dùng lại: `templates/M00-EVIDENCE-PACKET.md`
- File làm việc thực tế của learner: `evidence/M00/M00-EVIDENCE-PACKET.md`

Từ M00.2 trở đi, mọi public observation (quan sát công khai) thật dùng để claim E1 phải được ghi vào `evidence/M00/M00-EVIDENCE-PACKET.md`.

## Quy tắc

1. Không sửa `templates/M00-EVIDENCE-PACKET.md` để lưu dữ liệu cá nhân của một lần học; template chỉ là mẫu chuẩn.
2. Mỗi lần quan sát thật có `observation_id` riêng.
3. Cùng một offer/sản phẩm dùng cùng `subject_id`, nhưng quan sát ở thời điểm khác phải có `observation_id` khác.
4. E1 cần URL công khai thật, `observed_at`, `access_method`, claim/value, `claim_kind`, trạng thái quan sát và `limitation`.
5. `fact` chỉ ghi điều nguồn hỗ trợ trực tiếp. Suy luận như “bán nhiều nên chắc chắn Affiliate tốt” phải là `assumption` hoặc được giữ ngoài supported facts.
6. `0 != missing != pending != not_yet_observable != inconclusive`.
7. Không commit raw credential (thông tin xác thực thô), account data (dữ liệu tài khoản), personal data (dữ liệu cá nhân) hoặc customer data (dữ liệu khách hàng).
8. M00 chỉ read-only (chỉ đọc): không publish (đăng), send (gửi), spend (chi tiền), mua hàng hay thay đổi tài khoản để hoàn thành evidence.

## Workflow (quy trình) đề xuất

```text
Mở source công khai bằng trình duyệt
→ ghi đúng điều đang thấy
→ cấp observation_id
→ ghi observed_at + access_method
→ phân loại fact/estimate/assumption/unknown
→ ghi limitation
→ kiểm tra lại URL
→ commit vào evidence/M00/M00-EVIDENCE-PACKET.md
```

Khi có ít nhất 3 E1 public observations thật, tiếp tục M00.3 để lập Human DecisionPacket (gói quyết định do người lập) và bind (liên kết) chính xác `evidence_ids`.
