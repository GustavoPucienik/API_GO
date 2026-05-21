-- ── Fiscal Status ─────────────────────────────────────────────────────────────

-- name: ListFiscalOrderStatuses :many
SELECT fiscal_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_default, is_pending_approval
FROM fiscal_order_status
ORDER BY sort_index;

-- name: GetFiscalOrderStatusByID :one
SELECT fiscal_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_default, is_pending_approval
FROM fiscal_order_status
WHERE fiscal_order_status_id = ?;

-- name: CreateFiscalOrderStatus :execresult
INSERT INTO fiscal_order_status (name, background_color, border_color, text_color, sort_index,
    is_closed, is_default, is_pending_approval)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateFiscalStatusRecord :exec
UPDATE fiscal_order_status
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?,
    is_closed=?, is_default=?, is_pending_approval=?
WHERE fiscal_order_status_id=?;

-- name: DeleteFiscalOrderStatus :exec
DELETE FROM fiscal_order_status WHERE fiscal_order_status_id=?;

-- ── Fiscal Priority ───────────────────────────────────────────────────────────

-- name: ListFiscalOrderPriorities :many
SELECT fiscal_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM fiscal_order_priorities
ORDER BY sort_index;

-- name: GetFiscalOrderPriorityByID :one
SELECT fiscal_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM fiscal_order_priorities
WHERE fiscal_order_priority_id=?;

-- name: CreateFiscalOrderPriority :execresult
INSERT INTO fiscal_order_priorities (name, background_color, border_color, text_color, sort_index)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateFiscalOrderPriority :exec
UPDATE fiscal_order_priorities
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?
WHERE fiscal_order_priority_id=?;

-- name: DeleteFiscalOrderPriority :exec
DELETE FROM fiscal_order_priorities WHERE fiscal_order_priority_id=?;

-- ── Fiscal Reason ─────────────────────────────────────────────────────────────

-- name: ListFiscalOrderReasons :many
SELECT r.fiscal_order_reason_id, r.name, r.initial_status_id,
       s.name AS initial_status_name
FROM fiscal_order_reason r
JOIN fiscal_order_status s ON s.fiscal_order_status_id = r.initial_status_id
WHERE r.deleted_at IS NULL
ORDER BY r.name;

-- name: GetFiscalOrderReasonByID :one
SELECT r.fiscal_order_reason_id, r.name, r.initial_status_id,
       s.name AS initial_status_name
FROM fiscal_order_reason r
JOIN fiscal_order_status s ON s.fiscal_order_status_id = r.initial_status_id
WHERE r.fiscal_order_reason_id=? AND r.deleted_at IS NULL;

-- name: CreateFiscalOrderReason :execresult
INSERT INTO fiscal_order_reason (name, initial_status_id) VALUES (?, ?);

-- name: UpdateFiscalOrderReason :exec
UPDATE fiscal_order_reason SET name=?, initial_status_id=?
WHERE fiscal_order_reason_id=? AND deleted_at IS NULL;

-- name: DeleteFiscalOrderReason :exec
UPDATE fiscal_order_reason SET deleted_at=NOW()
WHERE fiscal_order_reason_id=? AND deleted_at IS NULL;

-- ── Fiscal Status Transitions ─────────────────────────────────────────────────

-- name: ListFiscalStatusTransitions :many
SELECT id, from_status_id, to_status_id, access_level_id
FROM fiscal_order_status_transition;

-- name: CreateFiscalStatusTransition :execresult
INSERT INTO fiscal_order_status_transition (from_status_id, to_status_id, access_level_id)
VALUES (?, ?, ?);

-- name: DeleteFiscalStatusTransition :exec
DELETE FROM fiscal_order_status_transition WHERE id=?;

-- ── Fiscal Orders ─────────────────────────────────────────────────────────────

-- name: CreateFiscalOrder :execresult
INSERT INTO fiscal_order (requester_id, requester_name, assigned_to_id, assigned_to_name,
    salesperson_name, client_code, client_name, client_alias_name,
    status_id, priority_id, reason_id, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetFiscalOrderByID :one
SELECT o.fiscal_order_id, o.requester_id, o.requester_name, o.assigned_to_id, o.assigned_to_name,
       o.salesperson_name, o.client_code, o.client_name, o.client_alias_name,
       o.status_id, o.priority_id, o.reason_id, o.description,
       o.captured_at, o.resolution, o.closed_at, o.created_at, o.updated_at,
       s.name AS status_name, s.background_color AS status_bg, s.border_color AS status_border,
       s.text_color AS status_text, s.is_closed AS status_is_closed,
       p.name AS priority_name, p.background_color AS priority_bg,
       r.name AS reason_name
FROM fiscal_order o
JOIN fiscal_order_status s     ON s.fiscal_order_status_id   = o.status_id
JOIN fiscal_order_priorities p ON p.fiscal_order_priority_id = o.priority_id
JOIN fiscal_order_reason r     ON r.fiscal_order_reason_id   = o.reason_id
WHERE o.fiscal_order_id=? AND o.deleted_at IS NULL;

-- name: UpdateFiscalOrder :exec
UPDATE fiscal_order
SET assigned_to_id=?, assigned_to_name=?, salesperson_name=?,
    client_code=?, client_name=?, client_alias_name=?,
    status_id=?, priority_id=?, reason_id=?, description=?,
    captured_at=?, resolution=?, closed_at=?
WHERE fiscal_order_id=? AND deleted_at IS NULL;

-- name: UpdateFiscalOrderStatus :exec
UPDATE fiscal_order SET status_id=?, captured_at=?, closed_at=?
WHERE fiscal_order_id=? AND deleted_at IS NULL;

-- name: AssignFiscalOrder :exec
UPDATE fiscal_order SET assigned_to_id=?, assigned_to_name=?
WHERE fiscal_order_id=? AND deleted_at IS NULL;

-- name: DeleteFiscalOrder :exec
UPDATE fiscal_order SET deleted_at=NOW()
WHERE fiscal_order_id=? AND deleted_at IS NULL;

-- ── Fiscal Attachments ────────────────────────────────────────────────────────

-- name: CreateFiscalOrderAttachment :execresult
INSERT INTO fiscal_order_attachment (fiscal_order_id, file_name, file_path, file_size)
VALUES (?, ?, ?, ?);

-- name: ListFiscalOrderAttachments :many
SELECT fiscal_order_attachment_id, fiscal_order_id, file_name, file_path, file_size, created_at
FROM fiscal_order_attachment
WHERE fiscal_order_id=? AND deleted_at IS NULL
ORDER BY created_at;

-- name: GetFiscalOrderAttachmentByID :one
SELECT fiscal_order_attachment_id, fiscal_order_id, file_name, file_path, file_size, created_at
FROM fiscal_order_attachment
WHERE fiscal_order_attachment_id=? AND deleted_at IS NULL;

-- name: DeleteFiscalOrderAttachment :exec
UPDATE fiscal_order_attachment SET deleted_at=NOW()
WHERE fiscal_order_attachment_id=? AND deleted_at IS NULL;
