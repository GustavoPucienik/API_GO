package logistic

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
	statuses, err := s.q.ListLogisticOrderStatuses(ctx)
	if err != nil {
		return nil, err
	}
	priorities, err := s.q.ListLogisticOrderPriorities(ctx)
	if err != nil {
		return nil, err
	}
	reasons, err := s.q.ListLogisticOrderReasons(ctx)
	if err != nil {
		return nil, err
	}
	transitions, err := s.q.ListLogisticStatusTransitions(ctx)
	if err != nil {
		return nil, err
	}
	operators, err := s.q.ListLogisticOperators(ctx)
	if err != nil {
		return nil, err
	}

	resp := &LookupsResponse{
		Status:      make([]StatusResponse, len(statuses)),
		Priorities:  make([]PriorityResponse, len(priorities)),
		Reasons:     make([]ReasonResponse, len(reasons)),
		Transitions: make([]TransitionResponse, len(transitions)),
		Operators:   make([]OperatorResponse, len(operators)),
	}
	for i, st := range statuses {
		resp.Status[i] = statusToResp(st)
	}
	for i, p := range priorities {
		resp.Priorities[i] = priorityToResp(p)
	}
	for i, r := range reasons {
		resp.Reasons[i] = ReasonResponse{
			ID:                r.LogisticOrderReasonID,
			Name:              r.Name,
			InitialStatusID:   r.InitialStatusID,
			InitialStatusName: r.InitialStatusName,
		}
	}
	for i, t := range transitions {
		resp.Transitions[i] = transitionToResp(t)
	}
	for i, o := range operators {
		resp.Operators[i] = operatorToResp(o)
	}
	return resp, nil
}

// ── Status ────────────────────────────────────────────────────────────────────

func (s *LookupsService) ListStatuses(ctx context.Context) ([]StatusResponse, error) {
	rows, err := s.q.ListLogisticOrderStatuses(ctx)
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
	row, err := s.q.GetLogisticOrderStatusByID(ctx, id)
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
	res, err := s.q.CreateLogisticOrderStatus(ctx, db.CreateLogisticOrderStatusParams{
		Name:              req.Name,
		BackgroundColor:   req.BackgroundColor,
		BorderColor:       req.BorderColor,
		TextColor:         req.TextColor,
		SortIndex:         req.SortIndex,
		IsClosed:          req.IsClosed,
		IsPendingApproval: req.IsPendingApproval,
		IsRhStep:          req.IsRhStep,
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
	if err := s.q.UpdateLogisticStatusRecord(ctx, db.UpdateLogisticStatusRecordParams{
		Name:                  req.Name,
		BackgroundColor:       req.BackgroundColor,
		BorderColor:           req.BorderColor,
		TextColor:             req.TextColor,
		SortIndex:             req.SortIndex,
		IsClosed:              req.IsClosed,
		IsPendingApproval:     req.IsPendingApproval,
		IsRhStep:              req.IsRhStep,
		LogisticOrderStatusID: id,
	}); err != nil {
		return nil, err
	}
	return s.GetStatusByID(ctx, id)
}

func (s *LookupsService) DeleteStatus(ctx context.Context, id int32) error {
	if _, err := s.GetStatusByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteLogisticOrderStatus(ctx, id)
}

// ── Priority ──────────────────────────────────────────────────────────────────

func (s *LookupsService) ListPriorities(ctx context.Context) ([]PriorityResponse, error) {
	rows, err := s.q.ListLogisticOrderPriorities(ctx)
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
	row, err := s.q.GetLogisticOrderPriorityByID(ctx, id)
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
	res, err := s.q.CreateLogisticOrderPriority(ctx, db.CreateLogisticOrderPriorityParams{
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
	if err := s.q.UpdateLogisticOrderPriority(ctx, db.UpdateLogisticOrderPriorityParams{
		Name:                    req.Name,
		BackgroundColor:         req.BackgroundColor,
		BorderColor:             req.BorderColor,
		TextColor:               req.TextColor,
		SortIndex:               req.SortIndex,
		LogisticOrderPriorityID: id,
	}); err != nil {
		return nil, err
	}
	return s.GetPriorityByID(ctx, id)
}

func (s *LookupsService) DeletePriority(ctx context.Context, id int32) error {
	if _, err := s.GetPriorityByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteLogisticOrderPriority(ctx, id)
}

// ── Reason ────────────────────────────────────────────────────────────────────

func (s *LookupsService) ListReasons(ctx context.Context) ([]ReasonResponse, error) {
	rows, err := s.q.ListLogisticOrderReasons(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]ReasonResponse, len(rows))
	for i, r := range rows {
		out[i] = ReasonResponse{
			ID:                r.LogisticOrderReasonID,
			Name:              r.Name,
			InitialStatusID:   r.InitialStatusID,
			InitialStatusName: r.InitialStatusName,
		}
	}
	return out, nil
}

func (s *LookupsService) GetReasonByID(ctx context.Context, id int32) (*ReasonResponse, error) {
	row, err := s.q.GetLogisticOrderReasonByID(ctx, id)
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
	res, err := s.q.CreateLogisticOrderReason(ctx, db.CreateLogisticOrderReasonParams{
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
	if err := s.q.UpdateLogisticOrderReason(ctx, db.UpdateLogisticOrderReasonParams{
		Name:                  req.Name,
		InitialStatusID:       req.InitialStatusID,
		LogisticOrderReasonID: id,
	}); err != nil {
		return nil, err
	}
	return s.GetReasonByID(ctx, id)
}

func (s *LookupsService) DeleteReason(ctx context.Context, id int32) error {
	if _, err := s.GetReasonByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteLogisticOrderReason(ctx, id)
}

// ── Transitions ───────────────────────────────────────────────────────────────

func (s *LookupsService) ListTransitions(ctx context.Context) ([]TransitionResponse, error) {
	rows, err := s.q.ListLogisticStatusTransitions(ctx)
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
	res, err := s.q.CreateLogisticStatusTransition(ctx, db.CreateLogisticStatusTransitionParams{
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
	return s.q.DeleteLogisticStatusTransition(ctx, id)
}

// ── Operators ─────────────────────────────────────────────────────────────────

func (s *LookupsService) ListOperators(ctx context.Context) ([]OperatorResponse, error) {
	rows, err := s.q.ListLogisticOperators(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]OperatorResponse, len(rows))
	for i, o := range rows {
		out[i] = operatorToResp(o)
	}
	return out, nil
}

func (s *LookupsService) GetOperatorByID(ctx context.Context, id int32) (*OperatorResponse, error) {
	row, err := s.q.GetLogisticOperatorByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("operador não encontrado")
	}
	if err != nil {
		return nil, err
	}
	r := operatorToResp(row)
	return &r, nil
}

func (s *LookupsService) CreateOperator(ctx context.Context, req CreateOperatorRequest) (*OperatorResponse, error) {
	res, err := s.q.CreateLogisticOperator(ctx, db.CreateLogisticOperatorParams{
		Name: req.Name,
		Type: db.LogisticOperatorType(req.Type),
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.GetOperatorByID(ctx, int32(id))
}

func (s *LookupsService) UpdateOperator(ctx context.Context, id int32, req UpdateOperatorRequest) (*OperatorResponse, error) {
	if _, err := s.GetOperatorByID(ctx, id); err != nil {
		return nil, err
	}
	if err := s.q.UpdateLogisticOperator(ctx, db.UpdateLogisticOperatorParams{
		Name:               req.Name,
		Type:               db.LogisticOperatorType(req.Type),
		LogisticOperatorID: id,
	}); err != nil {
		return nil, err
	}
	return s.GetOperatorByID(ctx, id)
}

func (s *LookupsService) DeleteOperator(ctx context.Context, id int32) error {
	if _, err := s.GetOperatorByID(ctx, id); err != nil {
		return err
	}
	return s.q.DeleteLogisticOperator(ctx, id)
}

// ── Converters ────────────────────────────────────────────────────────────────

func statusToResp(s db.LogisticOrderStatus) StatusResponse {
	return StatusResponse{
		ID:                s.LogisticOrderStatusID,
		Name:              s.Name,
		BackgroundColor:   s.BackgroundColor,
		BorderColor:       s.BorderColor,
		TextColor:         s.TextColor,
		SortIndex:         s.SortIndex,
		IsClosed:          s.IsClosed,
		IsPendingApproval: s.IsPendingApproval,
		IsRhStep:          s.IsRhStep,
	}
}

func priorityToResp(p db.LogisticOrderPriority) PriorityResponse {
	return PriorityResponse{
		ID:              p.LogisticOrderPriorityID,
		Name:            p.Name,
		BackgroundColor: p.BackgroundColor,
		BorderColor:     p.BorderColor,
		TextColor:       p.TextColor,
		SortIndex:       p.SortIndex,
	}
}

func reasonToResp(r db.GetLogisticOrderReasonByIDRow) ReasonResponse {
	return ReasonResponse{
		ID:                r.LogisticOrderReasonID,
		Name:              r.Name,
		InitialStatusID:   r.InitialStatusID,
		InitialStatusName: r.InitialStatusName,
	}
}

func transitionToResp(t db.LogisticOrderStatusTransition) TransitionResponse {
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

func operatorToResp(o db.LogisticOperator) OperatorResponse {
	return OperatorResponse{
		ID:   o.LogisticOperatorID,
		Name: o.Name,
		Type: string(o.Type),
	}
}
