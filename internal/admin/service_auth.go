package admin

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"systemapi/internal/db"
	"systemapi/pkg/jwtutil"

	"golang.org/x/crypto/bcrypt"
)

var scopePriority = map[string]int{"ALL": 4, "GROUP": 3, "RESPONSIBLE": 2, "OWN": 1}

type AuthService struct {
	sqlDB   *sql.DB
	queries *db.Queries
}

func NewAuthService(sqlDB *sql.DB, q *db.Queries) *AuthService {
	return &AuthService{sqlDB: sqlDB, queries: q}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (AuthResponse, error) {
	if email == "" || password == "" {
		return AuthResponse{}, errors.New("email e senha são obrigatórios")
	}
	user, err := s.queries.GetUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return AuthResponse{}, errors.New("credenciais inválidas")
	}
	if err != nil {
		return AuthResponse{}, err
	}
	if !user.IsActive {
		return AuthResponse{}, errors.New("usuário não está ativo")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return AuthResponse{}, errors.New("credenciais inválidas")
	}

	token, err := jwtutil.Generate(int64(user.UserID), user.Name, user.Email)
	if err != nil {
		return AuthResponse{}, err
	}
	perms, err := s.buildPermissions(ctx, int64(user.UserID))
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: dbUserToResponse(user), Permissions: perms}, nil
}

func (s *AuthService) RefreshUser(ctx context.Context, userID int64) (AuthResponse, error) {
	user, err := s.queries.GetUserByID(ctx, int32(userID))
	if errors.Is(err, sql.ErrNoRows) {
		return AuthResponse{}, errors.New("usuário não encontrado")
	}
	if err != nil {
		return AuthResponse{}, err
	}
	if !user.IsActive {
		return AuthResponse{}, errors.New("usuário não está ativo")
	}
	token, err := jwtutil.Generate(int64(user.UserID), user.Name, user.Email)
	if err != nil {
		return AuthResponse{}, err
	}
	perms, err := s.buildPermissions(ctx, userID)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: dbUserToResponse(user), Permissions: perms}, nil
}

func (s *AuthService) buildPermissions(ctx context.Context, userID int64) (UserPermissionsDTO, error) {
	empty := UserPermissionsDTO{Permissions: []PermissionDTO{}, Menus: []MenuItemDTO{}}

	// 1. Access levels via UNION (diretos + via grupos)
	rows, err := s.sqlDB.QueryContext(ctx, `
		SELECT ual.access_level_id, NULL AS group_id
		FROM users_access_levels ual WHERE ual.user_id = ?
		UNION
		SELECT gal.access_level_id, ug.group_id
		FROM groups_access_levels gal
		INNER JOIN users_groups ug ON ug.group_id = gal.group_id
		WHERE ug.user_id = ?`, userID, userID)
	if err != nil {
		return empty, err
	}
	defer rows.Close()

	levelSet := map[int64]struct{}{}
	for rows.Next() {
		var alID int64
		var gID sql.NullInt64
		if err := rows.Scan(&alID, &gID); err != nil {
			return empty, err
		}
		levelSet[alID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return empty, err
	}
	if len(levelSet) == 0 {
		return empty, nil
	}

	levelIDs := make([]any, 0, len(levelSet))
	for id := range levelSet {
		levelIDs = append(levelIDs, id)
	}

	// 2. Permissões de módulos (OR bitmask + escopo mais permissivo)
	modRows, err := s.sqlDB.QueryContext(ctx,
		`SELECT m.module_id, m.name, alm.data_scope, alm.permission
		 FROM access_levels_modules alm
		 INNER JOIN module m ON m.module_id = alm.module_id
		 WHERE alm.access_level_id IN `+inPlaceholders(len(levelIDs)),
		levelIDs...)
	if err != nil {
		return empty, err
	}
	defer modRows.Close()

	type consRow struct {
		moduleID, permission int64
		moduleName, scope    string
	}
	consolidated := map[int64]*consRow{}

	for modRows.Next() {
		var moduleID, permission int64
		var moduleName, scope string
		if err := modRows.Scan(&moduleID, &moduleName, &scope, &permission); err != nil {
			return empty, err
		}
		if ex, ok := consolidated[moduleID]; ok {
			ex.permission |= permission
			if scopePriority[scope] > scopePriority[ex.scope] {
				ex.scope = scope
			}
		} else {
			consolidated[moduleID] = &consRow{moduleID, permission, moduleName, scope}
		}
	}
	if err := modRows.Err(); err != nil {
		return empty, err
	}
	if len(consolidated) == 0 {
		return empty, nil
	}

	// 3. Menus filtrados por scope
	moduleIDs := make([]any, 0, len(consolidated))
	for id := range consolidated {
		moduleIDs = append(moduleIDs, id)
	}
	menuRows, err := s.sqlDB.QueryContext(ctx,
		`SELECT menu_id, name, sort_index, paths, module_id, parent_id, required_scope
		 FROM menu WHERE module_id IN `+inPlaceholders(len(moduleIDs))+` ORDER BY sort_index ASC`,
		moduleIDs...)
	if err != nil {
		return empty, err
	}
	defer menuRows.Close()

	var menus []MenuItemDTO
	for menuRows.Next() {
		var (
			menuID, moduleID int64
			sortIndex        int32
			name, paths      string
			parentID         sql.NullInt64
			reqScope         sql.NullString
		)
		if err := menuRows.Scan(&menuID, &name, &sortIndex, &paths, &moduleID, &parentID, &reqScope); err != nil {
			return empty, err
		}
		if reqScope.Valid {
			mod := consolidated[moduleID]
			if mod == nil || scopePriority[mod.scope] < scopePriority[reqScope.String] {
				continue
			}
		}
		var pID *int64
		if parentID.Valid {
			pID = &parentID.Int64
		}
		menus = append(menus, MenuItemDTO{ID: menuID, Name: name, Paths: paths, SortIndex: sortIndex, ModuleID: moduleID, ParentID: pID})
	}
	if err := menuRows.Err(); err != nil {
		return empty, err
	}

	permissions := make([]PermissionDTO, 0, len(consolidated))
	for _, c := range consolidated {
		permissions = append(permissions, PermissionDTO{Module: c.moduleName, Permission: int(c.permission), DataScope: c.scope})
	}
	if menus == nil {
		menus = []MenuItemDTO{}
	}
	return UserPermissionsDTO{Permissions: permissions, Menus: menus}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func dbUserToResponse(u db.User) UserResponse {
	r := UserResponse{ID: int64(u.UserID), Name: u.Name, Email: u.Email, IsActive: u.IsActive}
	if u.LastName.Valid {
		r.Lastname = &u.LastName.String
	}
	return r
}

func inPlaceholders(n int) string {
	ph := make([]string, n)
	for i := range ph {
		ph[i] = "?"
	}
	return "(" + strings.Join(ph, ",") + ")"
}
