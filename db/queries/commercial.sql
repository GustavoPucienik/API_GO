-- ── Commercial Status ─────────────────────────────────────────────────────────

-- name: ListCommercialOrderStatuses :many
SELECT commercial_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_pending_approval
FROM commercial_order_status
ORDER BY sort_index;

-- name: GetCommercialOrderStatusByID :one
SELECT commercial_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_pending_approval
FROM commercial_order_status
WHERE commercial_order_status_id=?;

-- name: CreateCommercialOrderStatus :execresult
INSERT INTO commercial_order_status (name, background_color, border_color, text_color,
    sort_index, is_closed, is_pending_approval)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateCommercialStatusRecord :exec
UPDATE commercial_order_status
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?,
    is_closed=?, is_pending_approval=?
WHERE commercial_order_status_id=?;

-- name: DeleteCommercialOrderStatus :exec
DELETE FROM commercial_order_status WHERE commercial_order_status_id=?;

-- ── Commercial Priority ───────────────────────────────────────────────────────

-- name: ListCommercialOrderPriorities :many
SELECT commercial_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM commercial_order_priorities
ORDER BY sort_index;

-- name: GetCommercialOrderPriorityByID :one
SELECT commercial_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM commercial_order_priorities
WHERE commercial_order_priority_id=?;

-- name: CreateCommercialOrderPriority :execresult
INSERT INTO commercial_order_priorities (name, background_color, border_color, text_color, sort_index)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateCommercialOrderPriority :exec
UPDATE commercial_order_priorities
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?
WHERE commercial_order_priority_id=?;

-- name: DeleteCommercialOrderPriority :exec
DELETE FROM commercial_order_priorities WHERE commercial_order_priority_id=?;

-- ── Commercial Reason ─────────────────────────────────────────────────────────

-- name: ListCommercialOrderReasons :many
SELECT r.commercial_order_reason_id, r.name, r.initial_status_id,
       s.name AS initial_status_name
FROM commercial_order_reason r
JOIN commercial_order_status s ON s.commercial_order_status_id = r.initial_status_id
WHERE r.deleted_at IS NULL
ORDER BY r.name;

-- name: GetCommercialOrderReasonByID :one
SELECT r.commercial_order_reason_id, r.name, r.initial_status_id,
       s.name AS initial_status_name
FROM commercial_order_reason r
JOIN commercial_order_status s ON s.commercial_order_status_id = r.initial_status_id
WHERE r.commercial_order_reason_id=? AND r.deleted_at IS NULL;

-- name: CreateCommercialOrderReason :execresult
INSERT INTO commercial_order_reason (name, initial_status_id) VALUES (?, ?);

-- name: UpdateCommercialOrderReason :exec
UPDATE commercial_order_reason SET name=?, initial_status_id=?
WHERE commercial_order_reason_id=? AND deleted_at IS NULL;

-- name: DeleteCommercialOrderReason :exec
UPDATE commercial_order_reason SET deleted_at=NOW()
WHERE commercial_order_reason_id=? AND deleted_at IS NULL;

-- ── Commercial Status Transitions ─────────────────────────────────────────────

-- name: ListCommercialStatusTransitions :many
SELECT id, from_status_id, to_status_id, access_level_id
FROM commercial_order_status_transition;

-- name: CreateCommercialStatusTransition :execresult
INSERT INTO commercial_order_status_transition (from_status_id, to_status_id, access_level_id)
VALUES (?, ?, ?);

-- name: DeleteCommercialStatusTransition :exec
DELETE FROM commercial_order_status_transition WHERE id=?;

-- ── Commercial Orders ─────────────────────────────────────────────────────────

-- name: CreateCommercialOrder :execresult
INSERT INTO commercial_order (requester_id, requester_name, assigned_to_id, assigned_to_name,
    salesperson_name, client_code, client_name, client_alias_name,
    status_id, priority_id, reason_id, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetCommercialOrderByID :one
SELECT o.commercial_order_id, o.requester_id, o.requester_name, o.assigned_to_id, o.assigned_to_name,
       o.salesperson_name, o.client_code, o.client_name, o.client_alias_name,
       o.status_id, o.priority_id, o.reason_id, o.description,
       o.captured_at, o.resolution, o.closed_at, o.created_at, o.updated_at,
       s.name AS status_name, s.background_color AS status_bg, s.border_color AS status_border,
       s.text_color AS status_text, s.is_closed AS status_is_closed,
       p.name AS priority_name, p.background_color AS priority_bg,
       r.name AS reason_name
FROM commercial_order o
JOIN commercial_order_status s     ON s.commercial_order_status_id   = o.status_id
JOIN commercial_order_priorities p ON p.commercial_order_priority_id = o.priority_id
JOIN commercial_order_reason r     ON r.commercial_order_reason_id   = o.reason_id
WHERE o.commercial_order_id=? AND o.deleted_at IS NULL;

-- name: UpdateCommercialOrder :exec
UPDATE commercial_order
SET assigned_to_id=?, assigned_to_name=?, salesperson_name=?,
    client_code=?, client_name=?, client_alias_name=?,
    status_id=?, priority_id=?, reason_id=?, description=?,
    captured_at=?, resolution=?, closed_at=?
WHERE commercial_order_id=? AND deleted_at IS NULL;

-- name: UpdateCommercialOrderStatus :exec
UPDATE commercial_order SET status_id=?, captured_at=?, closed_at=?
WHERE commercial_order_id=? AND deleted_at IS NULL;

-- name: AssignCommercialOrder :exec
UPDATE commercial_order SET assigned_to_id=?, assigned_to_name=?
WHERE commercial_order_id=? AND deleted_at IS NULL;

-- name: DeleteCommercialOrder :exec
UPDATE commercial_order SET deleted_at=NOW()
WHERE commercial_order_id=? AND deleted_at IS NULL;

-- ── Commercial Items ──────────────────────────────────────────────────────────

-- name: CreateCommercialOrderItem :execresult
INSERT INTO commercial_order_items (commercial_order_id, name, item_code, quantity)
VALUES (?, ?, ?, ?);

-- name: ListCommercialOrderItems :many
SELECT commercial_order_item_id, commercial_order_id, name, item_code, quantity
FROM commercial_order_items
WHERE commercial_order_id=? AND deleted_at IS NULL;

-- name: DeleteCommercialOrderItems :exec
UPDATE commercial_order_items SET deleted_at=NOW()
WHERE commercial_order_id=? AND deleted_at IS NULL;

-- ── Commercial Attachments ────────────────────────────────────────────────────

-- name: CreateCommercialOrderAttachment :execresult
INSERT INTO commercial_order_attachment (commercial_order_id, file_name, file_path, file_size)
VALUES (?, ?, ?, ?);

-- name: ListCommercialOrderAttachments :many
SELECT commercial_order_attachment_id, commercial_order_id, file_name, file_path, file_size, created_at
FROM commercial_order_attachment
WHERE commercial_order_id=? AND deleted_at IS NULL
ORDER BY created_at;

-- name: GetCommercialOrderAttachmentByID :one
SELECT commercial_order_attachment_id, commercial_order_id, file_name, file_path, file_size, created_at
FROM commercial_order_attachment
WHERE commercial_order_attachment_id=? AND deleted_at IS NULL;

-- name: DeleteCommercialOrderAttachment :exec
UPDATE commercial_order_attachment SET deleted_at=NOW()
WHERE commercial_order_attachment_id=? AND deleted_at IS NULL;

-- ── Commercial Config ─────────────────────────────────────────────────────────

-- name: GetOrCreateCommercialConfig :one
SELECT commercial_config_id FROM commercial_config LIMIT 1;

-- name: InsertCommercialConfig :execresult
INSERT INTO commercial_config () VALUES ();

-- name: GetCommercialConfigAssignableAccessLevels :many
SELECT a.access_level_id, a.name
FROM commercial_config_assignable_access_level ca
JOIN access_level a ON a.access_level_id = ca.access_level_id
WHERE ca.commercial_config_id = ?;

-- name: DeleteCommercialConfigAssignableAccessLevels :exec
DELETE FROM commercial_config_assignable_access_level WHERE commercial_config_id=?;

-- name: InsertCommercialConfigAssignableAccessLevel :exec
INSERT INTO commercial_config_assignable_access_level (commercial_config_id, access_level_id)
VALUES (?, ?);
