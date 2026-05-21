-- ── SAP Clients ───────────────────────────────────────────────────────────────

-- name: SearchSapClients :many
SELECT sap_client_code, name, alias_name
FROM sap_client
WHERE deleted_at IS NULL
  AND (name LIKE CONCAT('%', ?, '%')
    OR alias_name LIKE CONCAT('%', ?, '%')
    OR sap_client_code LIKE CONCAT('%', ?, '%'))
ORDER BY name
LIMIT 20;

-- name: GetSapClientByCode :one
SELECT sap_client_code, name, alias_name, sap_updated_at
FROM sap_client
WHERE sap_client_code = ? AND deleted_at IS NULL;

-- name: UpsertSapClient :exec
INSERT INTO sap_client (sap_client_code, name, alias_name, sap_updated_at)
VALUES (?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  alias_name = VALUES(alias_name),
  sap_updated_at = VALUES(sap_updated_at),
  updated_at = CURRENT_TIMESTAMP;

-- name: DeleteSapClientsNotIn :exec
UPDATE sap_client SET deleted_at = CURRENT_TIMESTAMP
WHERE deleted_at IS NULL AND sap_client_code NOT IN (sqlc.slice(codes));

-- name: GetSapClientContacts :many
SELECT sap_client_contact_id, sap_client_code, name, contact_name, phone, mobile, email
FROM sap_client_contact
WHERE sap_client_code = ? AND deleted_at IS NULL
ORDER BY name;

-- name: UpsertSapClientContact :exec
INSERT INTO sap_client_contact (sap_client_code, name, contact_name, phone, mobile, email)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  contact_name = VALUES(contact_name),
  phone = VALUES(phone),
  mobile = VALUES(mobile),
  email = VALUES(email),
  updated_at = CURRENT_TIMESTAMP;

-- name: DeleteSapClientContacts :exec
UPDATE sap_client_contact SET deleted_at = CURRENT_TIMESTAMP
WHERE sap_client_code = ? AND deleted_at IS NULL;

-- name: GetSapClientAddresses :many
SELECT sap_client_address_id, sap_client_code, address_name, address_type, type_of_address,
       street, street_no, block, city, county, state, country, zip_code, building_floor_room
FROM sap_client_address
WHERE sap_client_code = ? AND deleted_at IS NULL
ORDER BY address_name;

-- name: UpsertSapClientAddress :exec
INSERT INTO sap_client_address (sap_client_code, address_name, address_type, type_of_address,
    street, street_no, block, city, county, state, country, zip_code, building_floor_room)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  type_of_address = VALUES(type_of_address),
  street = VALUES(street),
  street_no = VALUES(street_no),
  block = VALUES(block),
  city = VALUES(city),
  county = VALUES(county),
  state = VALUES(state),
  country = VALUES(country),
  zip_code = VALUES(zip_code),
  building_floor_room = VALUES(building_floor_room),
  updated_at = CURRENT_TIMESTAMP;

-- name: DeleteSapClientAddresses :exec
UPDATE sap_client_address SET deleted_at = CURRENT_TIMESTAMP
WHERE sap_client_code = ? AND deleted_at IS NULL;

-- ── SAP Items ─────────────────────────────────────────────────────────────────

-- name: SearchSapItems :many
SELECT sap_item_code, name
FROM sap_item
WHERE deleted_at IS NULL
  AND (name LIKE CONCAT('%', ?, '%')
    OR sap_item_code LIKE CONCAT('%', ?, '%'))
ORDER BY name
LIMIT 20;

-- name: GetSapItemByCode :one
SELECT sap_item_code, name, sap_updated_at
FROM sap_item
WHERE sap_item_code = ? AND deleted_at IS NULL;

-- name: UpsertSapItem :exec
INSERT INTO sap_item (sap_item_code, name, sap_updated_at)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  sap_updated_at = VALUES(sap_updated_at),
  updated_at = CURRENT_TIMESTAMP;

-- name: DeleteSapItemsNotIn :exec
UPDATE sap_item SET deleted_at = CURRENT_TIMESTAMP
WHERE deleted_at IS NULL AND sap_item_code NOT IN (sqlc.slice(codes));
