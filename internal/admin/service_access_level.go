package admin

import (
	"context"
	"database/sql"
	"errors"

	"systemapi/internal/db"
)

type AccessLevelService struct {
	sqlDB   *sql.DB
	queries *db.Queries
}

func NewAccessLevelService(sqlDB *sql.DB, q *db.Queries) *AccessLevelService {
	return &AccessLevelService{sqlDB: sqlDB, queries: q}
}

func calcPermission(m AccessLevelModuleInput) int32 {
	var p int32
	if m.Read {
		p |= 1
	}
	if m.Create {
		p |= 2
	}
	if m.Update {
		p |= 4
	}
	if m.Delete {
		p |= 8
	}
	return p
}

func (s *AccessLevelService) Create(ctx context.Context, req CreateAccessLevelRequest) (AccessLevelResponse, error) {
	if req.Name == "" {
		return AccessLevelResponse{}, errors.New("nome do nível de acesso é obrigatório")
	}

	_, err := s.queries.GetAccessLevelByName(ctx, req.Name)
	if err == nil {
		return AccessLevelResponse{}, errors.New("nível de acesso já existe")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return AccessLevelResponse{}, err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return AccessLevelResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	result, err := qtx.CreateAccessLevel(ctx, req.Name)
	if err != nil {
		return AccessLevelResponse{}, err
	}
	alID, _ := result.LastInsertId()

	for _, mod := range req.Modules {
		perm := calcPermission(mod)
		if perm == 0 {
			continue
		}
		if err := qtx.CreateAccessLevelModule(ctx, db.CreateAccessLevelModuleParams{
			AccessLevelID: int32(alID),
			ModuleID:      int32(mod.ModuleID),
			Permission:    perm,
			DataScope:     mod.DataScope,
			CanBeAssigned: mod.CanBeAssigned,
		}); err != nil {
			return AccessLevelResponse{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return AccessLevelResponse{}, err
	}
	return AccessLevelResponse{ID: alID, Name: req.Name}, nil
}

func (s *AccessLevelService) Update(ctx context.Context, id int64, req UpdateAccessLevelRequest) (AccessLevelResponse, error) {
	al, err := s.queries.GetAccessLevelByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return AccessLevelResponse{}, errors.New("nível de acesso não encontrado")
	}
	if err != nil {
		return AccessLevelResponse{}, err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return AccessLevelResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	if req.Name != nil && *req.Name != "" {
		if err := qtx.UpdateAccessLevelName(ctx, db.UpdateAccessLevelNameParams{
			Name: *req.Name, AccessLevelID: int32(id),
		}); err != nil {
			return AccessLevelResponse{}, err
		}
		al.Name = *req.Name
	}

	if req.Modules != nil {
		existing, err := qtx.ListModulePermsByAccessLevelID(ctx, int32(id))
		if err != nil {
			return AccessLevelResponse{}, err
		}
		existingMap := map[int32]db.ListModulePermsByAccessLevelIDRow{}
		for _, e := range existing {
			existingMap[e.ModuleID] = e
		}

		incomingIDs := map[int32]bool{}
		for _, mod := range req.Modules {
			mID := int32(mod.ModuleID)
			incomingIDs[mID] = true
			perm := calcPermission(mod)

			_, exists := existingMap[mID]
			if perm == 0 {
				if exists {
					if err := qtx.DeleteAccessLevelModule(ctx, db.DeleteAccessLevelModuleParams{
						AccessLevelID: int32(id), ModuleID: mID,
					}); err != nil {
						return AccessLevelResponse{}, err
					}
				}
				continue
			}
			if exists {
				if err := qtx.UpdateAccessLevelModule(ctx, db.UpdateAccessLevelModuleParams{
					Permission: perm, DataScope: mod.DataScope,
					CanBeAssigned: mod.CanBeAssigned,
					AccessLevelID: int32(id), ModuleID: mID,
				}); err != nil {
					return AccessLevelResponse{}, err
				}
			} else {
				if err := qtx.CreateAccessLevelModule(ctx, db.CreateAccessLevelModuleParams{
					AccessLevelID: int32(id), ModuleID: mID,
					Permission: perm, DataScope: mod.DataScope,
					CanBeAssigned: mod.CanBeAssigned,
				}); err != nil {
					return AccessLevelResponse{}, err
				}
			}
		}
		// Remove módulos que não vieram no request
		for _, e := range existing {
			if !incomingIDs[e.ModuleID] {
				if err := qtx.DeleteAccessLevelModule(ctx, db.DeleteAccessLevelModuleParams{
					AccessLevelID: int32(id), ModuleID: e.ModuleID,
				}); err != nil {
					return AccessLevelResponse{}, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return AccessLevelResponse{}, err
	}
	return AccessLevelResponse{ID: int64(al.AccessLevelID), Name: al.Name}, nil
}

func (s *AccessLevelService) FindPaginated(ctx context.Context, page, limit int, search string) ([]AccessLevelResponse, int64, error) {
	like := "%" + search + "%"
	offset := int32((page - 1) * limit)

	items, err := s.queries.ListAccessLevelsPaginated(ctx, db.ListAccessLevelsPaginatedParams{
		Column1: search, Name: like, Limit: int32(limit), Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.queries.CountAccessLevels(ctx, db.CountAccessLevelsParams{
		Column1: search, Name: like,
	})
	if err != nil {
		return nil, 0, err
	}

	resp := make([]AccessLevelResponse, len(items))
	for i, al := range items {
		resp[i] = AccessLevelResponse{ID: int64(al.AccessLevelID), Name: al.Name}
	}
	return resp, total, nil
}

func (s *AccessLevelService) FindByID(ctx context.Context, id int64) (AccessLevelResponse, error) {
	al, err := s.queries.GetAccessLevelByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return AccessLevelResponse{}, errors.New("nível de acesso não encontrado")
	}
	if err != nil {
		return AccessLevelResponse{}, err
	}
	return AccessLevelResponse{ID: int64(al.AccessLevelID), Name: al.Name}, nil
}

func (s *AccessLevelService) Delete(ctx context.Context, id int64) error {
	_, err := s.queries.GetAccessLevelByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("nível de acesso não encontrado")
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
	if err := qtx.DeleteAllModulesFromAccessLevel(ctx, int32(id)); err != nil {
		return err
	}
	if _, err := qtx.DeleteAccessLevel(ctx, int32(id)); err != nil {
		return err
	}
	return tx.Commit()
}
