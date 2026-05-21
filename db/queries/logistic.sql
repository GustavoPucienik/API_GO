-- ── Logistic Status ───────────────────────────────────────────────────────────

-- name: ListLogisticOrderStatuses :many
SELECT logistic_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_pending_approval, is_rh_step
FROM logistic_order_status
ORDER BY sort_index;

-- name: GetLogisticOrderStatusByID :one
SELECT logistic_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_pending_approval, is_rh_step
FROM logistic_order_status
WHERE logistic_order_status_id=?;

-- name: CreateLogisticOrderStatus :execresult
INSERT INTO logistic_order_status (name, background_color, border_color, text_color,
    sort_index, is_closed, is_pending_approval, is_rh_step)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateLogisticStatusRecord :exec
UPDATE logistic_order_status
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?,
    is_closed=?, is_pending_approval=?, is_rh_step=?
WHERE logistic_order_status_id=?;

-- name: DeleteLogisticOrderStatus :exec
DELETE FROM logistic_order_status WHERE logistic_order_status_id=?;

-- ── Logistic Priority ─────────────────────────────────────────────────────────

-- name: ListLogisticOrderPriorities :many
SELECT logistic_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM logistic_order_priorities
ORDER BY sort_index;

-- name: GetLogisticOrderPriorityByID :one
SELECT logistic_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM logistic_order_priorities
WHERE logistic_order_priority_id=?;

-- name: CreateLogisticOrderPriority :execresult
INSERT INTO logistic_order_priorities (name, background_color, border_color, text_color, sort_index)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateLogisticOrderPriority :exec
UPDATE logistic_order_priorities
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?
WHERE logistic_order_priority_id=?;

-- name: DeleteLogisticOrderPriority :exec
DELETE FROM logistic_order_priorities WHERE logistic_order_priority_id=?;

-- ── Logistic Reason ───────────────────────────────────────────────────────────

-- name: ListLogisticOrderReasons :many
SELECT r.logistic_order_reason_id, r.name, r.initial_status_id,
       s.name AS initial_status_name
FROM logistic_order_reason r
JOIN logistic_order_status s ON s.logistic_order_status_id = r.initial_status_id
WHERE r.deleted_at IS NULL
ORDER BY r.name;

-- name: GetLogisticOrderReasonByID :one
SELECT r.logistic_order_reason_id, r.name, r.initial_status_id,
       s.name AS initial_status_name
FROM logistic_order_reason r
JOIN logistic_order_status s ON s.logistic_order_status_id = r.initial_status_id
WHERE r.logistic_order_reason_id=? AND r.deleted_at IS NULL;

-- name: CreateLogisticOrderReason :execresult
INSERT INTO logistic_order_reason (name, initial_status_id) VALUES (?, ?);

-- name: UpdateLogisticOrderReason :exec
UPDATE logistic_order_reason SET name=?, initial_status_id=?
WHERE logistic_order_reason_id=? AND deleted_at IS NULL;

-- name: DeleteLogisticOrderReason :exec
UPDATE logistic_order_reason SET deleted_at=NOW()
WHERE logistic_order_reason_id=? AND deleted_at IS NULL;

-- ── Logistic Status Transitions ───────────────────────────────────────────────

-- name: ListLogisticStatusTransitions :many
SELECT id, from_status_id, to_status_id, access_level_id
FROM logistic_order_status_transition;

-- name: CreateLogisticStatusTransition :execresult
INSERT INTO logistic_order_status_transition (from_status_id, to_status_id, access_level_id)
VALUES (?, ?, ?);

-- name: DeleteLogisticStatusTransition :exec
DELETE FROM logistic_order_status_transition WHERE id=?;

-- ── Logistic Operators ────────────────────────────────────────────────────────

-- name: ListLogisticOperators :many
SELECT logistic_operator_id, name, type FROM logistic_operator ORDER BY name;

-- name: GetLogisticOperatorByID :one
SELECT logistic_operator_id, name, type FROM logistic_operator WHERE logistic_operator_id=?;

-- name: CreateLogisticOperator :execresult
INSERT INTO logistic_operator (name, type) VALUES (?, ?);

-- name: UpdateLogisticOperator :exec
UPDATE logistic_operator SET name=?, type=? WHERE logistic_operator_id=?;

-- name: DeleteLogisticOperator :exec
DELETE FROM logistic_operator WHERE logistic_operator_id=?;

-- ── Logistic Orders ───────────────────────────────────────────────────────────

-- name: CreateLogisticOrder :execresult
INSERT INTO logistic_order (requester_id, requester_name, assigned_to_id, assigned_to_name,
    salesperson_name, client_code, client_name, client_alias_name,
    contact_type, contact_name, contact_phone, contact_mobile, contact_email,
    address_name, address_type, type_of_address, street, street_no, block,
    city, county, state, country, zip_code, building_floor_room,
    status_id, priority_id, reason_id, schedule_date, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetLogisticOrderByID :one
SELECT o.logistic_order_id, o.requester_id, o.requester_name, o.assigned_to_id, o.assigned_to_name,
       o.driver_name, o.checker_name, o.checker_signature, o.salesperson_name,
       o.client_code, o.client_name, o.client_alias_name,
       o.contact_type, o.contact_name, o.contact_phone, o.contact_mobile, o.contact_email,
       o.address_name, o.address_type, o.type_of_address, o.street, o.street_no, o.block,
       o.city, o.county, o.state, o.country, o.zip_code, o.building_floor_room,
       o.status_id, o.priority_id, o.reason_id, o.schedule_date, o.description,
       o.has_salary_discount, o.salary_discount_reason,
       o.captured_at, o.resolution, o.closed_at, o.created_at, o.updated_at,
       s.name AS status_name, s.background_color AS status_bg, s.border_color AS status_border,
       s.text_color AS status_text, s.is_closed AS status_is_closed, s.is_rh_step AS status_is_rh_step,
       p.name AS priority_name, p.background_color AS priority_bg,
       r.name AS reason_name
FROM logistic_order o
JOIN logistic_order_status s     ON s.logistic_order_status_id   = o.status_id
JOIN logistic_order_priorities p ON p.logistic_order_priority_id = o.priority_id
JOIN logistic_order_reason r     ON r.logistic_order_reason_id   = o.reason_id
WHERE o.logistic_order_id=? AND o.deleted_at IS NULL;

-- name: UpdateLogisticOrder :exec
UPDATE logistic_order
SET assigned_to_id=?, assigned_to_name=?, driver_name=?, checker_name=?,
    salesperson_name=?, client_code=?, client_name=?, client_alias_name=?,
    contact_type=?, contact_name=?, contact_phone=?, contact_mobile=?, contact_email=?,
    address_name=?, address_type=?, type_of_address=?, street=?, street_no=?, block=?,
    city=?, county=?, state=?, country=?, zip_code=?, building_floor_room=?,
    status_id=?, priority_id=?, reason_id=?, schedule_date=?, description=?,
    has_salary_discount=?, salary_discount_reason=?,
    captured_at=?, resolution=?, closed_at=?
WHERE logistic_order_id=? AND deleted_at IS NULL;

-- name: UpdateLogisticOrderStatus :exec
UPDATE logistic_order SET status_id=?, captured_at=?, closed_at=?
WHERE logistic_order_id=? AND deleted_at IS NULL;

-- name: AssignLogisticOrder :exec
UPDATE logistic_order SET assigned_to_id=?, assigned_to_name=?
WHERE logistic_order_id=? AND deleted_at IS NULL;

-- name: SaveLogisticCheckerSignature :exec
UPDATE logistic_order SET checker_name=?, checker_signature=?
WHERE logistic_order_id=? AND deleted_at IS NULL;

-- name: DeleteLogisticOrder :exec
UPDATE logistic_order SET deleted_at=NOW()
WHERE logistic_order_id=? AND deleted_at IS NULL;

-- ── Logistic Items ────────────────────────────────────────────────────────────

-- name: CreateLogisticOrderItem :execresult
INSERT INTO logistic_order_items (logistic_order_id, name, item_code, quantity)
VALUES (?, ?, ?, ?);

-- name: ListLogisticOrderItems :many
SELECT logistic_order_item_id, logistic_order_id, name, item_code, quantity
FROM logistic_order_items
WHERE logistic_order_id=? AND deleted_at IS NULL;

-- name: DeleteLogisticOrderItems :exec
UPDATE logistic_order_items SET deleted_at=NOW()
WHERE logistic_order_id=? AND deleted_at IS NULL;

-- ── Logistic Attachments ──────────────────────────────────────────────────────

-- name: CreateLogisticOrderAttachment :execresult
INSERT INTO logistic_order_attachment (logistic_order_id, file_name, file_path, file_size)
VALUES (?, ?, ?, ?);

-- name: ListLogisticOrderAttachments :many
SELECT logistic_order_attachment_id, logistic_order_id, file_name, file_path, file_size, created_at
FROM logistic_order_attachment
WHERE logistic_order_id=? AND deleted_at IS NULL
ORDER BY created_at;

-- name: GetLogisticOrderAttachmentByID :one
SELECT logistic_order_attachment_id, logistic_order_id, file_name, file_path, file_size, created_at
FROM logistic_order_attachment
WHERE logistic_order_attachment_id=? AND deleted_at IS NULL;

-- name: DeleteLogisticOrderAttachment :exec
UPDATE logistic_order_attachment SET deleted_at=NOW()
WHERE logistic_order_attachment_id=? AND deleted_at IS NULL;

-- ── Logistic Config ───────────────────────────────────────────────────────────

-- name: GetOrCreateLogisticConfig :one
SELECT logistic_config_id, salary_discount_access_level_id FROM logistic_config LIMIT 1;

-- name: InsertLogisticConfig :execresult
INSERT INTO logistic_config (salary_discount_access_level_id) VALUES (NULL);

-- name: UpdateLogisticConfigSalaryDiscount :exec
UPDATE logistic_config SET salary_discount_access_level_id=? WHERE logistic_config_id=?;

-- name: GetLogisticConfigAssignableAccessLevels :many
SELECT a.access_level_id, a.name
FROM logistic_config_assignable_access_level ca
JOIN access_level a ON a.access_level_id = ca.access_level_id
WHERE ca.logistic_config_id=?;

-- name: DeleteLogisticConfigAssignableAccessLevels :exec
DELETE FROM logistic_config_assignable_access_level WHERE logistic_config_id=?;

-- name: InsertLogisticConfigAssignableAccessLevel :exec
INSERT INTO logistic_config_assignable_access_level (logistic_config_id, access_level_id)
VALUES (?, ?);
