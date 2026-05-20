package maintenance

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"systemapi/internal/db"
)

type ConfigService struct {
	sqlDB   *sql.DB
	queries *db.Queries
}

func NewConfigService(sqlDB *sql.DB, q *db.Queries) *ConfigService {
	return &ConfigService{sqlDB: sqlDB, queries: q}
}

func (s *ConfigService) getOrCreateConfigID(ctx context.Context) (int32, error) {
	id, err := s.queries.GetMaintenanceConfig(ctx)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}
	result, err := s.queries.CreateMaintenanceConfig(ctx)
	if err != nil {
		return 0, err
	}
	newID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int32(newID), nil
}

func (s *ConfigService) GetConfig(ctx context.Context) (MaintenanceConfigResponse, error) {
	configID, err := s.getOrCreateConfigID(ctx)
	if err != nil {
		return MaintenanceConfigResponse{}, err
	}

	accessLevelIDs, err := s.queries.GetConfigAssignableAccessLevelIDs(ctx, configID)
	if err != nil {
		return MaintenanceConfigResponse{}, err
	}

	ids := make([]int64, len(accessLevelIDs))
	for i, id := range accessLevelIDs {
		ids[i] = int64(id)
	}
	return MaintenanceConfigResponse{AssignableAccessLevelIDs: ids}, nil
}

func (s *ConfigService) UpdateAssignableAccessLevels(ctx context.Context, accessLevelIDs []int32) (MaintenanceConfigResponse, error) {
	configID, err := s.getOrCreateConfigID(ctx)
	if err != nil {
		return MaintenanceConfigResponse{}, err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return MaintenanceConfigResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	if err := qtx.DeleteConfigAssignableAccessLevels(ctx, configID); err != nil {
		return MaintenanceConfigResponse{}, err
	}

	for _, alID := range accessLevelIDs {
		if err := qtx.AddConfigAssignableAccessLevel(ctx, db.AddConfigAssignableAccessLevelParams{
			MaintenanceOrderConfigID: configID,
			AccessLevelID:            alID,
		}); err != nil {
			return MaintenanceConfigResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return MaintenanceConfigResponse{}, err
	}

	ids := make([]int64, len(accessLevelIDs))
	for i, id := range accessLevelIDs {
		ids[i] = int64(id)
	}
	return MaintenanceConfigResponse{AssignableAccessLevelIDs: ids}, nil
}

func (s *ConfigService) GetAssignableUsers(ctx context.Context) ([]AssignableUserResponse, error) {
	configID, err := s.getOrCreateConfigID(ctx)
	if err != nil {
		return nil, err
	}

	accessLevelIDs, err := s.queries.GetConfigAssignableAccessLevelIDs(ctx, configID)
	if err != nil {
		return nil, err
	}

	if len(accessLevelIDs) == 0 {
		return []AssignableUserResponse{}, nil
	}

	placeholders := strings.Repeat("?,", len(accessLevelIDs))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, len(accessLevelIDs)*2)
	for i, id := range accessLevelIDs {
		args[i] = id
		args[len(accessLevelIDs)+i] = id
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT u.user_id, u.name, u.last_name, u.email
		FROM `+"`user`"+` u
		WHERE u.is_active = 1 AND u.deleted_at IS NULL
		  AND (
		    u.user_id IN (
		      SELECT ual.user_id FROM users_access_levels ual
		      WHERE ual.access_level_id IN (%s)
		    )
		    OR
		    u.user_id IN (
		      SELECT ug.user_id FROM users_groups ug
		      INNER JOIN groups_access_levels gal ON gal.group_id = ug.group_id
		      WHERE gal.access_level_id IN (%s)
		    )
		  )
		ORDER BY u.name ASC
	`, placeholders, placeholders)

	rows, err := s.sqlDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AssignableUserResponse
	for rows.Next() {
		var u AssignableUserResponse
		var lastName sql.NullString
		if err := rows.Scan(&u.ID, &u.Name, &lastName, &u.Email); err != nil {
			return nil, err
		}
		if lastName.Valid {
			u.Lastname = &lastName.String
		}
		result = append(result, u)
	}
	return result, rows.Err()
}
