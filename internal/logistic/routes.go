package logistic

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"systemapi/internal/db"
	mw "systemapi/internal/middleware"
)

const uploadDir = "uploads/orders"

func NewHandler(sqlDB *sql.DB, q *db.Queries) *Handler {
	return &Handler{
		lookups:    NewLookupsService(q),
		order:      NewOrderService(sqlDB, q),
		attachment: NewAttachmentService(sqlDB, q, uploadDir),
		config:     NewConfigService(sqlDB, q),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router, sqlDB *sql.DB) {
	perm := func(mod string, bit int) func(http.Handler) http.Handler {
		return mw.Permission(sqlDB, mod, bit)
	}

	// ─── Orders ────────────────────────────────────────────────────────────────
	r.Route("/logistic-order", func(r chi.Router) {
		r.Use(mw.Auth)

		r.With(perm("logistica", mw.PermCreate)).Post("/", h.createOrder)
		r.With(perm("logistica", mw.PermRead)).Get("/", h.findOrdersPaginated)
		r.With(perm("logistica", mw.PermRead)).Get("/pending-approval", h.findPendingApproval)
		r.With(perm("logistica", mw.PermRead)).Get("/{id}", h.findOrderByID)
		r.With(perm("logistica", mw.PermUpdate)).Put("/{id}", h.updateOrder)
		r.With(perm("logistica", mw.PermUpdate)).Patch("/{id}/assign", h.assignOrder)
		r.With(perm("logistica", mw.PermUpdate)).Patch("/{id}/status", h.changeStatus)
		r.With(perm("logistica", mw.PermDelete)).Delete("/{id}", h.deleteOrder)
		r.Get("/{id}/audit-trail", h.getAuditTrail)
		r.Post("/{id}/download-log", h.logDownload)

		// Attachments (nested under order)
		r.With(perm("logistica", mw.PermCreate)).Post("/{id}/attachments", h.uploadAttachments)
		r.With(perm("logistica", mw.PermRead)).Get("/{id}/attachments", h.listAttachments)

		// Attachments (direct by attachment ID)
		r.With(perm("logistica", mw.PermRead)).Get("/attachments/{id}/download", h.downloadAttachment)
		r.With(perm("logistica", mw.PermDelete)).Delete("/attachments/{id}", h.deleteAttachmentByID)
	})

	// ─── Config ────────────────────────────────────────────────────────────────
	r.Route("/logistic-config", func(r chi.Router) {
		r.Use(mw.Auth)

		r.Get("/", h.getConfig)
		r.Get("/assignable-users", h.getAssignableUsers)
		r.With(perm("Configurações", mw.PermUpdate)).Put("/salary-discount-access-level", h.updateSalaryDiscountAccessLevel)
		r.With(perm("Logística", mw.PermUpdate)).Put("/assignable-access-levels", h.updateAssignableAccessLevels)
	})

	// ─── Lookups ───────────────────────────────────────────────────────────────
	r.Route("/logisticlookups", func(r chi.Router) {
		r.Use(mw.Auth)

		r.Get("/", h.getAllLookups)

		r.Route("/status", func(r chi.Router) {
			r.Get("/", h.listStatuses)
			r.With(perm("logistica", mw.PermCreate)).Post("/", h.createStatus)
			r.Get("/{id}", h.getStatusByID)
			r.With(perm("logistica", mw.PermUpdate)).Put("/{id}", h.updateStatus)
			r.With(perm("logistica", mw.PermDelete)).Delete("/{id}", h.deleteStatus)
		})

		r.Route("/priorities", func(r chi.Router) {
			r.Get("/", h.listPriorities)
			r.With(perm("logistica", mw.PermCreate)).Post("/", h.createPriority)
			r.Get("/{id}", h.getPriorityByID)
			r.With(perm("logistica", mw.PermUpdate)).Put("/{id}", h.updatePriority)
			r.With(perm("logistica", mw.PermDelete)).Delete("/{id}", h.deletePriority)
		})

		r.Route("/reasons", func(r chi.Router) {
			r.Get("/", h.listReasons)
			r.With(perm("logistica", mw.PermCreate)).Post("/", h.createReason)
			r.Get("/{id}", h.getReasonByID)
			r.With(perm("logistica", mw.PermUpdate)).Put("/{id}", h.updateReason)
			r.With(perm("logistica", mw.PermDelete)).Delete("/{id}", h.deleteReason)
		})

		r.Route("/transitions", func(r chi.Router) {
			r.Get("/", h.listTransitions)
			r.With(perm("Configurações", mw.PermCreate)).Post("/", h.createTransition)
			r.With(perm("Configurações", mw.PermDelete)).Delete("/{id}", h.deleteTransition)
		})

		r.Route("/operators", func(r chi.Router) {
			r.Get("/", h.listOperators)
			r.Post("/", h.createOperator)
			r.Get("/{id}", h.getOperatorByID)
			r.With(perm("logistica", mw.PermUpdate)).Put("/{id}", h.updateOperator)
			r.With(perm("logistica", mw.PermDelete)).Delete("/{id}", h.deleteOperator)
		})
	})
}
