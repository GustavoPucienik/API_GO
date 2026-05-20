-- ============================================================
-- USER
-- ============================================================

-- name: GetUserByEmail :one
SELECT user_id, name, last_name, email, password, is_active, created_at, updated_at, deleted_at
FROM `user`
WHERE email = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByID :one
SELECT user_id, name, last_name, email, password, is_active, created_at, updated_at, deleted_at
FROM `user`
WHERE user_id = ? AND deleted_at IS NULL
LIMIT 1;

-- name: ListUsers :many
SELECT user_id, name, last_name, email, is_active, created_at, updated_at
FROM `user`
WHERE deleted_at IS NULL
ORDER BY name ASC;

-- name: ListUsersPaginated :many
SELECT user_id, name, last_name, email, is_active, created_at, updated_at
FROM `user`
WHERE deleted_at IS NULL
  AND (? = '' OR name LIKE ? OR email LIKE ?)
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*)
FROM `user`
WHERE deleted_at IS NULL
  AND (? = '' OR name LIKE ? OR email LIKE ?);

-- name: CreateUser :execresult
INSERT INTO `user` (name, last_name, email, password, is_active)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateUser :exec
UPDATE `user`
SET name = ?, last_name = ?, email = ?, is_active = ?
WHERE user_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE `user`
SET deleted_at = NOW()
WHERE user_id = ? AND deleted_at IS NULL;

-- name: UpdateUserPassword :exec
UPDATE `user`
SET password = ?
WHERE user_id = ? AND deleted_at IS NULL;

-- ============================================================
-- GROUP
-- ============================================================

-- name: GetGroupByID :one
SELECT group_id, name, parent_id, created_at, updated_at, deleted_at
FROM `groups`
WHERE group_id = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetGroupByName :one
SELECT group_id, name, parent_id, created_at, updated_at, deleted_at
FROM `groups`
WHERE name = ? AND deleted_at IS NULL
LIMIT 1;

-- name: ListGroupsPaginated :many
SELECT group_id, name, parent_id, created_at, updated_at
FROM `groups`
WHERE deleted_at IS NULL
  AND (? = '' OR name LIKE ?)
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountGroups :one
SELECT COUNT(*)
FROM `groups`
WHERE deleted_at IS NULL
  AND (? = '' OR name LIKE ?);

-- name: CreateGroup :execresult
INSERT INTO `groups` (name, parent_id)
VALUES (?, ?);

-- name: UpdateGroupName :exec
UPDATE `groups`
SET name = ?
WHERE group_id = ? AND deleted_at IS NULL;

-- name: SoftDeleteGroup :exec
UPDATE `groups`
SET deleted_at = NOW()
WHERE group_id = ? AND deleted_at IS NULL;

-- ============================================================
-- ACCESS LEVEL
-- ============================================================

-- name: GetAccessLevelByID :one
SELECT access_level_id, name
FROM access_level
WHERE access_level_id = ?
LIMIT 1;

-- name: GetAccessLevelByName :one
SELECT access_level_id, name
FROM access_level
WHERE name = ?
LIMIT 1;

-- name: ListAccessLevelsPaginated :many
SELECT access_level_id, name
FROM access_level
WHERE (? = '' OR name LIKE ?)
ORDER BY name ASC
LIMIT ? OFFSET ?;

-- name: CountAccessLevels :one
SELECT COUNT(*)
FROM access_level
WHERE (? = '' OR name LIKE ?);

-- name: CreateAccessLevel :execresult
INSERT INTO access_level (name) VALUES (?);

-- name: UpdateAccessLevelName :exec
UPDATE access_level SET name = ? WHERE access_level_id = ?;

-- name: DeleteAccessLevel :execrows
DELETE FROM access_level WHERE access_level_id = ?;

-- ============================================================
-- MODULE
-- ============================================================

-- name: GetModuleByID :one
SELECT module_id, name FROM module WHERE module_id = ? LIMIT 1;

-- name: GetModuleByName :one
SELECT module_id, name FROM module WHERE name = ? LIMIT 1;

-- name: ListModules :many
SELECT module_id, name FROM module ORDER BY name ASC;

-- name: CreateModule :execresult
INSERT INTO module (name) VALUES (?);

-- name: UpdateModule :exec
UPDATE module SET name = ? WHERE module_id = ?;

-- name: DeleteModule :execrows
DELETE FROM module WHERE module_id = ?;

-- ============================================================
-- MENU
-- ============================================================

-- name: GetMenuByID :one
SELECT menu_id, name, sort_index, paths, module_id, parent_id, required_scope
FROM menu WHERE menu_id = ? LIMIT 1;

-- name: ListMenus :many
SELECT menu_id, name, sort_index, paths, module_id, parent_id, required_scope
FROM menu ORDER BY sort_index ASC;

-- name: CreateMenu :execresult
INSERT INTO menu (name, sort_index, paths, module_id, parent_id, required_scope)
VALUES (?, ?, ?, ?, ?, ?);

-- name: UpdateMenu :exec
UPDATE menu
SET name = ?, sort_index = ?, paths = ?, module_id = ?, parent_id = ?, required_scope = ?
WHERE menu_id = ?;

-- name: DeleteMenu :execrows
DELETE FROM menu WHERE menu_id = ?;

-- ============================================================
-- USER-GROUP RELATIONS
-- ============================================================

-- name: ListGroupsByUserID :many
SELECT g.group_id, g.name
FROM `groups` g
INNER JOIN users_groups ug ON ug.group_id = g.group_id
WHERE ug.user_id = ? AND g.deleted_at IS NULL;

-- name: AddUserToGroup :exec
INSERT INTO users_groups (user_id, group_id) VALUES (?, ?)
ON DUPLICATE KEY UPDATE user_id = user_id;

-- name: RemoveAllGroupsFromUser :exec
DELETE FROM users_groups WHERE user_id = ?;

-- ============================================================
-- USER-ACCESS LEVEL RELATIONS
-- ============================================================

-- name: ListAccessLevelsByUserID :many
SELECT al.access_level_id, al.name
FROM access_level al
INNER JOIN users_access_levels ual ON ual.access_level_id = al.access_level_id
WHERE ual.user_id = ?;

-- name: AddUserAccessLevel :exec
INSERT INTO users_access_levels (user_id, access_level_id) VALUES (?, ?)
ON DUPLICATE KEY UPDATE user_id = user_id;

-- name: RemoveAllAccessLevelsFromUser :exec
DELETE FROM users_access_levels WHERE user_id = ?;

-- ============================================================
-- GROUP-ACCESS LEVEL RELATIONS
-- ============================================================

-- name: ListAccessLevelsByGroupID :many
SELECT al.access_level_id, al.name
FROM access_level al
INNER JOIN groups_access_levels gal ON gal.access_level_id = al.access_level_id
WHERE gal.group_id = ?;

-- name: AddGroupAccessLevel :exec
INSERT INTO groups_access_levels (group_id, access_level_id) VALUES (?, ?)
ON DUPLICATE KEY UPDATE group_id = group_id;

-- name: RemoveAllAccessLevelsFromGroup :exec
DELETE FROM groups_access_levels WHERE group_id = ?;

-- ============================================================
-- ACCESS LEVEL-MODULE RELATIONS
-- ============================================================

-- name: ListModulePermsByAccessLevelID :many
SELECT alm.access_level_module_id, alm.access_level_id, alm.module_id,
       alm.data_scope, alm.permission, alm.can_be_assigned, m.name AS module_name
FROM access_levels_modules alm
INNER JOIN module m ON m.module_id = alm.module_id
WHERE alm.access_level_id = ?;

-- name: GetModulePermForAccessLevel :one
SELECT access_level_module_id, access_level_id, module_id, data_scope, permission, can_be_assigned
FROM access_levels_modules
WHERE access_level_id = ? AND module_id = ?
LIMIT 1;

-- name: CreateAccessLevelModule :exec
INSERT INTO access_levels_modules (access_level_id, module_id, permission, data_scope, can_be_assigned)
VALUES (?, ?, ?, ?, ?);

-- name: UpdateAccessLevelModule :exec
UPDATE access_levels_modules
SET permission = ?, data_scope = ?, can_be_assigned = ?
WHERE access_level_id = ? AND module_id = ?;

-- name: DeleteAccessLevelModule :exec
DELETE FROM access_levels_modules
WHERE access_level_id = ? AND module_id = ?;

-- name: DeleteAllModulesFromAccessLevel :exec
DELETE FROM access_levels_modules WHERE access_level_id = ?;

-- ============================================================
-- USER-GROUP (by group)
-- ============================================================

-- name: ListUsersByGroupID :many
SELECT u.user_id, u.name, u.last_name, u.email, u.is_active
FROM `user` u
INNER JOIN users_groups ug ON ug.user_id = u.user_id
WHERE ug.group_id = ? AND u.deleted_at IS NULL
ORDER BY u.name ASC;

-- name: RemoveAllUsersFromGroup :exec
DELETE FROM users_groups WHERE group_id = ?;
