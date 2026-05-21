package admin

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"

	"systemapi/internal/db"
	mw "systemapi/internal/middleware"
)

func NewHandler(sqlDB *sql.DB, q *db.Queries) *Handler {
	return &Handler{
		auth:        NewAuthService(sqlDB, q),
		user:        NewUserService(sqlDB, q),
		group:       NewGroupService(sqlDB, q),
		userGroup:   NewUserGroupService(sqlDB, q),
		accessLevel: NewAccessLevelService(sqlDB, q),
		module:      NewModuleService(sqlDB, q),
		menu:        NewMenuService(sqlDB, q),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router, sqlDB *sql.DB) {
	perm := func(mod string, bit int) func(http.Handler) http.Handler {
		return mw.Permission(sqlDB, mod, bit)
	}

	// ─── Auth ──────────────────────────────────────────────────────
	loginLimiter := httprate.LimitByIP(7, 15*time.Minute)

	r.With(loginLimiter).Post("/auth/login", h.login)
	r.With(mw.Auth).Post("/auth/refresh", h.refresh)
	r.Post("/auth/register", h.registerSelf)

	// ─── User ──────────────────────────────────────────────────────
	r.Route("/user", func(r chi.Router) {
		r.Use(mw.Auth)
		r.With(perm("user", mw.PermRead)).Get("/", h.findUsersPaginated)
		r.With(perm("user", mw.PermRead)).Get("/all", h.findAllUsers)
		r.With(perm("user", mw.PermRead)).Get("/by-module", h.findUsersByModule)
		r.With(perm("user", mw.PermCreate)).Post("/", h.createUser)
		r.Patch("/me/password", h.resetPasswordSelf)
		r.Route("/{id}", func(r chi.Router) {
			r.With(perm("user", mw.PermRead)).Get("/", h.findUserByID)
			r.With(perm("user", mw.PermRead)).Get("/edit-data", h.getUserWithRelations)
			r.With(perm("user", mw.PermUpdate)).Put("/", h.updateUser)
			r.With(perm("user", mw.PermDelete)).Delete("/", h.deleteUser)
			r.With(perm("user", mw.PermUpdate)).Patch("/password", h.resetPasswordAdmin)
		})
	})

	// ─── Group ─────────────────────────────────────────────────────
	r.Route("/group", func(r chi.Router) {
		r.Use(mw.Auth)
		r.With(perm("group", mw.PermRead)).Get("/", h.findGroupsPaginated)
		r.With(perm("group", mw.PermRead)).Get("/{id}", h.findGroupByID)
		r.With(perm("group", mw.PermCreate)).Post("/", h.createGroup)
		r.With(perm("group", mw.PermUpdate)).Put("/{id}", h.updateGroup)
		r.With(perm("group", mw.PermDelete)).Delete("/{id}", h.deleteGroup)
	})

	// ─── Access Level ───────────────────────────────────────────────
	r.Route("/access-level", func(r chi.Router) {
		r.Use(mw.Auth)
		r.With(perm("access-level", mw.PermRead)).Get("/", h.findAccessLevelsPaginated)
		r.With(perm("access-level", mw.PermRead)).Get("/{id}", h.findAccessLevelByID)
		r.With(perm("access-level", mw.PermCreate)).Post("/", h.createAccessLevel)
		r.With(perm("access-level", mw.PermUpdate)).Put("/{id}", h.updateAccessLevel)
		r.With(perm("access-level", mw.PermDelete)).Delete("/{id}", h.deleteAccessLevel)
	})

	// ─── Module ─────────────────────────────────────────────────────
	r.Route("/module", func(r chi.Router) {
		r.Use(mw.Auth)
		r.With(perm("module", mw.PermRead)).Get("/", h.findAllModules)
		r.With(perm("module", mw.PermRead)).Get("/{id}", h.findModuleByID)
		r.With(perm("module", mw.PermCreate)).Post("/", h.createModule)
		r.With(perm("module", mw.PermUpdate)).Put("/{id}", h.updateModule)
		r.With(perm("module", mw.PermDelete)).Delete("/{id}", h.deleteModule)
	})

	// ─── Menu ───────────────────────────────────────────────────────
	r.Route("/menu", func(r chi.Router) {
		r.Use(mw.Auth)
		r.With(perm("menu", mw.PermRead)).Get("/", h.findAllMenus)
		r.With(perm("menu", mw.PermCreate)).Post("/", h.createMenu)
		r.With(perm("menu", mw.PermRead)).Get("/{id}", h.findMenuByID)
		r.With(perm("menu", mw.PermUpdate)).Put("/{id}", h.updateMenu)
		r.With(perm("menu", mw.PermDelete)).Delete("/{id}", h.deleteMenu)
	})

	// ─── User Groups ─────────────────────────────────────────────────
	r.Route("/user-groups", func(r chi.Router) {
		r.Use(mw.Auth)
		r.With(perm("Administração", mw.PermRead)).Get("/group/{groupId}/users", h.listUsersByGroup)
		r.With(perm("Administração", mw.PermUpdate)).Put("/group/{groupId}/users", h.syncUsersInGroup)
	})
}
