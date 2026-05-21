package fiscal

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
	statuses, err := s.q.ListFiscalOrderStatuses(ctx)
	if err != nil {
		return nil, err
	}
	priorities, err := s.q.ListFiscalOrderPriorities(ctx)
	if err != nil {
		return nil, err
	}
	reasons, err := s.q.ListFiscalOrderReasons(ctx)
	if err != nil {
		return nil, err
	}
	transitions, err := s.q.ListFiscalStatusTransitions(ctx)
	if err != nil {
		return nil, err
	}

	resp := &LookupsResponse{
		Status:      make([]StatusResponse, len(statuses)),
		Priorities:  make([]PriorityResponse, len(priorities)),
		Reasons:     make([]ReasonResponse, len(reasons)),
		Transitions: make([]TransitionResponse, len(transitions)),
	}
	for i, s := range statuses {
		resp.Status[i] = statusToResp(s)
	}
	for i, p := range priorities {
		resp.Priorities[i] = priorityToResp(p)
	}
	for i, r := range reasons {
		resp.Reasons[i] = ReasonResponse{
			ID:                r.FiscalOrderReasonID,
			Name:              r.Name,
			InitialStatusID:   r.InitialStatusID,
			InitialStatusName: r.InitialStatusName,
		}
	}
	for i, t := range transitions {
		resp.Transitions[i] = transitionToResp(t)
	}
	return resp, nil
}

// ── Status ────────────────────────────────────────────────────────────────────

func (s *LookupsService) ListStatuses(ctx context.Context) ([]StatusResponse, error) {
	rows, err := s.q.ListFiscalOrderStatuses(ctx)
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
	row, err := s.q.GetFiscalOrderStatusByID(ctx, id)
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
	res, err := s.q.CreateFiscalOrderStatus(ctx, db.CreateFiscalOrderStatusParams{
		Name:              req.Name,
		BackgroundColor:   req.BackgroundColor,
		BorderColor:       req.BorderColor,
		TextColor:         req.TextColor,
		SortIndex:         req.SortIndex,
		IsClosed:          req.IsClosed,
		IsDefault:         req.IsDefault,
		IsPendingApproval: req.IsPendingApproval,
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
	if err := s.q.UpdateFiscalStatusRecord(ctx, db.UpdateFiscalStatusRecordParams{
		Name:                req.Name,
		BackgroundColor:     req.BackgroundColor,
		BorderColor:         req.BorderColor,
		TextColor:           req.TextColor,
		SortIndex:           req.SortIndex,
		IsClosed:            req.IsClosed,
		IsDefault:           req.IsDefault,
		IsPendingApproval:   req.IsPendingApproval,
		FiscalOrderStatusID: id,
	}); err != nil {
		return nil, err
	}
	return s.GetStatusByID(ctx, id)
}

func (s *LookupsService) DeleteStatus(ctx context.Context, id int32) error {
	if _, err := s.GetStatusByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteFiscalOrderStatus(ctx, id)
}

// ── Priority ──────────────────────────────────────────────────────────────────

func (s *LookupsService) ListPriorities(ctx context.Context) ([]PriorityResponse, error) {
	rows, err := s.q.ListFiscalOrderPriorities(ctx)
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
	row, err := s.q.GetFiscalOrderPriorityByID(ctx, id)
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
	res, err := s.q.CreateFiscalOrderPriority(ctx, db.CreateFiscalOrderPriorityParams{
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
	if err := s.q.UpdateFiscalOrderPriority(ctx, db.UpdateFiscalOrderPriorityParams{
		Name:                  req.Name,
		BackgroundColor:       req.BackgroundColor,
		BorderColor:           req.BorderColor,
		TextColor:             req.TextColor,
		SortIndex:             req.SortIndex,
		FiscalOrderPriorityID: id,
	}); err != nil {
		return nil, err
	}
	return s.GetPriorityByID(ctx, id)
}

func (s *LookupsService) DeletePriority(ctx context.Context, id int32) error {
	if _, err := s.GetPriorityByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteFiscalOrderPriority(ctx, id)
}

// ── Reason ────────────────────────────────────────────────────────────────────

func (s *LookupsService) ListReasons(ctx context.Context) ([]ReasonResponse, error) {
	rows, err := s.q.ListFiscalOrderReasons(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ReasonResponse, len(rows))
	for i, r := range rows {
		out[i] = ReasonResponse{
			ID:                r.FiscalOrderReasonID,
			Name:              r.Name,
			InitialStatusID:   r.InitialStatusID,
			InitialStatusName: r.InitialStatusName,
		}
	}
	return out, nil
}

func (s *LookupsService) GetReasonByID(ctx context.Context, id int32) (*ReasonResponse, error) {
	row, err := s.q.GetFiscalOrderReasonByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("motivo não encontrado")
	}
	if err != nil {
		return nil, err
	}
	r := reasonToResp(row)
	return &r, nil
}

func (s *LookupsService) CreateReason(ctx context.Context, req CreateReasonRequest) (*ReasonResponse, error) {
	res, err := s.q.CreateFiscalOrderReason(ctx, db.CreateFiscalOrderReasonParams{
		Name:            req.Name,
		InitialStatusID: req.InitialStatusID,
	})
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
	if err := s.q.UpdateFiscalOrderReason(ctx, db.UpdateFiscalOrderReasonParams{
		Name:                req.Name,
		InitialStatusID:     req.InitialStatusID,
		FiscalOrderReasonID: id,
	}); err != nil {
		return nil, err
	}
	return s.GetReasonByID(ctx, id)
}

func (s *LookupsService) DeleteReason(ctx context.Context, id int32) error {
	if _, err := s.GetReasonByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteFiscalOrderReason(ctx, id)
}

// ── Transitions ───────────────────────────────────────────────────────────────

func (s *LookupsService) ListTransitions(ctx context.Context) ([]TransitionResponse, error) {
	rows, err := s.q.ListFiscalStatusTransitions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TransitionResponse, len(rows))
	for i, r := range rows {
		out[i] = transitionToResp(r)
	}
	return out, nil
}

func (s *LookupsService) CreateTransition(ctx context.Context, req CreateTransitionRequest) (*TransitionResponse, error) {
	fromID := sql.NullInt32{}
	if req.FromStatusID != nil {
		fromID = sql.NullInt32{Int32: *req.FromStatusID, Valid: true}
	}
	res, err := s.q.CreateFiscalStatusTransition(ctx, db.CreateFiscalStatusTransitionParams{
		FromStatusID:  fromID,
		ToStatusID:    req.ToStatusID,
		AccessLevelID: req.AccessLevelID,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	r := TransitionResponse{
		ID:            int32(id),
		ToStatusID:    req.ToStatusID,
		AccessLevelID: req.AccessLevelID,
	}
	if req.FromStatusID != nil {
		r.FromStatusID = req.FromStatusID
	}
	return &r, nil
}

func (s *LookupsService) DeleteTransition(ctx context.Context, id int32) error {
	return s.q.DeleteFiscalStatusTransition(ctx, id)
}

// ── Converters ────────────────────────────────────────────────────────────────

func statusToResp(s db.FiscalOrderStatus) StatusResponse {
	return StatusResponse{
		ID:                s.FiscalOrderStatusID,
		Name:              s.Name,
		BackgroundColor:   s.BackgroundColor,
		BorderColor:       s.BorderColor,
		TextColor:         s.TextColor,
		SortIndex:         s.SortIndex,
		IsClosed:          s.IsClosed,
		IsDefault:         s.IsDefault,
		IsPendingApproval: s.IsPendingApproval,
	}
}

func priorityToResp(p db.FiscalOrderPriority) PriorityResponse {
	return PriorityResponse{
		ID:              p.FiscalOrderPriorityID,
		Name:            p.Name,
		BackgroundColor: p.BackgroundColor,
		BorderColor:     p.BorderColor,
		TextColor:       p.TextColor,
		SortIndex:       p.SortIndex,
	}
}

func reasonToResp(r db.GetFiscalOrderReasonByIDRow) ReasonResponse {
	return ReasonResponse{
		ID:                r.FiscalOrderReasonID,
		Name:              r.Name,
		InitialStatusID:   r.InitialStatusID,
		InitialStatusName: r.InitialStatusName,
	}
}

func transitionToResp(t db.FiscalOrderStatusTransition) TransitionResponse {
	r := TransitionResponse{
		ID:            t.ID,
		ToStatusID:    t.ToStatusID,
		AccessLevelID: t.AccessLevelID,
	}
	if t.FromStatusID.Valid {
		v := t.FromStatusID.Int32
		r.FromStatusID = &v
	}
	return r
}
