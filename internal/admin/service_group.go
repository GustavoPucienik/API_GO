package admin

import (
	"context"
	"database/sql"
	"errors"

	"systemapi/internal/db"
)

type GroupService struct {
	sqlDB   *sql.DB
	queries *db.Queries
}

func NewGroupService(sqlDB *sql.DB, q *db.Queries) *GroupService {
	return &GroupService{sqlDB: sqlDB, queries: q}
}

func (s *GroupService) Create(ctx context.Context, req CreateGroupRequest) (GroupResponse, error) {
	if req.Name == "" {
		return GroupResponse{}, errors.New("nome do grupo é obrigatório")
	}

	_, err := s.queries.GetGroupByName(ctx, req.Name)
	if err == nil {
		return GroupResponse{}, errors.New("grupo já cadastrado")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return GroupResponse{}, err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return GroupResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	result, err := qtx.CreateGroup(ctx, db.CreateGroupParams{
		Name:     req.Name,
		ParentID: sql.NullInt32{},
	})
	if err != nil {
		return GroupResponse{}, err
	}
	groupID, _ := result.LastInsertId()

	for _, alID := range req.AccessLevelIDs {
		if err := qtx.AddGroupAccessLevel(ctx, db.AddGroupAccessLevelParams{
			GroupID: int32(groupID), AccessLevelID: int32(alID),
		}); err != nil {
			return GroupResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return GroupResponse{}, err
	}
	return GroupResponse{ID: groupID, Name: req.Name}, nil
}

func (s *GroupService) Update(ctx context.Context, id int64, req UpdateGroupRequest) (GroupResponse, error) {
	group, err := s.queries.GetGroupByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return GroupResponse{}, errors.New("grupo não encontrado")
	}
	if err != nil {
		return GroupResponse{}, err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return GroupResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	if req.Name != nil && *req.Name != "" {
		if err := qtx.UpdateGroupName(ctx, db.UpdateGroupNameParams{
			Name: *req.Name, GroupID: int32(id),
		}); err != nil {
			return GroupResponse{}, err
		}
		group.Name = *req.Name
	}

	if req.AccessLevelIDs != nil {
		current, err := qtx.ListAccessLevelsByGroupID(ctx, int32(id))
		if err != nil {
			return GroupResponse{}, err
		}
		currentSet := map[int32]bool{}
		for _, al := range current {
			currentSet[al.AccessLevelID] = true
		}
		newSet := map[int32]bool{}
		for _, alID := range req.AccessLevelIDs {
			newSet[int32(alID)] = true
		}
		// Remove os que não estão no novo set
		for _, al := range current {
			if !newSet[al.AccessLevelID] {
				if err := qtx.RemoveAllAccessLevelsFromGroup(ctx, int32(id)); err != nil {
					return GroupResponse{}, err
				}
				break
			}
		}
		// Adiciona os novos
		for _, alID := range req.AccessLevelIDs {
			if !currentSet[int32(alID)] {
				if err := qtx.AddGroupAccessLevel(ctx, db.AddGroupAccessLevelParams{
					GroupID: int32(id), AccessLevelID: int32(alID),
				}); err != nil {
					return GroupResponse{}, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return GroupResponse{}, err
	}

	var parentID *int64
	if group.ParentID.Valid {
		v := int64(group.ParentID.Int32)
		parentID = &v
	}
	return GroupResponse{ID: int64(group.GroupID), Name: group.Name, ParentID: parentID}, nil
}

func (s *GroupService) FindPaginated(ctx context.Context, page, limit int, search string) ([]GroupResponse, int64, error) {
	like := "%" + search + "%"
	offset := int32((page - 1) * limit)

	groups, err := s.queries.ListGroupsPaginated(ctx, db.ListGroupsPaginatedParams{
		Column1: search, Name: like, Limit: int32(limit), Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.queries.CountGroups(ctx, db.CountGroupsParams{
		Column1: search, Name: like,
	})
	if err != nil {
		return nil, 0, err
	}

	resp := make([]GroupResponse, len(groups))
	for i, g := range groups {
		r := GroupResponse{ID: int64(g.GroupID), Name: g.Name}
		if g.ParentID.Valid {
			v := int64(g.ParentID.Int32)
			r.ParentID = &v
		}
		resp[i] = r
	}
	return resp, total, nil
}

func (s *GroupService) FindByID(ctx context.Context, id int64) (GroupResponse, error) {
	group, err := s.queries.GetGroupByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return GroupResponse{}, errors.New("grupo não encontrado")
	}
	if err != nil {
		return GroupResponse{}, err
	}
	r := GroupResponse{ID: int64(group.GroupID), Name: group.Name}
	if group.ParentID.Valid {
		v := int64(group.ParentID.Int32)
		r.ParentID = &v
	}
	return r, nil
}

func (s *GroupService) Delete(ctx context.Context, id int64) error {
	_, err := s.queries.GetGroupByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("grupo não encontrado")
	}
	if err != nil {
		return err
	}
	return s.queries.SoftDeleteGroup(ctx, int32(id))
}
