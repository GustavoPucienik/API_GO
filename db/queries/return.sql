-- ── Return Status ─────────────────────────────────────────────────────────────

-- name: ListReturnOrderStatuses :many
SELECT return_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_default, is_pending_approval
FROM return_order_status
ORDER BY sort_index;

-- name: GetReturnOrderStatusByID :one
SELECT return_order_status_id, name, background_color, border_color, text_color,
       sort_index, is_closed, is_default, is_pending_approval
FROM return_order_status
WHERE return_order_status_id=?;

-- name: CreateReturnOrderStatus :execresult
INSERT INTO return_order_status (name, background_color, border_color, text_color,
    sort_index, is_closed, is_default, is_pending_approval)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateReturnStatusRecord :exec
UPDATE return_order_status
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?,
    is_closed=?, is_default=?, is_pending_approval=?
WHERE return_order_status_id=?;

-- name: DeleteReturnOrderStatus :exec
DELETE FROM return_order_status WHERE return_order_status_id=?;

-- ── Return Priority ───────────────────────────────────────────────────────────

-- name: ListReturnOrderPriorities :many
SELECT return_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM return_order_priorities
ORDER BY sort_index;

-- name: GetReturnOrderPriorityByID :one
SELECT return_order_priority_id, name, background_color, border_color, text_color, sort_index
FROM return_order_priorities
WHERE return_order_priority_id=?;

-- name: CreateReturnOrderPriority :execresult
INSERT INTO return_order_priorities (name, background_color, border_color, text_color, sort_index)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateReturnOrderPriority :exec
UPDATE return_order_priorities
SET name=?, background_color=?, border_color=?, text_color=?, sort_index=?
WHERE return_order_priority_id=?;

-- name: DeleteReturnOrderPriority :exec
DELETE FROM return_order_priorities WHERE return_order_priority_id=?;

-- ── Return Reason ─────────────────────────────────────────────────────────────

-- name: ListReturnOrderReasons :many
SELECT r.return_order_reason_id, r.name, r.description, r.initial_status_id,
       s.name AS initial_status_name
FROM return_order_reason r
JOIN return_order_status s ON s.return_order_status_id = r.initial_status_id
WHERE r.deleted_at IS NULL
ORDER BY r.name;

-- name: GetReturnOrderReasonByID :one
SELECT r.return_order_reason_id, r.name, r.description, r.initial_status_id,
       s.name AS initial_status_name
FROM return_order_reason r
JOIN return_order_status s ON s.return_order_status_id = r.initial_status_id
WHERE r.return_order_reason_id=? AND r.deleted_at IS NULL;

-- name: CreateReturnOrderReason :execresult
INSERT INTO return_order_reason (name, description, initial_status_id) VALUES (?, ?, ?);

-- name: UpdateReturnOrderReason :exec
UPDATE return_order_reason SET name=?, description=?, initial_status_id=?
WHERE return_order_reason_id=? AND deleted_at IS NULL;

-- name: DeleteReturnOrderReason :exec
UPDATE return_order_reason SET deleted_at=NOW()
WHERE return_order_reason_id=? AND deleted_at IS NULL;

-- ── Return Credit Form ────────────────────────────────────────────────────────

-- name: ListReturnOrderCreditForms :many
SELECT return_order_credit_form_id, name
FROM return_order_credit_form
WHERE deleted_at IS NULL
ORDER BY name;

-- name: GetReturnOrderCreditFormByID :one
SELECT return_order_credit_form_id, name
FROM return_order_credit_form
WHERE return_order_credit_form_id=? AND deleted_at IS NULL;

-- name: CreateReturnOrderCreditForm :execresult
INSERT INTO return_order_credit_form (name) VALUES (?);

-- name: UpdateReturnOrderCreditForm :exec
UPDATE return_order_credit_form SET name=?
WHERE return_order_credit_form_id=? AND deleted_at IS NULL;

-- name: DeleteReturnOrderCreditForm :exec
UPDATE return_order_credit_form SET deleted_at=NOW()
WHERE return_order_credit_form_id=? AND deleted_at IS NULL;

-- ── Return Status Transitions ─────────────────────────────────────────────────

-- name: ListReturnStatusTransitions :many
SELECT id, from_status_id, to_status_id, access_level_id
FROM return_order_status_transition;

-- name: CreateReturnStatusTransition :execresult
INSERT INTO return_order_status_transition (from_status_id, to_status_id, access_level_id)
VALUES (?, ?, ?);

-- name: DeleteReturnStatusTransition :exec
DELETE FROM return_order_status_transition WHERE id=?;

-- ── Return Orders ─────────────────────────────────────────────────────────────

-- name: CreateReturnOrder :execresult
INSERT INTO return_order (requester_id, requester_name, assigned_to_id, assigned_to_name,
    salesperson_name, client_code, client_name, client_alias_name,
    contact_type, contact_name, contact_phone, contact_mobile, contact_email,
    address_name, address_type, type_of_address, street, street_no, block,
    city, county, state, country, zip_code, building_floor_room,
    status_id, priority_id, reason_id, credit_form_id,
    pix_key_type, pix_key_owner, pix_key,
    transfer_bank, transfer_agency, transfer_account,
    nfe, client_sent_return_note, internal_return, partial_return,
    schedule_date, description)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetReturnOrderByID :one
SELECT o.return_order_id, o.requester_id, o.requester_name, o.assigned_to_id, o.assigned_to_name,
       o.salesperson_name, o.client_code, o.client_name, o.client_alias_name,
       o.contact_type, o.contact_name, o.contact_phone, o.contact_mobile, o.contact_email,
       o.address_name, o.address_type, o.type_of_address, o.street, o.street_no, o.block,
       o.city, o.county, o.state, o.country, o.zip_code, o.building_floor_room,
       o.status_id, o.priority_id, o.reason_id, o.credit_form_id,
       o.pix_key_type, o.pix_key_owner, o.pix_key,
       o.transfer_bank, o.transfer_agency, o.transfer_account,
       o.nfe, o.client_sent_return_note, o.internal_return, o.partial_return,
       o.schedule_date, o.description, o.captured_at, o.resolution, o.closed_at,
       o.created_at, o.updated_at,
       s.name AS status_name, s.background_color AS status_bg, s.border_color AS status_border,
       s.text_color AS status_text, s.is_closed AS status_is_closed,
       p.name AS priority_name, p.background_color AS priority_bg,
       r.name AS reason_name,
       cf.name AS credit_form_name
FROM return_order o
JOIN return_order_status s     ON s.return_order_status_id   = o.status_id
JOIN return_order_priorities p ON p.return_order_priority_id = o.priority_id
JOIN return_order_reason r     ON r.return_order_reason_id   = o.reason_id
LEFT JOIN return_order_credit_form cf ON cf.return_order_credit_form_id = o.credit_form_id
WHERE o.return_order_id=? AND o.deleted_at IS NULL;

-- name: UpdateReturnOrder :exec
UPDATE return_order
SET assigned_to_id=?, assigned_to_name=?, salesperson_name=?,
    client_code=?, client_name=?, client_alias_name=?,
    contact_type=?, contact_name=?, contact_phone=?, contact_mobile=?, contact_email=?,
    address_name=?, address_type=?, type_of_address=?, street=?, street_no=?, block=?,
    city=?, county=?, state=?, country=?, zip_code=?, building_floor_room=?,
    status_id=?, priority_id=?, reason_id=?, credit_form_id=?,
    pix_key_type=?, pix_key_owner=?, pix_key=?,
    transfer_bank=?, transfer_agency=?, transfer_account=?,
    nfe=?, client_sent_return_note=?, internal_return=?, partial_return=?,
    schedule_date=?, description=?, captured_at=?, resolution=?, closed_at=?
WHERE return_order_id=? AND deleted_at IS NULL;

-- name: UpdateReturnOrderStatus :exec
UPDATE return_order SET status_id=?, captured_at=?, closed_at=?
WHERE return_order_id=? AND deleted_at IS NULL;

-- name: AssignReturnOrder :exec
UPDATE return_order SET assigned_to_id=?, assigned_to_name=?
WHERE return_order_id=? AND deleted_at IS NULL;

-- name: DeleteReturnOrder :exec
UPDATE return_order SET deleted_at=NOW()
WHERE return_order_id=? AND deleted_at IS NULL;

-- ── Return Items ──────────────────────────────────────────────────────────────

-- name: CreateReturnOrderItem :execresult
INSERT INTO return_order_items (return_order_id, name, item_code, quantity)
VALUES (?, ?, ?, ?);

-- name: ListReturnOrderItems :many
SELECT return_order_item_id, return_order_id, name, item_code, quantity
FROM return_order_items
WHERE return_order_id=? AND deleted_at IS NULL;

-- name: DeleteReturnOrderItems :exec
UPDATE return_order_items SET deleted_at=NOW()
WHERE return_order_id=? AND deleted_at IS NULL;

-- ── Return Attachments ────────────────────────────────────────────────────────

-- name: CreateReturnOrderAttachment :execresult
INSERT INTO return_order_attachment (return_order_id, file_name, file_path, file_size)
VALUES (?, ?, ?, ?);

-- name: ListReturnOrderAttachments :many
SELECT return_order_attachment_id, return_order_id, file_name, file_path, file_size, created_at
FROM return_order_attachment
WHERE return_order_id=? AND deleted_at IS NULL
ORDER BY created_at;

-- name: GetReturnOrderAttachmentByID :one
SELECT return_order_attachment_id, return_order_id, file_name, file_path, file_size, created_at
FROM return_order_attachment
WHERE return_order_attachment_id=? AND deleted_at IS NULL;

-- name: DeleteReturnOrderAttachment :exec
UPDATE return_order_attachment SET deleted_at=NOW()
WHERE return_order_attachment_id=? AND deleted_at IS NULL;

-- ── Return Config ─────────────────────────────────────────────────────────────

-- name: GetOrCreateReturnConfig :one
SELECT return_config_id FROM return_config LIMIT 1;

-- name: InsertReturnConfig :execresult
INSERT INTO return_config () VALUES ();

-- name: GetReturnConfigAssignableAccessLevels :many
SELECT a.access_level_id, a.name
FROM return_config_assignable_access_level ca
JOIN access_level a ON a.access_level_id = ca.access_level_id
WHERE ca.return_config_id=?;

-- name: DeleteReturnConfigAssignableAccessLevels :exec
DELETE FROM return_config_assignable_access_level WHERE return_config_id=?;

-- name: InsertReturnConfigAssignableAccessLevel :exec
INSERT INTO return_config_assignable_access_level (return_config_id, access_level_id)
VALUES (?, ?);
