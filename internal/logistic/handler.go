package logistic

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	mw "systemapi/internal/middleware"
	"systemapi/pkg/response"
)

type Handler struct {
	lookups    *LookupsService
	order      *OrderService
	attachment *AttachmentService
	config     *ConfigService
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

// ── Lookup handlers ───────────────────────────────────────────────────────────

func (h *Handler) getAllLookups(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.GetAll(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) listStatuses(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListStatuses(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) getStatusByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	data, err := h.lookups.GetStatusByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createStatus(w http.ResponseWriter, r *http.Request) {
	var req CreateStatusRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
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
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
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

func (h *Handler) listPriorities(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListPriorities(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) getPriorityByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	data, err := h.lookups.GetPriorityByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createPriority(w http.ResponseWriter, r *http.Request) {
	var req CreatePriorityRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
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
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
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

func (h *Handler) listReasons(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListReasons(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) getReasonByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	data, err := h.lookups.GetReasonByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createReason(w http.ResponseWriter, r *http.Request) {
	var req CreateReasonRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
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
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
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

func (h *Handler) listTransitions(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListTransitions(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createTransition(w http.ResponseWriter, r *http.Request) {
	var req CreateTransitionRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	data, err := h.lookups.CreateTransition(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) deleteTransition(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.lookups.DeleteTransition(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Transição excluída com sucesso")
}

func (h *Handler) listOperators(w http.ResponseWriter, r *http.Request) {
	data, err := h.lookups.ListOperators(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) getOperatorByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	data, err := h.lookups.GetOperatorByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) createOperator(w http.ResponseWriter, r *http.Request) {
	var req CreateOperatorRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	data, err := h.lookups.CreateOperator(r.Context(), req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, data)
}

func (h *Handler) updateOperator(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req UpdateOperatorRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	data, err := h.lookups.UpdateOperator(r.Context(), id, req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) deleteOperator(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.lookups.DeleteOperator(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Operador excluído com sucesso")
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
	statusID := parseOptionalInt32(q.Get("statusId"))
	priorityID := parseOptionalInt32(q.Get("priorityId"))
	reasonID := parseOptionalInt32(q.Get("reasonId"))
	assignedToID := parseOptionalInt32(q.Get("assignedToId"))
	startDate := q.Get("startDate")
	endDate := q.Get("endDate")
	driverName := q.Get("driverName")
	onlyWithScheduleDate := q.Get("onlyWithScheduleDate") == "true"

	p := FindPaginatedParams{
		Page:                 page,
		Limit:                limit,
		Search:               q.Get("search"),
		SortBy:               q.Get("sortBy"),
		SortOrder:            q.Get("sortOrder"),
		StatusID:             statusID,
		PriorityID:           priorityID,
		ReasonID:             reasonID,
		AssignedToID:         assignedToID,
		OnlyWithScheduleDate: onlyWithScheduleDate,
	}
	if startDate != "" {
		p.StartDate = &startDate
	}
	if endDate != "" {
		p.EndDate = &endDate
	}
	if driverName != "" {
		p.DriverName = &driverName
	}

	data, total, err := h.order.FindPaginated(r.Context(), p)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalPages := int(total) / p.Limit
	if p.Limit > 0 && int(total)%p.Limit != 0 {
		totalPages++
	}
	response.Ok(w, PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       p.Page,
		TotalPages: totalPages,
		Limit:      p.Limit,
	})
}

func (h *Handler) findPendingApproval(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	p := FindPaginatedParams{
		Page:      page,
		Limit:     limit,
		Search:    q.Get("search"),
		SortOrder: q.Get("sortOrder"),
	}
	data, total, err := h.order.FindPendingApproval(r.Context(), p)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	totalPages := int(total) / p.Limit
	if int(total)%p.Limit != 0 {
		totalPages++
	}
	response.Ok(w, PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       p.Page,
		TotalPages: totalPages,
		Limit:      p.Limit,
	})
}

func (h *Handler) findOrderByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	order, err := h.order.FindByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
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

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req ChangeStatusRequest
	if err := decode(r, &req); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	user := mw.GetUser(r)
	if user == nil {
		response.Err(w, http.StatusUnauthorized, "não autenticado")
		return
	}
	order, err := h.order.ChangeStatus(r.Context(), id, user.ID, req.StatusID)
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
	response.Msg(w, "Ordem logística excluída com sucesso")
}

func (h *Handler) getAuditTrail(w http.ResponseWriter, _ *http.Request) {
	response.Ok(w, []any{})
}

func (h *Handler) logDownload(w http.ResponseWriter, _ *http.Request) {
	response.Msg(w, "Log registrado")
}

// ── Attachment handlers ───────────────────────────────────────────────────────

func (h *Handler) uploadAttachments(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		response.Err(w, http.StatusBadRequest, "falha ao processar multipart")
		return
	}
	files, err := saveUploadedFiles(r, uploadDir)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	atts, err := h.attachment.Upload(r.Context(), id, files)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Created(w, atts)
}

func (h *Handler) listAttachments(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	atts, err := h.attachment.FindByOrder(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Ok(w, atts)
}

func (h *Handler) downloadAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	att, err := h.attachment.GetByID(r.Context(), id)
	if err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	fullPath := filepath.Join(uploadDir, att.FilePath)
	f, err := os.Open(fullPath)
	if err != nil {
		response.Err(w, http.StatusNotFound, "arquivo não encontrado no servidor")
		return
	}
	defer f.Close()
	w.Header().Set("Content-Disposition", "attachment; filename=\""+att.FileName+"\"")
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, f)
}

func (h *Handler) deleteAttachmentByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.attachment.Delete(r.Context(), id); err != nil {
		response.Err(w, http.StatusNotFound, err.Error())
		return
	}
	response.Msg(w, "Arquivo excluído com sucesso")
}

// ── Config handlers ───────────────────────────────────────────────────────────

func (h *Handler) getConfig(w http.ResponseWriter, r *http.Request) {
	data, err := h.config.GetConfig(r.Context())
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) updateSalaryDiscountAccessLevel(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SalaryDiscountAccessLevelID *int32 `json:"salaryDiscountAccessLevelId"`
	}
	if err := decode(r, &body); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	data, err := h.config.UpdateSalaryDiscount(r.Context(), body.SalaryDiscountAccessLevelID)
	if err != nil {
		response.Err(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.Ok(w, data)
}

func (h *Handler) updateAssignableAccessLevels(w http.ResponseWriter, r *http.Request) {
	var body UpdateConfigRequest
	if err := decode(r, &body); err != nil {
		response.Err(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	data, err := h.config.UpdateAssignableAccessLevels(r.Context(), body.AssignableAccessLevelIDs)
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

// ── Helpers ───────────────────────────────────────────────────────────────────

func parseOptionalInt32(s string) *int32 {
	if s == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return nil
	}
	v := int32(n)
	return &v
}

func saveUploadedFiles(r *http.Request, dir string) ([]UploadedFile, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	var files []UploadedFile
	for _, headers := range r.MultipartForm.File {
		for _, header := range headers {
			f, err := saveMultipartFile(header, dir)
			if err != nil {
				return nil, err
			}
			files = append(files, f)
		}
	}
	return files, nil
}

func saveMultipartFile(header *multipart.FileHeader, dir string) (UploadedFile, error) {
	src, err := header.Open()
	if err != nil {
		return UploadedFile{}, err
	}
	defer src.Close()

	ext := filepath.Ext(header.Filename)
	savedName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst, err := os.Create(filepath.Join(dir, savedName))
	if err != nil {
		return UploadedFile{}, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return UploadedFile{}, err
	}

	return UploadedFile{
		OriginalName: header.Filename,
		SavedPath:    savedName,
		Size:         size,
	}, nil
}
