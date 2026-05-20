package maintenance

import (
	"context"
	"database/sql"
	"errors"

	"systemapi/internal/db"
)

type LookupsService struct {
	q *db.Queries
}

func NewLookupsService(q *db.Queries) *LookupsService {
	return &LookupsService{q: q}
}

func (s *LookupsService) GetAll(ctx context.Context) (*LookupsResponse, error) {
	statuses, err := s.q.ListMaintenanceOrderStatuses(ctx)
	if err != nil {
		return nil, err
	}
	priorities, err := s.q.ListMaintenanceOrderPriorities(ctx)
	if err != nil {
		return nil, err
	}
	reasons, err := s.q.ListMaintenanceOrderReasons(ctx)
	if err != nil {
		return nil, err
	}
	noteTemplates, err := s.q.ListMaintenanceOrderNoteTemplates(ctx)
	if err != nil {
		return nil, err
	}
	timeEntryTypes, err := s.q.ListMaintenanceOrderTimeEntryTypes(ctx)
	if err != nil {
		return nil, err
	}

	resp := &LookupsResponse{
		Status:         make([]StatusResponse, len(statuses)),
		Priorities:     make([]PriorityResponse, len(priorities)),
		Reasons:        make([]ReasonResponse, len(reasons)),
		NoteTemplates:  make([]NoteTemplateResponse, len(noteTemplates)),
		TimeEntryTypes: make([]TimeEntryTypeResponse, len(timeEntryTypes)),
	}
	for i, s := range statuses {
		resp.Status[i] = statusToResp(s)
	}
	for i, p := range priorities {
		resp.Priorities[i] = priorityToResp(p)
	}
	for i, r := range reasons {
		resp.Reasons[i] = ReasonResponse{ID: r.MaintenanceOrderReasonID, Name: r.Name}
	}
	for i, t := range noteTemplates {
		resp.NoteTemplates[i] = NoteTemplateResponse{ID: t.MaintenanceOrderNoteTemplateID, Name: t.Name, Description: t.Description, CreatedAt: t.CreatedAt}
	}
	for i, t := range timeEntryTypes {
		resp.TimeEntryTypes[i] = timeEntryTypeToResp(t)
	}
	return resp, nil
}

// ── Status ────────────────────────────────────────────────────────────────────

func (s *LookupsService) ListStatuses(ctx context.Context) ([]StatusResponse, error) {
	rows, err := s.q.ListMaintenanceOrderStatuses(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]StatusResponse, len(rows))
	for i, r := range rows {
		out[i] = statusToResp(r)
	}
	return out, nil
}

func (s *LookupsService) GetStatusByID(ctx context.Context, id int32) (*StatusResponse, error) {
	row, err := s.q.GetMaintenanceOrderStatusByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("status não encontrado")
	}
	if err != nil {
		return nil, err
	}
	r := statusToResp(row)
	return &r, nil
}

func (s *LookupsService) CreateStatus(ctx context.Context, req CreateStatusRequest) (*StatusResponse, error) {
	res, err := s.q.CreateMaintenanceOrderStatus(ctx, db.CreateMaintenanceOrderStatusParams{
		Name:            req.Name,
		BackgroundColor: req.BackgroundColor,
		BorderColor:     req.BorderColor,
		TextColor:       req.TextColor,
		SortIndex:       req.SortIndex,
		IsClosed:        req.IsClosed,
		IsDefault:       req.IsDefault,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetStatusByID(ctx, int32(id))
}

func (s *LookupsService) UpdateStatus(ctx context.Context, id int32, req UpdateStatusRequest) (*StatusResponse, error) {
	if _, err := s.GetStatusByID(ctx, id); err != nil {
		return nil, err
	}
	err := s.q.UpdateMaintenanceOrderStatus(ctx, db.UpdateMaintenanceOrderStatusParams{
		Name:                     req.Name,
		BackgroundColor:          req.BackgroundColor,
		BorderColor:              req.BorderColor,
		TextColor:                req.TextColor,
		SortIndex:                req.SortIndex,
		IsClosed:                 req.IsClosed,
		IsDefault:                req.IsDefault,
		MaintenanceOrderStatusID: id,
	})
	if err != nil {
		return nil, err
	}
	return s.GetStatusByID(ctx, id)
}

func (s *LookupsService) DeleteStatus(ctx context.Context, id int32) error {
	if _, err := s.GetStatusByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteMaintenanceOrderStatus(ctx, id)
}

// ── Priority ──────────────────────────────────────────────────────────────────

func (s *LookupsService) ListPriorities(ctx context.Context) ([]PriorityResponse, error) {
	rows, err := s.q.ListMaintenanceOrderPriorities(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]PriorityResponse, len(rows))
	for i, r := range rows {
		out[i] = priorityToResp(r)
	}
	return out, nil
}

func (s *LookupsService) GetPriorityByID(ctx context.Context, id int32) (*PriorityResponse, error) {
	row, err := s.q.GetMaintenanceOrderPriorityByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("prioridade não encontrada")
	}
	if err != nil {
		return nil, err
	}
	r := priorityToResp(row)
	return &r, nil
}

func (s *LookupsService) CreatePriority(ctx context.Context, req CreatePriorityRequest) (*PriorityResponse, error) {
	res, err := s.q.CreateMaintenanceOrderPriority(ctx, db.CreateMaintenanceOrderPriorityParams{
		Name:            req.Name,
		BackgroundColor: req.BackgroundColor,
		BorderColor:     req.BorderColor,
		TextColor:       req.TextColor,
		SortIndex:       req.SortIndex,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetPriorityByID(ctx, int32(id))
}

func (s *LookupsService) UpdatePriority(ctx context.Context, id int32, req UpdatePriorityRequest) (*PriorityResponse, error) {
	if _, err := s.GetPriorityByID(ctx, id); err != nil {
		return nil, err
	}
	err := s.q.UpdateMaintenanceOrderPriority(ctx, db.UpdateMaintenanceOrderPriorityParams{
		Name:                         req.Name,
		BackgroundColor:              req.BackgroundColor,
		BorderColor:                  req.BorderColor,
		TextColor:                    req.TextColor,
		SortIndex:                    req.SortIndex,
		MaintenanceOrderPrioritiesID: id,
	})
	if err != nil {
		return nil, err
	}
	return s.GetPriorityByID(ctx, id)
}

func (s *LookupsService) DeletePriority(ctx context.Context, id int32) error {
	if _, err := s.GetPriorityByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteMaintenanceOrderPriority(ctx, id)
}

// ── Reason ────────────────────────────────────────────────────────────────────

func (s *LookupsService) ListReasons(ctx context.Context) ([]ReasonResponse, error) {
	rows, err := s.q.ListMaintenanceOrderReasons(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ReasonResponse, len(rows))
	for i, r := range rows {
		out[i] = ReasonResponse{ID: r.MaintenanceOrderReasonID, Name: r.Name}
	}
	return out, nil
}

func (s *LookupsService) GetReasonByID(ctx context.Context, id int32) (*ReasonResponse, error) {
	row, err := s.q.GetMaintenanceOrderReasonByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("motivo não encontrado")
	}
	if err != nil {
		return nil, err
	}
	r := ReasonResponse{ID: row.MaintenanceOrderReasonID, Name: row.Name}
	return &r, nil
}

func (s *LookupsService) CreateReason(ctx context.Context, req CreateReasonRequest) (*ReasonResponse, error) {
	res, err := s.q.CreateMaintenanceOrderReason(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetReasonByID(ctx, int32(id))
}

func (s *LookupsService) UpdateReason(ctx context.Context, id int32, req UpdateReasonRequest) (*ReasonResponse, error) {
	if _, err := s.GetReasonByID(ctx, id); err != nil {
		return nil, err
	}
	err := s.q.UpdateMaintenanceOrderReason(ctx, db.UpdateMaintenanceOrderReasonParams{
		Name:                     req.Name,
		MaintenanceOrderReasonID: id,
	})
	if err != nil {
		return nil, err
	}
	return s.GetReasonByID(ctx, id)
}

func (s *LookupsService) DeleteReason(ctx context.Context, id int32) error {
	if _, err := s.GetReasonByID(ctx, id); err != nil {
		return err
	}
	return s.q.SoftDeleteMaintenanceOrderReason(ctx, id)
}

// ── Note template ─────────────────────────────────────────────────────────────

func (s *LookupsService) ListNoteTemplates(ctx context.Context) ([]NoteTemplateResponse, error) {
	rows, err := s.q.ListMaintenanceOrderNoteTemplates(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]NoteTemplateResponse, len(rows))
	for i, r := range rows {
		out[i] = NoteTemplateResponse{ID: r.MaintenanceOrderNoteTemplateID, Name: r.Name, Description: r.Description, CreatedAt: r.CreatedAt}
	}
	return out, nil
}

func (s *LookupsService) GetNoteTemplateByID(ctx context.Context, id int32) (*NoteTemplateResponse, error) {
	row, err := s.q.GetMaintenanceOrderNoteTemplateByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("modelo de nota não encontrado")
	}
	if err != nil {
		return nil, err
	}
	r := NoteTemplateResponse{ID: row.MaintenanceOrderNoteTemplateID, Name: row.Name, Description: row.Description, CreatedAt: row.CreatedAt}
	return &r, nil
}

func (s *LookupsService) CreateNoteTemplate(ctx context.Context, req CreateNoteTemplateRequest) (*NoteTemplateResponse, error) {
	res, err := s.q.CreateMaintenanceOrderNoteTemplate(ctx, db.CreateMaintenanceOrderNoteTemplateParams{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetNoteTemplateByID(ctx, int32(id))
}

func (s *LookupsService) UpdateNoteTemplate(ctx context.Context, id int32, req UpdateNoteTemplateRequest) (*NoteTemplateResponse, error) {
	if _, err := s.GetNoteTemplateByID(ctx, id); err != nil {
		return nil, err
	}
	err := s.q.UpdateMaintenanceOrderNoteTemplate(ctx, db.UpdateMaintenanceOrderNoteTemplateParams{
		Name:                          req.Name,
		Description:                   req.Description,
		MaintenanceOrderNoteTemplateID: id,
	})
	if err != nil {
		return nil, err
	}
	return s.GetNoteTemplateByID(ctx, id)
}

func (s *LookupsService) DeleteNoteTemplate(ctx context.Context, id int32) error {
	if _, err := s.GetNoteTemplateByID(ctx, id); err != nil {
		return err
	}
	return s.q.SoftDeleteMaintenanceOrderNoteTemplate(ctx, id)
}

// ── Time entry type ───────────────────────────────────────────────────────────

func (s *LookupsService) ListTimeEntryTypes(ctx context.Context) ([]TimeEntryTypeResponse, error) {
	rows, err := s.q.ListMaintenanceOrderTimeEntryTypes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TimeEntryTypeResponse, len(rows))
	for i, r := range rows {
		out[i] = timeEntryTypeToResp(r)
	}
	return out, nil
}

func (s *LookupsService) GetTimeEntryTypeByID(ctx context.Context, id int32) (*TimeEntryTypeResponse, error) {
	row, err := s.q.GetMaintenanceOrderTimeEntryTypeByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("tipo de apontamento não encontrado")
	}
	if err != nil {
		return nil, err
	}
	r := timeEntryTypeToResp(row)
	return &r, nil
}

func (s *LookupsService) CreateTimeEntryType(ctx context.Context, req CreateTimeEntryTypeRequest) (*TimeEntryTypeResponse, error) {
	color := sql.NullString{}
	if req.Color != nil {
		color = sql.NullString{String: *req.Color, Valid: true}
	}
	res, err := s.q.CreateMaintenanceOrderTimeEntryType(ctx, db.CreateMaintenanceOrderTimeEntryTypeParams{
		Name:  req.Name,
		Color: color,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetTimeEntryTypeByID(ctx, int32(id))
}

func (s *LookupsService) UpdateTimeEntryType(ctx context.Context, id int32, req UpdateTimeEntryTypeRequest) (*TimeEntryTypeResponse, error) {
	if _, err := s.GetTimeEntryTypeByID(ctx, id); err != nil {
		return nil, err
	}
	color := sql.NullString{}
	if req.Color != nil {
		color = sql.NullString{String: *req.Color, Valid: true}
	}
	err := s.q.UpdateMaintenanceOrderTimeEntryType(ctx, db.UpdateMaintenanceOrderTimeEntryTypeParams{
		Name:                            req.Name,
		Color:                           color,
		MaintenanceOrderTimeEntryTypeID: id,
	})
	if err != nil {
		return nil, err
	}
	return s.GetTimeEntryTypeByID(ctx, id)
}

func (s *LookupsService) DeleteTimeEntryType(ctx context.Context, id int32) error {
	if _, err := s.GetTimeEntryTypeByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteMaintenanceOrderTimeEntryType(ctx, id)
}

// ── Converters ────────────────────────────────────────────────────────────────

func statusToResp(s db.MaintenanceOrderStatus) StatusResponse {
	return StatusResponse{
		ID:              s.MaintenanceOrderStatusID,
		Name:            s.Name,
		BackgroundColor: s.BackgroundColor,
		BorderColor:     s.BorderColor,
		TextColor:       s.TextColor,
		SortIndex:       s.SortIndex,
		IsClosed:        s.IsClosed,
		IsDefault:       s.IsDefault,
	}
}

func priorityToResp(p db.MaintenanceOrderPriority) PriorityResponse {
	return PriorityResponse{
		ID:              p.MaintenanceOrderPrioritiesID,
		Name:            p.Name,
		BackgroundColor: p.BackgroundColor,
		BorderColor:     p.BorderColor,
		TextColor:       p.TextColor,
		SortIndex:       p.SortIndex,
	}
}

func timeEntryTypeToResp(t db.MaintenanceOrderTimeEntryType) TimeEntryTypeResponse {
	r := TimeEntryTypeResponse{ID: t.MaintenanceOrderTimeEntryTypeID, Name: t.Name}
	if t.Color.Valid {
		r.Color = &t.Color.String
	}
	return r
}
