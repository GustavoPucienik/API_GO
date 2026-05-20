package admin

import (
	"context"
	"database/sql"
	"errors"

	"systemapi/internal/db"
)

type MenuService struct {
	queries *db.Queries
}

func NewMenuService(sqlDB *sql.DB, q *db.Queries) *MenuService {
	return &MenuService{queries: q}
}

func (s *MenuService) Create(ctx context.Context, req CreateMenuRequest) (MenuResponse, error) {
	if req.Name == "" {
		return MenuResponse{}, errors.New("nome do menu é obrigatório")
	}
	if req.Paths == "" {
		return MenuResponse{}, errors.New("paths do menu é obrigatório")
	}
	if req.ModuleID == 0 {
		return MenuResponse{}, errors.New("moduleId é obrigatório")
	}

	parentID := sql.NullInt32{}
	if req.ParentID != nil {
		parentID = sql.NullInt32{Int32: int32(*req.ParentID), Valid: true}
	}
	reqScope := sql.NullString{}
	if req.RequiredScope != nil && *req.RequiredScope != "" {
		reqScope = sql.NullString{String: *req.RequiredScope, Valid: true}
	}

	result, err := s.queries.CreateMenu(ctx, db.CreateMenuParams{
		Name:          req.Name,
		SortIndex:     req.SortIndex,
		Paths:         req.Paths,
		ModuleID:      int32(req.ModuleID),
		ParentID:      parentID,
		RequiredScope: reqScope,
	})
	if err != nil {
		return MenuResponse{}, err
	}
	id, _ := result.LastInsertId()
	return menuToResponse(db.Menu{
		MenuID:        int32(id),
		Name:          req.Name,
		SortIndex:     req.SortIndex,
		Paths:         req.Paths,
		ModuleID:      int32(req.ModuleID),
		ParentID:      parentID,
		RequiredScope: reqScope,
	}), nil
}

func (s *MenuService) Update(ctx context.Context, id int64, req UpdateMenuRequest) (MenuResponse, error) {
	_, err := s.queries.GetMenuByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return MenuResponse{}, errors.New("menu não encontrado")
	}
	if err != nil {
		return MenuResponse{}, err
	}

	parentID := sql.NullInt32{}
	if req.ParentID != nil {
		parentID = sql.NullInt32{Int32: int32(*req.ParentID), Valid: true}
	}
	reqScope := sql.NullString{}
	if req.RequiredScope != nil && *req.RequiredScope != "" {
		reqScope = sql.NullString{String: *req.RequiredScope, Valid: true}
	}

	if err := s.queries.UpdateMenu(ctx, db.UpdateMenuParams{
		Name:          req.Name,
		SortIndex:     req.SortIndex,
		Paths:         req.Paths,
		ModuleID:      int32(req.ModuleID),
		ParentID:      parentID,
		RequiredScope: reqScope,
		MenuID:        int32(id),
	}); err != nil {
		return MenuResponse{}, err
	}

	updated, err := s.queries.GetMenuByID(ctx, int32(id))
	if err != nil {
		return MenuResponse{}, err
	}
	return menuToResponse(updated), nil
}

func (s *MenuService) FindAll(ctx context.Context) ([]MenuResponse, error) {
	menus, err := s.queries.ListMenus(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]MenuResponse, len(menus))
	for i, m := range menus {
		resp[i] = menuToResponse(m)
	}
	return resp, nil
}

func (s *MenuService) FindByID(ctx context.Context, id int64) (MenuResponse, error) {
	m, err := s.queries.GetMenuByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return MenuResponse{}, errors.New("menu não encontrado")
	}
	if err != nil {
		return MenuResponse{}, err
	}
	return menuToResponse(m), nil
}

func (s *MenuService) Delete(ctx context.Context, id int64) error {
	_, err := s.queries.GetMenuByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("menu não encontrado")
	}
	if err != nil {
		return err
	}
	_, err = s.queries.DeleteMenu(ctx, int32(id))
	return err
}

func menuToResponse(m db.Menu) MenuResponse {
	r := MenuResponse{
		ID:        int64(m.MenuID),
		Name:      m.Name,
		SortIndex: m.SortIndex,
		Paths:     m.Paths,
		ModuleID:  int64(m.ModuleID),
	}
	if m.ParentID.Valid {
		v := int64(m.ParentID.Int32)
		r.ParentID = &v
	}
	if m.RequiredScope.Valid {
		r.RequiredScope = &m.RequiredScope.String
	}
	return r
}
