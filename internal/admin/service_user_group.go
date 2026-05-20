package admin

import (
	"context"
	"database/sql"
	"errors"

	"systemapi/internal/db"
)

type UserGroupService struct {
	sqlDB   *sql.DB
	queries *db.Queries
}

func NewUserGroupService(sqlDB *sql.DB, q *db.Queries) *UserGroupService {
	return &UserGroupService{sqlDB: sqlDB, queries: q}
}

func (s *UserGroupService) FindUsersByGroup(ctx context.Context, groupID int32) ([]UserGroupUserResponse, error) {
	_, err := s.queries.GetGroupByID(ctx, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("grupo não encontrado")
	}
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.ListUsersByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	resp := make([]UserGroupUserResponse, len(rows))
	for i, u := range rows {
		r := UserGroupUserResponse{
			ID:       int64(u.UserID),
			Name:     u.Name,
			Email:    u.Email,
			IsActive: u.IsActive,
		}
		if u.LastName.Valid {
			r.Lastname = &u.LastName.String
		}
		resp[i] = r
	}
	return resp, nil
}

func (s *UserGroupService) SyncUsersInGroup(ctx context.Context, groupID int32, userIDs []int32) error {
	_, err := s.queries.GetGroupByID(ctx, groupID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("grupo não encontrado")
	}
	if err != nil {
		return err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	if err := qtx.RemoveAllUsersFromGroup(ctx, groupID); err != nil {
		return err
	}

	for _, uid := range userIDs {
		if err := qtx.AddUserToGroup(ctx, db.AddUserToGroupParams{
			UserID:  uid,
			GroupID: groupID,
		}); err != nil {
			return err
		}
	}

	return tx.Commit()
}
