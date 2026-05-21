package logistic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"systemapi/internal/db"
)

type ConfigService struct {
	sqlDB *sql.DB
	q     *db.Queries
}

func NewConfigService(sqlDB *sql.DB, q *db.Queries) *ConfigService {
	return &ConfigService{sqlDB: sqlDB, q: q}
}

func (s *ConfigService) getOrCreateConfig(ctx context.Context) (db.LogisticConfig, error) {
	cfg, err := s.q.GetOrCreateLogisticConfig(ctx)
	if err == nil {
		return cfg, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return db.LogisticConfig{}, err
	}
	res, err := s.q.InsertLogisticConfig(ctx)
	if err != nil {
		return db.LogisticConfig{}, err
	}
	id64, _ := res.LastInsertId()
	cfg, err = s.q.GetOrCreateLogisticConfig(ctx)
	if err != nil {
		cfg.LogisticConfigID = int32(id64)
	}
	return cfg, nil
}

func (s *ConfigService) GetConfig(ctx context.Context) (ConfigResponse, error) {
	cfg, err := s.getOrCreateConfig(ctx)
	if err != nil {
		return ConfigResponse{}, err
	}

	levels, err := s.q.GetLogisticConfigAssignableAccessLevels(ctx, cfg.LogisticConfigID)
	if err != nil {
		return ConfigResponse{}, err
	}

	items := make([]AccessLevelItem, len(levels))
	for i, l := range levels {
		items[i] = AccessLevelItem{ID: l.AccessLevelID, Name: l.Name}
	}

	resp := ConfigResponse{AssignableAccessLevels: items}
	if cfg.SalaryDiscountAccessLevelID.Valid {
		v := cfg.SalaryDiscountAccessLevelID.Int32
		resp.SalaryDiscountAccessLevelID = &v
	}
	return resp, nil
}

func (s *ConfigService) UpdateSalaryDiscount(ctx context.Context, accessLevelID *int32) (ConfigResponse, error) {
	cfg, err := s.getOrCreateConfig(ctx)
	if err != nil {
		return ConfigResponse{}, err
	}
	nullID := sql.NullInt32{}
	if accessLevelID != nil {
		nullID = sql.NullInt32{Int32: *accessLevelID, Valid: true}
	}
	if err := s.q.UpdateLogisticConfigSalaryDiscount(ctx, db.UpdateLogisticConfigSalaryDiscountParams{
		SalaryDiscountAccessLevelID: nullID,
		LogisticConfigID:            cfg.LogisticConfigID,
	}); err != nil {
		return ConfigResponse{}, err
	}
	return s.GetConfig(ctx)
}

func (s *ConfigService) UpdateAssignableAccessLevels(ctx context.Context, accessLevelIDs []int32) (ConfigResponse, error) {
	cfg, err := s.getOrCreateConfig(ctx)
	if err != nil {
		return ConfigResponse{}, err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return ConfigResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.q.WithTx(tx)

	if err := qtx.DeleteLogisticConfigAssignableAccessLevels(ctx, cfg.LogisticConfigID); err != nil {
		return ConfigResponse{}, err
	}
	for _, alID := range accessLevelIDs {
		if err := qtx.InsertLogisticConfigAssignableAccessLevel(ctx, db.InsertLogisticConfigAssignableAccessLevelParams{
			LogisticConfigID: cfg.LogisticConfigID,
			AccessLevelID:    alID,
		}); err != nil {
			return ConfigResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return ConfigResponse{}, err
	}
	return s.GetConfig(ctx)
}

func (s *ConfigService) GetAssignableUsers(ctx context.Context) ([]AssignableUserResponse, error) {
	cfg, err := s.getOrCreateConfig(ctx)
	if err != nil {
		return nil, err
	}

	levels, err := s.q.GetLogisticConfigAssignableAccessLevels(ctx, cfg.LogisticConfigID)
	if err != nil {
		return nil, err
	}
	if len(levels) == 0 {
		return []AssignableUserResponse{}, nil
	}

	placeholders := strings.Repeat("?,", len(levels))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]any, len(levels)*2)
	for i, l := range levels {
		args[i] = l.AccessLevelID
		args[len(levels)+i] = l.AccessLevelID
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
	if result == nil {
		result = []AssignableUserResponse{}
	}
	return result, rows.Err()
}
