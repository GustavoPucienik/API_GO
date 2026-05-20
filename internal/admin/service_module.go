package admin

import (
	"context"
	"database/sql"
	"errors"

	"systemapi/internal/db"
)

type ModuleService struct {
	queries *db.Queries
}

func NewModuleService(sqlDB *sql.DB, q *db.Queries) *ModuleService {
	return &ModuleService{queries: q}
}

func (s *ModuleService) Create(ctx context.Context, req CreateModuleRequest) (ModuleResponse, error) {
	if req.Name == "" {
		return ModuleResponse{}, errors.New("nome do módulo é obrigatório")
	}
	_, err := s.queries.GetModuleByName(ctx, req.Name)
	if err == nil {
		return ModuleResponse{}, errors.New("módulo já existe")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return ModuleResponse{}, err
	}
	result, err := s.queries.CreateModule(ctx, req.Name)
	if err != nil {
		return ModuleResponse{}, err
	}
	id, _ := result.LastInsertId()
	return ModuleResponse{ID: id, Name: req.Name}, nil
}

func (s *ModuleService) Update(ctx context.Context, id int64, req UpdateModuleRequest) (ModuleResponse, error) {
	if req.Name == "" {
		return ModuleResponse{}, errors.New("nome do módulo é obrigatório")
	}
	_, err := s.queries.GetModuleByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return ModuleResponse{}, errors.New("módulo não encontrado")
	}
	if err != nil {
		return ModuleResponse{}, err
	}
	if err := s.queries.UpdateModule(ctx, db.UpdateModuleParams{Name: req.Name, ModuleID: int32(id)}); err != nil {
		return ModuleResponse{}, err
	}
	return ModuleResponse{ID: id, Name: req.Name}, nil
}

func (s *ModuleService) FindAll(ctx context.Context) ([]ModuleResponse, error) {
	modules, err := s.queries.ListModules(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]ModuleResponse, len(modules))
	for i, m := range modules {
		resp[i] = ModuleResponse{ID: int64(m.ModuleID), Name: m.Name}
	}
	return resp, nil
}

func (s *ModuleService) FindByID(ctx context.Context, id int64) (ModuleResponse, error) {
	m, err := s.queries.GetModuleByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return ModuleResponse{}, errors.New("módulo não encontrado")
	}
	if err != nil {
		return ModuleResponse{}, err
	}
	return ModuleResponse{ID: int64(m.ModuleID), Name: m.Name}, nil
}

func (s *ModuleService) Delete(ctx context.Context, id int64) error {
	_, err := s.queries.GetModuleByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("módulo não encontrado")
	}
	if err != nil {
		return err
	}
	n, err := s.queries.DeleteModule(ctx, int32(id))
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("módulo não encontrado")
	}
	return nil
}
