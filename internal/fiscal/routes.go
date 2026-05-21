package fiscal

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
	}
}

func (h *Handler) RegisterRoutes(r chi.Router, sqlDB *sql.DB) {
	perm := func(mod string, bit int) func(http.Handler) http.Handler {
		return mw.Permission(sqlDB, mod, bit)
	}

	// ─── Orders ────────────────────────────────────────────────────────────────
	r.Route("/fiscal-order", func(r chi.Router) {
		r.Use(mw.Auth)

		r.With(perm("fiscal", mw.PermCreate)).Post("/", h.createOrder)
		r.With(perm("fiscal", mw.PermRead)).Get("/", h.findOrdersPaginated)
		r.With(perm("fiscal", mw.PermRead)).Get("/pending-approval", h.findPendingApproval)
		r.With(perm("fiscal", mw.PermRead)).Get("/{id}", h.findOrderByID)
		r.With(perm("fiscal", mw.PermUpdate)).Put("/{id}", h.updateOrder)
		r.With(perm("fiscal", mw.PermUpdate)).Patch("/{id}/assign", h.assignOrder)
		r.With(perm("fiscal", mw.PermDelete)).Delete("/{id}", h.deleteOrder)
		r.Get("/{id}/audit-trail", h.getAuditTrail)
		r.Post("/{id}/download-log", h.logDownload)

		// Attachments (nested under order)
		r.With(perm("fiscal", mw.PermCreate)).Post("/{id}/attachments", h.uploadAttachments)
		r.With(perm("fiscal", mw.PermRead)).Get("/{id}/attachments", h.listAttachments)

		// Attachments (direct by attachment ID)
		r.With(perm("fiscal", mw.PermRead)).Get("/attachments/{id}/download", h.downloadAttachment)
		r.With(perm("fiscal", mw.PermDelete)).Delete("/attachments/{id}", h.deleteAttachmentByID)
	})

	// ─── Lookups ───────────────────────────────────────────────────────────────
	r.Route("/fiscallookups", func(r chi.Router) {
		r.Use(mw.Auth)

		r.Get("/", h.getAllLookups)

		r.Route("/status", func(r chi.Router) {
			r.Get("/", h.listStatuses)
			r.With(perm("fiscal", mw.PermCreate)).Post("/", h.createStatus)
			r.Get("/{id}", h.getStatusByID)
			r.With(perm("fiscal", mw.PermUpdate)).Put("/{id}", h.updateStatus)
			r.With(perm("fiscal", mw.PermDelete)).Delete("/{id}", h.deleteStatus)
		})

		r.Route("/priorities", func(r chi.Router) {
			r.Get("/", h.listPriorities)
			r.With(perm("fiscal", mw.PermCreate)).Post("/", h.createPriority)
			r.Get("/{id}", h.getPriorityByID)
			r.With(perm("fiscal", mw.PermUpdate)).Put("/{id}", h.updatePriority)
			r.With(perm("fiscal", mw.PermDelete)).Delete("/{id}", h.deletePriority)
		})

		r.Route("/reasons", func(r chi.Router) {
			r.Get("/", h.listReasons)
			r.With(perm("fiscal", mw.PermCreate)).Post("/", h.createReason)
			r.Get("/{id}", h.getReasonByID)
			r.With(perm("fiscal", mw.PermUpdate)).Put("/{id}", h.updateReason)
			r.With(perm("fiscal", mw.PermDelete)).Delete("/{id}", h.deleteReason)
		})

		r.Route("/transitions", func(r chi.Router) {
			r.Get("/", h.listTransitions)
			r.With(perm("Configurações", mw.PermCreate)).Post("/", h.createTransition)
			r.With(perm("Configurações", mw.PermDelete)).Delete("/{id}", h.deleteTransition)
		})
	})
}
