-- ── Status ────────────────────────────────────────────────────────────────────

-- name: ListMaintenanceOrderStatuses :many
SELECT * FROM maintenance_order_status ORDER BY sort_index ASC;

-- name: GetMaintenanceOrderStatusByID :one
SELECT * FROM maintenance_order_status WHERE maintenance_order_status_id = ?;

-- name: GetDefaultMaintenanceOrderStatus :one
SELECT * FROM maintenance_order_status WHERE is_default = 1 LIMIT 1;

-- name: GetFirstMaintenanceOrderStatusBySort :one
SELECT * FROM maintenance_order_status ORDER BY sort_index ASC LIMIT 1;

-- name: CreateMaintenanceOrderStatus :execresult
INSERT INTO maintenance_order_status (name, background_color, border_color, text_color, sort_index, is_closed, is_default)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: UpdateMaintenanceOrderStatus :exec
UPDATE maintenance_order_status
SET name = ?, background_color = ?, border_color = ?, text_color = ?, sort_index = ?, is_closed = ?, is_default = ?
WHERE maintenance_order_status_id = ?;

-- name: DeleteMaintenanceOrderStatus :exec
DELETE FROM maintenance_order_status WHERE maintenance_order_status_id = ?;

-- ── Priorities ────────────────────────────────────────────────────────────────

-- name: ListMaintenanceOrderPriorities :many
SELECT * FROM maintenance_order_priorities ORDER BY sort_index ASC;

-- name: GetMaintenanceOrderPriorityByID :one
SELECT * FROM maintenance_order_priorities WHERE maintenance_order_priorities_id = ?;

-- name: CreateMaintenanceOrderPriority :execresult
INSERT INTO maintenance_order_priorities (name, background_color, border_color, text_color, sort_index)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateMaintenanceOrderPriority :exec
UPDATE maintenance_order_priorities
SET name = ?, background_color = ?, border_color = ?, text_color = ?, sort_index = ?
WHERE maintenance_order_priorities_id = ?;

-- name: DeleteMaintenanceOrderPriority :exec
DELETE FROM maintenance_order_priorities WHERE maintenance_order_priorities_id = ?;

-- ── Reasons ───────────────────────────────────────────────────────────────────

-- name: ListMaintenanceOrderReasons :many
SELECT * FROM maintenance_order_reason WHERE deleted_at IS NULL ORDER BY name ASC;

-- name: GetMaintenanceOrderReasonByID :one
SELECT * FROM maintenance_order_reason WHERE maintenance_order_reason_id = ? AND deleted_at IS NULL;

-- name: CreateMaintenanceOrderReason :execresult
INSERT INTO maintenance_order_reason (name) VALUES (?);

-- name: UpdateMaintenanceOrderReason :exec
UPDATE maintenance_order_reason SET name = ? WHERE maintenance_order_reason_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteMaintenanceOrderReason :exec
UPDATE maintenance_order_reason SET deleted_at = NOW() WHERE maintenance_order_reason_id = ? AND deleted_at IS NULL;

-- ── Note templates ────────────────────────────────────────────────────────────

-- name: ListMaintenanceOrderNoteTemplates :many
SELECT * FROM maintenance_order_note_template WHERE deleted_at IS NULL ORDER BY name ASC;

-- name: GetMaintenanceOrderNoteTemplateByID :one
SELECT * FROM maintenance_order_note_template WHERE maintenance_order_note_template_id = ? AND deleted_at IS NULL;

-- name: CreateMaintenanceOrderNoteTemplate :execresult
INSERT INTO maintenance_order_note_template (name, description) VALUES (?, ?);

-- name: UpdateMaintenanceOrderNoteTemplate :exec
UPDATE maintenance_order_note_template
SET name = ?, description = ?
WHERE maintenance_order_note_template_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteMaintenanceOrderNoteTemplate :exec
UPDATE maintenance_order_note_template SET deleted_at = NOW()
WHERE maintenance_order_note_template_id = ? AND deleted_at IS NULL;

-- ── Time entry types ──────────────────────────────────────────────────────────

-- name: ListMaintenanceOrderTimeEntryTypes :many
SELECT * FROM maintenance_order_time_entry_type ORDER BY name ASC;

-- name: GetMaintenanceOrderTimeEntryTypeByID :one
SELECT * FROM maintenance_order_time_entry_type WHERE maintenance_order_time_entry_type_id = ?;

-- name: CreateMaintenanceOrderTimeEntryType :execresult
INSERT INTO maintenance_order_time_entry_type (name, color) VALUES (?, ?);

-- name: UpdateMaintenanceOrderTimeEntryType :exec
UPDATE maintenance_order_time_entry_type SET name = ?, color = ?
WHERE maintenance_order_time_entry_type_id = ?;

-- name: DeleteMaintenanceOrderTimeEntryType :exec
DELETE FROM maintenance_order_time_entry_type WHERE maintenance_order_time_entry_type_id = ?;

-- ── Main order ────────────────────────────────────────────────────────────────

-- name: GetMaintenanceOrderByID :one
SELECT * FROM maintenance_order WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- name: CreateMaintenanceOrder :execresult
INSERT INTO maintenance_order (
    requester_id, requester_name, salesperson_name,
    client_code, client_name, client_alias_name,
    contact_type, contact_name, contact_phone, contact_mobile, contact_email,
    address_name, address_type, type_of_address, street, street_no, block,
    city, county, state, country, zip_code, building_floor_room,
    status_id, priority_id, reason_id, description
) VALUES (
    ?, ?, ?,
    ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?
);

-- name: SoftDeleteMaintenanceOrder :execrows
UPDATE maintenance_order SET deleted_at = NOW() WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- name: UpdateMaintenanceOrderSignature :exec
UPDATE maintenance_order SET client_signature = ? WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- name: UpdateMaintenanceOrderCapturedAt :exec
UPDATE maintenance_order SET captured_at = ? WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- ── Order items ───────────────────────────────────────────────────────────────

-- name: CreateMaintenanceOrderItem :exec
INSERT INTO maintenance_order_items (maintenance_order_id, name, item_code, quantity)
VALUES (?, ?, ?, ?);

-- name: ListMaintenanceOrderItems :many
SELECT * FROM maintenance_order_items WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- name: DeleteMaintenanceOrderItems :exec
UPDATE maintenance_order_items SET deleted_at = NOW()
WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- ── Resolution items ──────────────────────────────────────────────────────────

-- name: CreateMaintenanceOrderResolutionItem :exec
INSERT INTO maintenance_order_resolution_items (maintenance_order_id, name, item_code, quantity)
VALUES (?, ?, ?, ?);

-- name: ListMaintenanceOrderResolutionItems :many
SELECT * FROM maintenance_order_resolution_items WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- name: DeleteMaintenanceOrderResolutionItems :exec
UPDATE maintenance_order_resolution_items SET deleted_at = NOW()
WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- ── Time entries ──────────────────────────────────────────────────────────────

-- name: CreateMaintenanceOrderTimeEntry :exec
INSERT INTO maintenance_order_time_entry (maintenance_order_id, type_id, started_at, ended_at)
VALUES (?, ?, ?, ?);

-- name: ListMaintenanceOrderTimeEntries :many
SELECT
    te.maintenance_order_time_entry_id,
    te.maintenance_order_id,
    te.type_id,
    te.started_at,
    te.ended_at,
    te.created_at,
    te.updated_at,
    te.deleted_at,
    t.name  AS type_name,
    t.color AS type_color
FROM maintenance_order_time_entry te
JOIN maintenance_order_time_entry_type t ON t.maintenance_order_time_entry_type_id = te.type_id
WHERE te.maintenance_order_id = ? AND te.deleted_at IS NULL;

-- name: DeleteMaintenanceOrderTimeEntries :exec
UPDATE maintenance_order_time_entry SET deleted_at = NOW()
WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- ── Schedule dates ────────────────────────────────────────────────────────────

-- name: CreateMaintenanceOrderScheduleDate :execresult
INSERT INTO maintenance_order_schedule_date (maintenance_order_id, date) VALUES (?, ?);

-- name: ListMaintenanceOrderScheduleDates :many
SELECT * FROM maintenance_order_schedule_date
WHERE maintenance_order_id = ? AND deleted_at IS NULL
ORDER BY date ASC;

-- name: DeleteMaintenanceOrderScheduleDates :exec
UPDATE maintenance_order_schedule_date SET deleted_at = NOW()
WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- ── Schedule date technicians ─────────────────────────────────────────────────

-- name: CreateScheduleDateTechnician :exec
INSERT INTO maintenance_order_schedule_date_technician
    (maintenance_order_schedule_date_id, technician_id, technician_name)
VALUES (?, ?, ?);

-- name: ListScheduleDateTechnicians :many
SELECT * FROM maintenance_order_schedule_date_technician
WHERE maintenance_order_schedule_date_id = ? AND deleted_at IS NULL;

-- name: DeleteScheduleDateTechnicians :exec
UPDATE maintenance_order_schedule_date_technician SET deleted_at = NOW()
WHERE maintenance_order_schedule_date_id = ? AND deleted_at IS NULL;

-- name: ListAllTechniciansForOrder :many
SELECT sdt.*
FROM maintenance_order_schedule_date_technician sdt
JOIN maintenance_order_schedule_date sd
    ON sd.maintenance_order_schedule_date_id = sdt.maintenance_order_schedule_date_id
WHERE sd.maintenance_order_id = ? AND sd.deleted_at IS NULL AND sdt.deleted_at IS NULL;

-- name: CountTechnicianOnDate :one
SELECT COUNT(*) FROM maintenance_order_schedule_date_technician
WHERE maintenance_order_schedule_date_id = ? AND technician_id = ? AND deleted_at IS NULL;

-- ── Attachments ───────────────────────────────────────────────────────────────

-- name: CreateMaintenanceOrderAttachment :execresult
INSERT INTO maintenance_order_attachment (maintenance_order_id, file_name, file_path, file_size)
VALUES (?, ?, ?, ?);

-- name: GetMaintenanceOrderAttachmentByID :one
SELECT * FROM maintenance_order_attachment
WHERE maintenance_order_attachment_id = ? AND deleted_at IS NULL;

-- name: ListMaintenanceOrderAttachments :many
SELECT * FROM maintenance_order_attachment
WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteMaintenanceOrderAttachment :exec
UPDATE maintenance_order_attachment SET deleted_at = NOW()
WHERE maintenance_order_attachment_id = ? AND deleted_at IS NULL;

-- ── Checklist templates ───────────────────────────────────────────────────────

-- name: ListChecklistTemplates :many
SELECT * FROM maintenance_order_checklist_template WHERE deleted_at IS NULL ORDER BY name ASC;

-- name: ListChecklistTemplatesByReason :many
SELECT * FROM maintenance_order_checklist_template
WHERE deleted_at IS NULL AND reason_id = ?
ORDER BY name ASC;

-- name: GetChecklistTemplateByID :one
SELECT * FROM maintenance_order_checklist_template
WHERE maintenance_order_checklist_template_id = ? AND deleted_at IS NULL;

-- name: CreateChecklistTemplate :execresult
INSERT INTO maintenance_order_checklist_template (name, reason_id, active) VALUES (?, ?, ?);

-- name: UpdateChecklistTemplate :exec
UPDATE maintenance_order_checklist_template
SET name = ?, reason_id = ?, active = ?
WHERE maintenance_order_checklist_template_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteChecklistTemplate :exec
UPDATE maintenance_order_checklist_template SET deleted_at = NOW()
WHERE maintenance_order_checklist_template_id = ? AND deleted_at IS NULL;

-- ── Checklist items ───────────────────────────────────────────────────────────

-- name: ListChecklistItemsByTemplate :many
SELECT * FROM maintenance_order_checklist_item
WHERE template_id = ? AND deleted_at IS NULL
ORDER BY sort_order ASC;

-- name: GetChecklistItemByID :one
SELECT * FROM maintenance_order_checklist_item
WHERE maintenance_order_checklist_item_id = ? AND deleted_at IS NULL;

-- name: CreateChecklistItem :execresult
INSERT INTO maintenance_order_checklist_item
    (template_id, section, label, type, options, metadata, required, sort_order, parent_item_id, show_when_values)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateChecklistItem :exec
UPDATE maintenance_order_checklist_item
SET section = ?, label = ?, type = ?, options = ?, metadata = ?, required = ?, sort_order = ?,
    parent_item_id = ?, show_when_values = ?
WHERE maintenance_order_checklist_item_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteChecklistItem :exec
UPDATE maintenance_order_checklist_item SET deleted_at = NOW()
WHERE maintenance_order_checklist_item_id = ? AND deleted_at IS NULL;

-- ── Order checklists ──────────────────────────────────────────────────────────

-- name: ListOrderChecklists :many
SELECT * FROM maintenance_order_checklist
WHERE maintenance_order_id = ? AND deleted_at IS NULL;

-- name: GetOrderChecklistByID :one
SELECT * FROM maintenance_order_checklist
WHERE maintenance_order_checklist_id = ? AND deleted_at IS NULL;

-- name: CreateOrderChecklist :execresult
INSERT INTO maintenance_order_checklist (maintenance_order_id, template_id) VALUES (?, ?);

-- name: SoftDeleteOrderChecklist :exec
UPDATE maintenance_order_checklist SET deleted_at = NOW()
WHERE maintenance_order_checklist_id = ? AND deleted_at IS NULL;

-- name: UpdateOrderChecklistCompletedAt :exec
UPDATE maintenance_order_checklist SET completed_at = ?
WHERE maintenance_order_checklist_id = ? AND deleted_at IS NULL;

-- ── Checklist answers ─────────────────────────────────────────────────────────

-- name: CreateChecklistAnswer :exec
INSERT INTO maintenance_order_checklist_answer (checklist_id, item_id, value) VALUES (?, ?, ?);

-- name: ListChecklistAnswers :many
SELECT * FROM maintenance_order_checklist_answer
WHERE checklist_id = ? AND deleted_at IS NULL;

-- ── Config ────────────────────────────────────────────────────────────────────

-- name: GetMaintenanceConfig :one
SELECT maintenance_order_config_id FROM maintenance_order_config LIMIT 1;

-- name: CreateMaintenanceConfig :execresult
INSERT INTO maintenance_order_config () VALUES ();

-- name: GetConfigAssignableAccessLevelIDs :many
SELECT access_level_id
FROM maintenance_order_config_assignable_access_level
WHERE maintenance_order_config_id = ?;

-- name: DeleteConfigAssignableAccessLevels :exec
DELETE FROM maintenance_order_config_assignable_access_level
WHERE maintenance_order_config_id = ?;

-- name: AddConfigAssignableAccessLevel :exec
INSERT INTO maintenance_order_config_assignable_access_level
  (maintenance_order_config_id, access_level_id) VALUES (?, ?);
