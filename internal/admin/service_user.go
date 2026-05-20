package admin

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"systemapi/internal/db"

	"golang.org/x/crypto/bcrypt"
)

const (
	allowedDomain = "@trcdistribuicao.com.br"
	saltRounds    = 10
)

type UserService struct {
	sqlDB   *sql.DB
	queries *db.Queries
}

func NewUserService(sqlDB *sql.DB, q *db.Queries) *UserService {
	return &UserService{sqlDB: sqlDB, queries: q}
}

func (s *UserService) Create(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return UserResponse{}, errors.New("nome é obrigatório")
	}
	if strings.TrimSpace(req.Email) == "" {
		return UserResponse{}, errors.New("email é obrigatório")
	}
	if strings.TrimSpace(req.Password) == "" {
		return UserResponse{}, errors.New("senha é obrigatória")
	}
	if !strings.HasSuffix(req.Email, allowedDomain) {
		return UserResponse{}, errors.New("apenas emails " + allowedDomain + " são permitidos")
	}

	_, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return UserResponse{}, errors.New("email já cadastrado")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return UserResponse{}, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), saltRounds)
	if err != nil {
		return UserResponse{}, err
	}

	lastName := sql.NullString{}
	if req.Lastname != nil && *req.Lastname != "" {
		lastName = sql.NullString{String: *req.Lastname, Valid: true}
	}

	result, err := s.queries.CreateUser(ctx, db.CreateUserParams{
		Name: req.Name, LastName: lastName,
		Email: req.Email, Password: string(hashed), IsActive: false,
	})
	if err != nil {
		return UserResponse{}, err
	}
	id, _ := result.LastInsertId()

	created, err := s.queries.GetUserByID(ctx, int32(id))
	if err != nil {
		return UserResponse{}, err
	}
	return dbUserToResponse(created), nil
}

func (s *UserService) Update(ctx context.Context, id int64, req UpdateUserRequest) (UserResponse, error) {
	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return UserResponse{}, err
	}
	defer tx.Rollback()

	qtx := s.queries.WithTx(tx)

	existing, err := qtx.GetUserByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return UserResponse{}, errors.New("usuário não encontrado")
	}
	if err != nil {
		return UserResponse{}, err
	}

	name := existing.Name
	if req.Name != "" {
		name = req.Name
	}
	email := existing.Email
	if req.Email != "" {
		email = req.Email
	}
	lastName := existing.LastName
	if req.Lastname != nil {
		lastName = sql.NullString{String: *req.Lastname, Valid: *req.Lastname != ""}
	}
	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	if err := qtx.UpdateUser(ctx, db.UpdateUserParams{
		Name: name, LastName: lastName, Email: email, IsActive: isActive, UserID: int32(id),
	}); err != nil {
		return UserResponse{}, err
	}

	if req.Groups != nil {
		if err := qtx.RemoveAllGroupsFromUser(ctx, int32(id)); err != nil {
			return UserResponse{}, err
		}
		for _, gID := range req.Groups {
			if err := qtx.AddUserToGroup(ctx, db.AddUserToGroupParams{
				UserID: int32(id), GroupID: int32(gID),
			}); err != nil {
				return UserResponse{}, err
			}
		}
	}

	if req.AccessLevels != nil {
		if err := qtx.RemoveAllAccessLevelsFromUser(ctx, int32(id)); err != nil {
			return UserResponse{}, err
		}
		for _, alID := range req.AccessLevels {
			if err := qtx.AddUserAccessLevel(ctx, db.AddUserAccessLevelParams{
				UserID: int32(id), AccessLevelID: int32(alID),
			}); err != nil {
				return UserResponse{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return UserResponse{}, err
	}

	updated, err := s.queries.GetUserByID(ctx, int32(id))
	if err != nil {
		return UserResponse{}, err
	}
	return dbUserToResponse(updated), nil
}

func (s *UserService) FindAll(ctx context.Context) ([]UserResponse, error) {
	users, err := s.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	resp := make([]UserResponse, len(users))
	for i, u := range users {
		r := UserResponse{ID: int64(u.UserID), Name: u.Name, Email: u.Email, IsActive: u.IsActive}
		if u.LastName.Valid {
			r.Lastname = &u.LastName.String
		}
		resp[i] = r
	}
	return resp, nil
}

func (s *UserService) FindPaginated(ctx context.Context, page, limit int, search string) ([]UserResponse, int64, error) {
	like := "%" + search + "%"
	offset := int32((page - 1) * limit)

	users, err := s.queries.ListUsersPaginated(ctx, db.ListUsersPaginatedParams{
		Column1: search, Name: like, Email: like,
		Limit: int32(limit), Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.queries.CountUsers(ctx, db.CountUsersParams{
		Column1: search, Name: like, Email: like,
	})
	if err != nil {
		return nil, 0, err
	}
	resp := make([]UserResponse, len(users))
	for i, u := range users {
		r := UserResponse{ID: int64(u.UserID), Name: u.Name, Email: u.Email, IsActive: u.IsActive}
		if u.LastName.Valid {
			r.Lastname = &u.LastName.String
		}
		resp[i] = r
	}
	return resp, total, nil
}

func (s *UserService) FindByID(ctx context.Context, id int64) (UserResponse, error) {
	user, err := s.queries.GetUserByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return UserResponse{}, errors.New("usuário não encontrado")
	}
	if err != nil {
		return UserResponse{}, err
	}
	return dbUserToResponse(user), nil
}

func (s *UserService) GetWithRelations(ctx context.Context, id int64) (UserWithRelationsResponse, error) {
	user, err := s.queries.GetUserByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return UserWithRelationsResponse{}, errors.New("usuário não encontrado")
	}
	if err != nil {
		return UserWithRelationsResponse{}, err
	}

	groups, err := s.queries.ListGroupsByUserID(ctx, int32(id))
	if err != nil {
		return UserWithRelationsResponse{}, err
	}
	accessLevels, err := s.queries.ListAccessLevelsByUserID(ctx, int32(id))
	if err != nil {
		return UserWithRelationsResponse{}, err
	}

	gBasics := make([]GroupBasic, len(groups))
	for i, g := range groups {
		gBasics[i] = GroupBasic{ID: int64(g.GroupID), Name: g.Name}
	}
	alBasics := make([]AccessLevelBasic, len(accessLevels))
	for i, al := range accessLevels {
		alBasics[i] = AccessLevelBasic{ID: int64(al.AccessLevelID), Name: al.Name}
	}
	return UserWithRelationsResponse{
		UserResponse: dbUserToResponse(user),
		Groups:       gBasics,
		AccessLevels: alBasics,
	}, nil
}

func (s *UserService) FindByModuleAndScope(ctx context.Context, moduleName, scope string) ([]UserResponse, error) {
	rows, err := s.sqlDB.QueryContext(ctx, `
		SELECT DISTINCT u.user_id, u.name, u.last_name, u.email, u.is_active
		FROM `+"`user`"+` u
		INNER JOIN users_access_levels ual ON ual.user_id = u.user_id
		INNER JOIN access_levels_modules alm ON alm.access_level_id = ual.access_level_id
		INNER JOIN module m ON m.module_id = alm.module_id
		WHERE m.name = ? AND alm.data_scope = ?
		  AND u.is_active = 1 AND u.deleted_at IS NULL
		ORDER BY u.name ASC`, moduleName, scope)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBasicUserRows(rows)
}

func (s *UserService) FindAssignableByModule(ctx context.Context, moduleName string) ([]UserResponse, error) {
	rows, err := s.sqlDB.QueryContext(ctx, `
		SELECT DISTINCT u.user_id, u.name, u.last_name, u.email, u.is_active
		FROM `+"`user`"+` u
		WHERE u.is_active = 1 AND u.deleted_at IS NULL
		  AND (
		    u.user_id IN (
		      SELECT ual.user_id FROM users_access_levels ual
		      INNER JOIN access_levels_modules alm ON alm.access_level_id = ual.access_level_id
		      INNER JOIN module m ON m.module_id = alm.module_id
		      WHERE m.name = ? AND alm.can_be_assigned = 1
		    )
		    OR u.user_id IN (
		      SELECT ug.user_id FROM users_groups ug
		      INNER JOIN groups_access_levels gal ON gal.group_id = ug.group_id
		      INNER JOIN access_levels_modules alm ON alm.access_level_id = gal.access_level_id
		      INNER JOIN module m ON m.module_id = alm.module_id
		      WHERE m.name = ? AND alm.can_be_assigned = 1
		    )
		  )
		ORDER BY u.name ASC`, moduleName, moduleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanBasicUserRows(rows)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	_, err := s.queries.GetUserByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("usuário não encontrado")
	}
	if err != nil {
		return err
	}
	return s.queries.SoftDeleteUser(ctx, int32(id))
}

func (s *UserService) ResetPasswordAdmin(ctx context.Context, id int64, req ResetPasswordAdminRequest) error {
	if strings.TrimSpace(req.NewPassword) == "" {
		return errors.New("nova senha é obrigatória")
	}
	_, err := s.queries.GetUserByID(ctx, int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("usuário não encontrado")
	}
	if err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), saltRounds)
	if err != nil {
		return err
	}
	return s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Password: string(hashed), UserID: int32(id),
	})
}

func (s *UserService) ResetPasswordSelf(ctx context.Context, userID int64, req ResetPasswordSelfRequest) error {
	if strings.TrimSpace(req.CurrentPassword) == "" {
		return errors.New("senha atual é obrigatória")
	}
	if strings.TrimSpace(req.NewPassword) == "" {
		return errors.New("nova senha é obrigatória")
	}
	user, err := s.queries.GetUserByID(ctx, int32(userID))
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("usuário não encontrado")
	}
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return errors.New("senha atual incorreta")
	}
	if req.CurrentPassword == req.NewPassword {
		return errors.New("a nova senha não pode ser igual à senha atual")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), saltRounds)
	if err != nil {
		return err
	}
	return s.queries.UpdateUserPassword(ctx, db.UpdateUserPasswordParams{
		Password: string(hashed), UserID: int32(userID),
	})
}

func scanBasicUserRows(rows *sql.Rows) ([]UserResponse, error) {
	var result []UserResponse
	for rows.Next() {
		var userID int32
		var name string
		var lastName sql.NullString
		var email string
		var isActive bool
		if err := rows.Scan(&userID, &name, &lastName, &email, &isActive); err != nil {
			return nil, err
		}
		r := UserResponse{ID: int64(userID), Name: name, Email: email, IsActive: isActive}
		if lastName.Valid {
			r.Lastname = &lastName.String
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
