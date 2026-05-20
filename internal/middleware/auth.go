package middleware

import (
	"context"
	"net/http"
	"strings"

	"systemapi/pkg/jwtutil"
	"systemapi/pkg/response"
)

type contextKey string

const userCtxKey contextKey = "user"

type AuthUser struct {
	ID    int64
	Name  string
	Email string
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			response.Err(w, http.StatusUnauthorized, "Token não informado")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Err(w, http.StatusUnauthorized, "Token inválido")
			return
		}
		claims, err := jwtutil.Parse(parts[1])
		if err != nil {
			response.Err(w, http.StatusUnauthorized, "Token inválido")
			return
		}
		ctx := context.WithValue(r.Context(), userCtxKey, &AuthUser{
			ID:    claims.ID,
			Name:  claims.Name,
			Email: claims.Email,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(r *http.Request) *AuthUser {
	u, _ := r.Context().Value(userCtxKey).(*AuthUser)
	return u
}
