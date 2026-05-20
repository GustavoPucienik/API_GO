package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"systemapi/pkg/response"
)

const (
	PermRead   = 1
	PermCreate = 2
	PermUpdate = 4
	PermDelete = 8
)

type DataScope string

const (
	ScopeOWN         DataScope = "OWN"
	ScopeRESPONSIBLE DataScope = "RESPONSIBLE"
	ScopeGROUP       DataScope = "GROUP"
	ScopeALL         DataScope = "ALL"
)

var scopePriority = map[DataScope]int{
	ScopeALL:         4,
	ScopeGROUP:       3,
	ScopeRESPONSIBLE: 2,
	ScopeOWN:         1,
}

type ResolvedScope struct {
	Type     DataScope
	UserID   int64
	GroupIDs []int64
}

const scopeCtxKey contextKey = "scope"

func Permission(db *sql.DB, moduleName string, requiredBit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUser(r)
			if user == nil {
				response.Err(w, http.StatusUnauthorized, "Não autenticado")
				return
			}

			// Busca todos os access_level_id do usuário (diretos + via grupos) em uma única query
			rows, err := db.QueryContext(r.Context(), `
				SELECT ual.access_level_id, NULL AS group_id
				FROM users_access_levels ual
				WHERE ual.user_id = ?
				UNION
				SELECT gal.access_level_id, ug.group_id
				FROM groups_access_levels gal
				INNER JOIN users_groups ug ON ug.group_id = gal.group_id
				WHERE ug.user_id = ?`, user.ID, user.ID)
			if err != nil {
				response.Err(w, http.StatusInternalServerError, "Erro ao verificar permissão")
				return
			}
			defer rows.Close()

			levelSet := map[int64]struct{}{}
			groupSet := map[int64]struct{}{}

			for rows.Next() {
				var alID int64
				var gID sql.NullInt64
				if err := rows.Scan(&alID, &gID); err != nil {
					continue
				}
				levelSet[alID] = struct{}{}
				if gID.Valid {
					groupSet[gID.Int64] = struct{}{}
				}
			}
			if err := rows.Err(); err != nil {
				response.Err(w, http.StatusInternalServerError, "Erro ao verificar permissão")
				return
			}

			if len(levelSet) == 0 {
				response.Err(w, http.StatusForbidden, "Sem permissão")
				return
			}

			// Busca permissões do módulo para todos os access levels com IN clause
			levelIDs := make([]any, 0, len(levelSet))
			for id := range levelSet {
				levelIDs = append(levelIDs, id)
			}
			query := `
				SELECT alm.permission, alm.data_scope
				FROM access_levels_modules alm
				INNER JOIN module m ON m.module_id = alm.module_id
				WHERE alm.access_level_id IN ` + inClause(len(levelIDs)) + ` AND m.name = ?`

			args := append(levelIDs, moduleName)
			modRows, err := db.QueryContext(r.Context(), query, args...)
			if err != nil {
				response.Err(w, http.StatusInternalServerError, "Erro ao verificar permissão")
				return
			}
			defer modRows.Close()

			combined := 0
			resolved := ScopeOWN

			for modRows.Next() {
				var perm int
				var scope DataScope
				if err := modRows.Scan(&perm, &scope); err != nil {
					continue
				}
				combined |= perm
				if scopePriority[scope] > scopePriority[resolved] {
					resolved = scope
				}
			}

			if (combined & requiredBit) != requiredBit {
				response.Err(w, http.StatusForbidden, "Sem permissão")
				return
			}

			groupIDs := []int64{}
			if resolved == ScopeGROUP {
				for id := range groupSet {
					groupIDs = append(groupIDs, id)
				}
			}

			ctx := context.WithValue(r.Context(), scopeCtxKey, &ResolvedScope{
				Type:     resolved,
				UserID:   user.ID,
				GroupIDs: groupIDs,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetScope(r *http.Request) *ResolvedScope {
	s, _ := r.Context().Value(scopeCtxKey).(*ResolvedScope)
	return s
}

func inClause(n int) string {
	placeholders := make([]string, n)
	for i := range placeholders {
		placeholders[i] = "?"
	}
	return "(" + strings.Join(placeholders, ",") + ")"
}
