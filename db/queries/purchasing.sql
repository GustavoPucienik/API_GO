-- ── Purchasing Status ─────────────────────────────────────────────────────────

-- name: ListPurchasingOrderStatuses :many
SELECT purchasing_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_default, is_pending_approval
FROM purchasing_order_status
ORDER BY sort_index;

-- name: GetPurchasingOrderStatusByID :one
SELECT purchasing_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_default, is_pending_approval
FROM purchasing_order_status
WHERE purchasing_order_status_id=?;

-- name: CreatePurchasingOrderStatus :execresult
INSERT INTO purchasing_order_status (name, background_color, border_color, text_color, sort_index,
    is_closed, is_default, is_pending_approval)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdatePurchasingStatusRecord :exec
UPDATE purchasing_order_status
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?,
    is_closed=?, is_default=?, is_pending_approval=?
WHERE purchasing_order_status_id=?;

-- name: DeletePurchasingOrderStatus :exec
DELETE FROM purchasing_order_status WHERE purchasing_order_status_id=?;

-- ── Purchasing Priority ───────────────────────────────────────────────────────

-- name: ListPurchasingOrderPriorities :many
SELECT purchasing_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM purchasing_order_priorities
ORDER BY sort_index;

-- name: GetPurchasingOrderPriorityByID :one
SELECT purchasing_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM purchasing_order_priorities
WHERE purchasing_order_priority_id=?;

-- name: CreatePurchasingOrderPriority :execresult
INSERT INTO purchasing_order_priorities (name, background_color, border_color, text_color, sort_index)
VALUES (?, ?, ?, ?, ?);

-- name: UpdatePurchasingOrderPriority :exec
UPDATE purchasing_order_priorities
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?
WHERE purchasing_order_priority_id=?;

-- name: DeletePurchasingOrderPriority :exec
DELETE FROM purchasing_order_priorities WHERE purchasing_order_priority_id=?;

-- ── Purchasing Reason ─────────────────────────────────────────────────────────

-- name: ListPurchasingOrderReasons :many
SELECT r.purchasing_order_reason_id, r.name, r.initial_status_id,
       s.name AS initial_status_name
FROM purchasing_order_reason r
JOIN purchasing_order_status s ON s.purchasing_order_status_id = r.initial_status_id
WHERE r.deleted_at IS NULL
ORDER BY r.name;

-- name: GetPurchasingOrderReasonByID :one
SELECT r.purchasing_order_reason_id, r.name, r.initial_status_id,
       s.name AS initial_status_name
FROM purchasing_order_reason r
JOIN purchasing_order_status s ON s.purchasing_order_status_id = r.initial_status_id
WHERE r.purchasing_order_reason_id=? AND r.deleted_at IS NULL;

-- name: CreatePurchasingOrderReason :execresult
INSERT INTO purchasing_order_reason (name, initial_status_id) VALUES (?, ?);

-- name: UpdatePurchasingOrderReason :exec
UPDATE purchasing_order_reason SET name=?, initial_status_id=?
WHERE purchasing_order_reason_id=? AND deleted_at IS NULL;

-- name: DeletePurchasingOrderReason :exec
UPDATE purchasing_order_reason SET deleted_at=NOW()
WHERE purchasing_order_reason_id=? AND deleted_at IS NULL;

-- ── Purchasing Status Transitions ─────────────────────────────────────────────

-- name: ListPurchasingStatusTransitions :many
SELECT id, from_status_id, to_status_id, access_level_id
FROM purchasing_order_status_transition;

-- name: CreatePurchasingStatusTransition :execresult
INSERT INTO purchasing_order_status_transition (from_status_id, to_status_id, access_level_id)
VALUES (?, ?, ?);

-- name: DeletePurchasingStatusTransition :exec
DELETE FROM purchasing_order_status_transition WHERE id=?;

-- ── Purchasing Orders ─────────────────────────────────────────────────────────

-- name: CreatePurchasingOrder :execresult
INSERT INTO purchasing_order (requester_id, requester_name, assigned_to_id, assigned_to_name,
    salesperson_name, client_code, client_name, client_alias_name,
    status_id, priority_id, reason_id, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetPurchasingOrderByID :one
SELECT o.purchasing_order_id, o.requester_id, o.requester_name, o.assigned_to_id, o.assigned_to_name,
       o.salesperson_name, o.client_code, o.client_name, o.client_alias_name,
       o.status_id, o.priority_id, o.reason_id, o.description,
       o.captured_at, o.resolution, o.closed_at, o.created_at, o.updated_at,
       s.name AS status_name, s.background_color AS status_bg, s.border_color AS status_border,
       s.text_color AS status_text, s.is_closed AS status_is_closed,
       p.name AS priority_name, p.background_color AS priority_bg,
       r.name AS reason_name
FROM purchasing_order o
JOIN purchasing_order_status s     ON s.purchasing_order_status_id   = o.status_id
JOIN purchasing_order_priorities p ON p.purchasing_order_priority_id = o.priority_id
JOIN purchasing_order_reason r     ON r.purchasing_order_reason_id   = o.reason_id
WHERE o.purchasing_order_id=? AND o.deleted_at IS NULL;

-- name: UpdatePurchasingOrder :exec
UPDATE purchasing_order
SET assigned_to_id=?, assigned_to_name=?, salesperson_name=?,
    client_code=?, client_name=?, client_alias_name=?,
    status_id=?, priority_id=?, reason_id=?, description=?,
    captured_at=?, resolution=?, closed_at=?
WHERE purchasing_order_id=? AND deleted_at IS NULL;

-- name: UpdatePurchasingOrderStatus :exec
UPDATE purchasing_order SET status_id=?, captured_at=?, closed_at=?
WHERE purchasing_order_id=? AND deleted_at IS NULL;

-- name: AssignPurchasingOrder :exec
UPDATE purchasing_order SET assigned_to_id=?, assigned_to_name=?
WHERE purchasing_order_id=? AND deleted_at IS NULL;

-- name: DeletePurchasingOrder :exec
UPDATE purchasing_order SET deleted_at=NOW()
WHERE purchasing_order_id=? AND deleted_at IS NULL;

-- ── Purchasing Attachments ────────────────────────────────────────────────────

-- name: CreatePurchasingOrderAttachment :execresult
INSERT INTO purchasing_order_attachment (purchasing_order_id, file_name, file_path, file_size)
VALUES (?, ?, ?, ?);

-- name: ListPurchasingOrderAttachments :many
SELECT purchasing_order_attachment_id, purchasing_order_id, file_name, file_path, file_size, created_at
FROM purchasing_order_attachment
WHERE purchasing_order_id=? AND deleted_at IS NULL
ORDER BY created_at;

-- name: GetPurchasingOrderAttachmentByID :one
SELECT purchasing_order_attachment_id, purchasing_order_id, file_name, file_path, file_size, created_at
FROM purchasing_order_attachment
WHERE purchasing_order_attachment_id=? AND deleted_at IS NULL;

-- name: DeletePurchasingOrderAttachment :exec
UPDATE purchasing_order_attachment SET deleted_at=NOW()
WHERE purchasing_order_attachment_id=? AND deleted_at IS NULL;
