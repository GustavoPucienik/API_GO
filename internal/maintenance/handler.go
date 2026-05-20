package maintenance

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	mw "systemapi/internal/middleware"
	"systemapi/pkg/response"
)

type Handler struct {
	order               *OrderService
	lookups             *LookupsService
	attachment          *AttachmentService
	checklist           *ChecklistService
	checklistTemplate   *ChecklistTemplateService
	dashboard           *DashboardService
	config              *ConfigService
}

func parseID(r *http.Request) (int32, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}

func decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ── Order handlers ────────────────────────────────────────────────────────────

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	order, err := h.order.Create(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, order)
}

func (h *Handler) findOrdersPaginated(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	params := FindPaginatedParams{
		Page:      page,
		Limit:     limit,
		Search:    q.Get("search"),
		SortBy:    q.Get("sortBy"),
		SortOrder: q.Get("sortOrder"),
		Scope:     mw.GetScope(r),
	}
	if v := q.Get("statusId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			n32 := int32(n)
			params.StatusID = &n32
		}
	}
	if v := q.Get("priorityId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			n32 := int32(n)
			params.PriorityID = &n32
		}
	}
	if v := q.Get("reasonId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			n32 := int32(n)
			params.ReasonID = &n32
		}
	}
	if v := q.Get("clientCode"); v != "" {
		params.ClientCode = &v
	}
	if v := q.Get("responsibleId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			params.ResponsibleID = &n
		}
	} else if v := q.Get("assignedToId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			params.ResponsibleID = &n
		}
	}
	if v := q.Get("startDate"); v != "" {
		params.StartDate = &v
	}
	if v := q.Get("endDate"); v != "" {
		params.EndDate = &v
	}
	if q.Get("onlyWithScheduleDate") == "true" {
		params.OnlyWithScheduleDate = true
	}

	orders, total, err := h.order.FindPaginated(r.Context(), params)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	response.Page(w, orders, response.Meta{Total: total, Page: page, TotalPages: totalPages, Limit: limit})
}

func (h *Handler) findOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	order, err := h.order.FindByID(r.Context(), id, mw.GetScope(r))
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "não encontrada") || strings.Contains(err.Error(), "não encontrado") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "sem permissão") {
			status = http.StatusForbidden
		}
		response.Err(w, status, err.Error())
		return
	}
	response.Ok(w, order)
}

func (h *Handler) updateOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateOrderRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	order, err := h.order.Update(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, order)
}

func (h *Handler) assignOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req AssignRequest
	if err := decode(r, &req); err != nil || req.UserID == 0 {
		response.Err(w, http.StatusBadRequest, "userId não informado")
		return
	}
	order, err := h.order.Assign(r.Context(), id, req.UserID)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, order)
}

func (h *Handler) saveSignature(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req SignatureRequest
	if err := decode(r, &req); err != nil || req.ClientSignature == "" {
		response.Err(w, http.StatusBadRequest, "assinatura não informada")
		return
	}
	if err := h.order.SaveSignature(r.Context(), id, req.ClientSignature); err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Msg(w, "Assinatura salva com sucesso")
}

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req ChangeStatusRequest
	if err := decode(r, &req); err != nil || req.StatusID == 0 {
		response.Err(w, http.StatusBadRequest, "statusId não informado")
		return
	}
	order, err := h.order.ChangeStatus(r.Context(), id, req.StatusID)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, order)
}

func (h *Handler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.order.Delete(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Ordem de serviço excluída com sucesso!")
}

// ── Lookups handlers ──────────────────────────────────────────────────────────

func (h *Handler) getAllLookups(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.GetAll(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

// Status
func (h *Handler) listStatuses(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListStatuses(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createStatus(w http.ResponseWriter, r *http.Request) {
	var req CreateStatusRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.CreateStatus(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateStatusRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.UpdateStatus(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deleteStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.lookups.DeleteStatus(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Status excluído com sucesso")
}

// Priority
func (h *Handler) listPriorities(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListPriorities(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createPriority(w http.ResponseWriter, r *http.Request) {
	var req CreatePriorityRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.CreatePriority(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) updatePriority(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdatePriorityRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.UpdatePriority(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deletePriority(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.lookups.DeletePriority(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Prioridade excluída com sucesso")
}

// Reason
func (h *Handler) listReasons(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListReasons(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createReason(w http.ResponseWriter, r *http.Request) {
	var req CreateReasonRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.CreateReason(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) updateReason(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateReasonRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.UpdateReason(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deleteReason(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.lookups.DeleteReason(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Motivo excluído com sucesso")
}

// Note templates
func (h *Handler) listNoteTemplates(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListNoteTemplates(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createNoteTemplate(w http.ResponseWriter, r *http.Request) {
	var req CreateNoteTemplateRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.CreateNoteTemplate(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) updateNoteTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateNoteTemplateRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.UpdateNoteTemplate(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deleteNoteTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.lookups.DeleteNoteTemplate(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Modelo de nota excluído com sucesso")
}

// Time entry types
func (h *Handler) listTimeEntryTypes(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListTimeEntryTypes(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createTimeEntryType(w http.ResponseWriter, r *http.Request) {
	var req CreateTimeEntryTypeRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.CreateTimeEntryType(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) updateTimeEntryType(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateTimeEntryTypeRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.lookups.UpdateTimeEntryType(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deleteTimeEntryType(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.lookups.DeleteTimeEntryType(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Tipo de apontamento excluído com sucesso")
}

// ── Attachment handlers ───────────────────────────────────────────────────────

func (h *Handler) uploadAttachments(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		response.Err(w, http.StatusBadRequest, "falha ao processar upload")
		return
	}

	var saved []UploadedFile
	for _, headers := range r.MultipartForm.File {
		for _, fh := range headers {
			f, err := h.saveUploadedFile(fh)
			if err != nil {
				response.Err(w, http.StatusInternalServerError, "erro ao salvar arquivo: "+err.Error())
				return
			}
			saved = append(saved, f)
		}
	}

	atts, err := h.attachment.Upload(r.Context(), id, saved)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, atts)
}

func (h *Handler) saveUploadedFile(fh *multipart.FileHeader) (UploadedFile, error) {
	src, err := fh.Open()
	if err != nil {
		return UploadedFile{}, err
	}
	defer src.Close()

	uploadDir := h.attachment.uploadDir
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return UploadedFile{}, err
	}

	savedName := strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + fh.Filename
	destPath := filepath.Join(uploadDir, savedName)

	dst, err := os.Create(destPath)
	if err != nil {
		return UploadedFile{}, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return UploadedFile{}, err
	}

	return UploadedFile{OriginalName: fh.Filename, SavedPath: savedName, Size: size}, nil
}

func (h *Handler) listAttachments(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	data, err := h.attachment.FindByOrder(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deleteAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "attachmentId"), 10, 64)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "attachmentId inválido")
		return
	}
	if err := h.attachment.Delete(r.Context(), int32(id)); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Arquivo excluído com sucesso")
}

// ── Checklist handlers ────────────────────────────────────────────────────────

func (h *Handler) listChecklists(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	data, err := h.checklist.GetByOrderID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) getChecklistByID(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	checklistID64, err := strconv.ParseInt(chi.URLParam(r, "checklistId"), 10, 64)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "checklistId inválido")
		return
	}
	data, err := h.checklist.GetByID(r.Context(), orderID, int32(checklistID64))
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) submitChecklist(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req SubmitChecklistRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.checklist.Submit(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) deleteChecklist(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	checklistID64, err := strconv.ParseInt(chi.URLParam(r, "checklistId"), 10, 64)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "checklistId inválido")
		return
	}
	if err := h.checklist.Delete(r.Context(), orderID, int32(checklistID64)); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Checklist excluído com sucesso")
}

// ── Checklist template handlers ───────────────────────────────────────────────

func (h *Handler) listChecklistTemplates(w http.ResponseWriter, r *http.Request) {
	var reasonID *int32
	if v := r.URL.Query().Get("reasonId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			n32 := int32(n)
			reasonID = &n32
		}
	}
	data, err := h.checklistTemplate.GetAll(r.Context(), reasonID)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) getChecklistTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	data, err := h.checklistTemplate.GetByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createChecklistTemplate(w http.ResponseWriter, r *http.Request) {
	var req CreateChecklistTemplateRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.checklistTemplate.Create(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) updateChecklistTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateChecklistTemplateRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.checklistTemplate.Update(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deleteChecklistTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.checklistTemplate.Delete(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Template excluído com sucesso")
}

func (h *Handler) createChecklistItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req CreateChecklistItemRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.checklistTemplate.CreateItem(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) updateChecklistItem(w http.ResponseWriter, r *http.Request) {
	templateID, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	itemID64, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "itemId inválido")
		return
	}
	var req UpdateChecklistItemRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.checklistTemplate.UpdateItem(r.Context(), templateID, int32(itemID64), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deleteChecklistItem(w http.ResponseWriter, r *http.Request) {
	templateID, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	itemID64, err := strconv.ParseInt(chi.URLParam(r, "itemId"), 10, 64)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "itemId inválido")
		return
	}
	if err := h.checklistTemplate.DeleteItem(r.Context(), templateID, int32(itemID64)); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Item excluído com sucesso")
}

// ── Dashboard handlers ────────────────────────────────────────────────────────

func parseDashboardFilter(r *http.Request) DashboardFilter {
	q := r.URL.Query()
	f := DashboardFilter{}
	if v := q.Get("clientCode"); v != "" {
		f.ClientCode = &v
	}
	if v := q.Get("assignedToId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			f.AssignedToID = &n
		}
	}
	if v := q.Get("statusId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			n32 := int32(n)
			f.StatusID = &n32
		}
	}
	if v := q.Get("priorityId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			n32 := int32(n)
			f.PriorityID = &n32
		}
	}
	if v := q.Get("reasonId"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			n32 := int32(n)
			f.ReasonID = &n32
		}
	}
	if v := q.Get("dateFrom"); v != "" {
		f.DateFrom = &v
	}
	if v := q.Get("dateTo"); v != "" {
		f.DateTo = &v
	}
	if v := q.Get("city"); v != "" {
		f.City = &v
	}
	if v := q.Get("state"); v != "" {
		f.State = &v
	}
	return f
}

func (h *Handler) dashboardSummary(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetSummary(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardByStatus(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetByStatus(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardByPriority(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetByPriority(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardByReason(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetByReason(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardByMonth(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetByMonth(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardByClient(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetByClient(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardByRequester(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetByRequester(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardByCity(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetByCity(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardTimeEntries(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetTimeEntriesSummary(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardTrend(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetTrend(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) dashboardByTechnician(w http.ResponseWriter, r *http.Request) {
	data, err := h.dashboard.GetSummaryByTechnician(r.Context(), parseDashboardFilter(r))
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

// ── Config handlers ───────────────────────────────────────────────────────────

func (h *Handler) getMaintenanceConfig(w http.ResponseWriter, r *http.Request) {
	data, err := h.config.GetConfig(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) getAssignableUsers(w http.ResponseWriter, r *http.Request) {
	data, err := h.config.GetAssignableUsers(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) updateAssignableAccessLevels(w http.ResponseWriter, r *http.Request) {
	var req UpdateAssignableAccessLevelsRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo inválido")
		return
	}
	data, err := h.config.UpdateAssignableAccessLevels(r.Context(), req.AssignableAccessLevelIDs)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}
