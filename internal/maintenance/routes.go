package maintenance

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
		order:             NewOrderService(sqlDB, q),
		lookups:           NewLookupsService(q),
		attachment:        NewAttachmentService(sqlDB, q, uploadDir),
		checklist:         NewChecklistService(sqlDB, q),
		checklistTemplate: NewChecklistTemplateService(q),
		dashboard:         NewDashboardService(sqlDB),
		config:            NewConfigService(sqlDB, q),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router, sqlDB *sql.DB) {
	perm := func(mod string, bit int) func(http.Handler) http.Handler {
		return mw.Permission(sqlDB, mod, bit)
	}

	// ─── Orders ────────────────────────────────────────────────────────────────
	r.Route("/maintenance", func(r chi.Router) {
		r.Use(mw.Auth)

		r.With(perm("Manutenção", mw.PermCreate)).Post("/", h.createOrder)
		r.With(perm("Manutenção", mw.PermRead)).Get("/", h.findOrdersPaginated)
		r.With(perm("Manutenção", mw.PermRead)).Get("/{id}", h.findOrderByID)
		r.With(perm("Manutenção", mw.PermUpdate)).Put("/{id}", h.updateOrder)
		r.With(perm("Manutenção", mw.PermUpdate)).Patch("/{id}/assign", h.assignOrder)
		r.With(perm("Manutenção", mw.PermUpdate)).Patch("/{id}/signature", h.saveSignature)
		r.With(perm("Manutenção", mw.PermUpdate)).Patch("/{id}/status", h.changeStatus)
		r.With(perm("Manutenção", mw.PermDelete)).Delete("/{id}", h.deleteOrder)
		r.With(perm("Manutenção", mw.PermRead)).Get("/{id}/audit-trail", h.getAuditTrail)
		r.With(perm("Manutenção", mw.PermRead)).Post("/{id}/download-log", h.logDownload)

		// Attachments (nested under order)
		r.With(perm("Manutenção", mw.PermCreate)).Post("/{id}/attachments", h.uploadAttachments)
		r.With(perm("Manutenção", mw.PermRead)).Get("/{id}/attachments", h.listAttachments)
		r.With(perm("Manutenção", mw.PermDelete)).Delete("/{id}/attachments/{attachmentId}", h.deleteAttachment)

		// Attachments (direct by attachment ID)
		r.With(perm("Manutenção", mw.PermRead)).Get("/attachments/{id}/download", h.downloadAttachment)
		r.With(perm("Manutenção", mw.PermDelete)).Delete("/attachments/{id}", h.deleteAttachmentByID)

		// Checklists
		r.With(perm("Manutenção", mw.PermRead)).Get("/{id}/checklists", h.listChecklists)
		r.With(perm("Manutenção", mw.PermRead)).Get("/{id}/checklists/{checklistId}", h.getChecklistByID)
		r.With(perm("Manutenção", mw.PermCreate)).Post("/{id}/checklists", h.submitChecklist)
		r.With(perm("Manutenção", mw.PermDelete)).Delete("/{id}/checklists/{checklistId}", h.deleteChecklist)
	})

	// ─── Lookups ───────────────────────────────────────────────────────────────
	r.Route("/maintenance-lookups", func(r chi.Router) {
		r.Use(mw.Auth)

		r.Get("/", h.getAllLookups)

		// Status
		r.Route("/status", func(r chi.Router) {
			r.Get("/", h.listStatuses)
			r.With(perm("Manutenção", mw.PermCreate)).Post("/", h.createStatus)
			r.Get("/{id}", h.getStatusByID)
			r.With(perm("Manutenção", mw.PermUpdate)).Put("/{id}", h.updateStatus)
			r.With(perm("Manutenção", mw.PermDelete)).Delete("/{id}", h.deleteStatus)
		})

		// Priority
		r.Route("/priorities", func(r chi.Router) {
			r.Get("/", h.listPriorities)
			r.With(perm("Manutenção", mw.PermCreate)).Post("/", h.createPriority)
			r.Get("/{id}", h.getPriorityByID)
			r.With(perm("Manutenção", mw.PermUpdate)).Put("/{id}", h.updatePriority)
			r.With(perm("Manutenção", mw.PermDelete)).Delete("/{id}", h.deletePriority)
		})

		// Reason
		r.Route("/reasons", func(r chi.Router) {
			r.Get("/", h.listReasons)
			r.With(perm("Manutenção", mw.PermCreate)).Post("/", h.createReason)
			r.Get("/{id}", h.getReasonByID)
			r.With(perm("Manutenção", mw.PermUpdate)).Put("/{id}", h.updateReason)
			r.With(perm("Manutenção", mw.PermDelete)).Delete("/{id}", h.deleteReason)
		})

		// Note template
		r.Route("/note-templates", func(r chi.Router) {
			r.Get("/", h.listNoteTemplates)
			r.With(perm("Manutenção", mw.PermCreate)).Post("/", h.createNoteTemplate)
			r.Get("/{id}", h.getNoteTemplateByID)
			r.With(perm("Manutenção", mw.PermUpdate)).Put("/{id}", h.updateNoteTemplate)
			r.With(perm("Manutenção", mw.PermDelete)).Delete("/{id}", h.deleteNoteTemplate)
		})

		// Time entry type
		r.Route("/time-entry-types", func(r chi.Router) {
			r.Get("/", h.listTimeEntryTypes)
			r.With(perm("Manutenção", mw.PermCreate)).Post("/", h.createTimeEntryType)
			r.Get("/{id}", h.getTimeEntryTypeByID)
			r.With(perm("Manutenção", mw.PermUpdate)).Put("/{id}", h.updateTimeEntryType)
			r.With(perm("Manutenção", mw.PermDelete)).Delete("/{id}", h.deleteTimeEntryType)
		})
	})

	// ─── Checklist templates ───────────────────────────────────────────────────
	r.Route("/maintenance-checklist-template", func(r chi.Router) {
		r.Use(mw.Auth)

		r.Get("/", h.listChecklistTemplates)
		r.With(perm("Manutenção", mw.PermCreate)).Post("/", h.createChecklistTemplate)
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.getChecklistTemplate)
			r.With(perm("Manutenção", mw.PermUpdate)).Put("/", h.updateChecklistTemplate)
			r.With(perm("Manutenção", mw.PermDelete)).Delete("/", h.deleteChecklistTemplate)
			r.With(perm("Manutenção", mw.PermCreate)).Post("/items", h.createChecklistItem)
			r.With(perm("Manutenção", mw.PermUpdate)).Put("/items/{itemId}", h.updateChecklistItem)
			r.With(perm("Manutenção", mw.PermDelete)).Delete("/items/{itemId}", h.deleteChecklistItem)
		})
	})

	// ─── Dashboard ─────────────────────────────────────────────────────────────
	r.Route("/maintenance-dashboard", func(r chi.Router) {
		r.Use(mw.Auth)
		r.Use(perm("Manutenção", mw.PermRead))

		r.Get("/summary", h.dashboardSummary)
		r.Get("/by-status", h.dashboardByStatus)
		r.Get("/by-priority", h.dashboardByPriority)
		r.Get("/by-reason", h.dashboardByReason)
		r.Get("/by-month", h.dashboardByMonth)
		r.Get("/by-client", h.dashboardByClient)
		r.Get("/by-requester", h.dashboardByRequester)
		r.Get("/by-city", h.dashboardByCity)
		r.Get("/time-entries", h.dashboardTimeEntries)
		r.Get("/trend", h.dashboardTrend)
		r.Get("/by-technician", h.dashboardByTechnician)
	})

	// ─── Config ────────────────────────────────────────────────────────────────
	r.Route("/maintenance-config", func(r chi.Router) {
		r.Use(mw.Auth)
		r.Get("/", h.getMaintenanceConfig)
		r.Get("/assignable-users", h.getAssignableUsers)
		r.With(perm("Manutenção", mw.PermUpdate)).Put("/assignable-access-levels", h.updateAssignableAccessLevels)
	})
}
