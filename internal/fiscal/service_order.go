package fiscal

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"systemapi/internal/db"
)

type OrderService struct {
	sqlDB *sql.DB
	q     *db.Queries
}

func NewOrderService(sqlDB *sql.DB, q *db.Queries) *OrderService {
	return &OrderService{sqlDB: sqlDB, q: q}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (s *OrderService) Create(ctx context.Context, req CreateOrderRequest) (*OrderResponse, error) {
	requester, err := s.q.GetUserByID(ctx, int32(req.RequesterID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("usuário requisitante não encontrado")
	}
	if err != nil {
		return nil, err
	}
	requesterName := requester.Name
	if requester.LastName.Valid {
		requesterName += " " + requester.LastName.String
	}

	var statusID int32
	if req.StatusID != nil {
		statusID = *req.StatusID
	} else {
		reason, err := s.q.GetFiscalOrderReasonByID(ctx, req.ReasonID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("motivo não encontrado")
		}
		if err != nil {
			return nil, err
		}
		statusID = reason.InitialStatusID
	}

	res, err := s.q.CreateFiscalOrder(ctx, db.CreateFiscalOrderParams{
		RequesterID:     int32(req.RequesterID),
		RequesterName:   requesterName,
		SalespersonName: nullStr(req.SalespersonName),
		ClientCode:      req.ClientCode,
		ClientName:      req.ClientName,
		ClientAliasName: req.AliasName,
		StatusID:        statusID,
		PriorityID:      req.PriorityID,
		ReasonID:        req.ReasonID,
		Description:     req.Description,
	})
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return s.loadOrder(ctx, int32(id))
}

// ── FindByID ──────────────────────────────────────────────────────────────────

func (s *OrderService) FindByID(ctx context.Context, id int32) (*OrderResponse, error) {
	return s.loadOrder(ctx, id)
}

func (s *OrderService) loadOrder(ctx context.Context, id int32) (*OrderResponse, error) {
	o, err := s.q.GetFiscalOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem fiscal não encontrada")
	}
	if err != nil {
		return nil, err
	}
	return rowToOrderResponse(o), nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *OrderService) Update(ctx context.Context, id int32, req UpdateOrderRequest) (*OrderResponse, error) {
	order, err := s.q.GetFiscalOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem fiscal não encontrada")
	}
	if err != nil {
		return nil, err
	}

	setClauses := []string{}
	args := []any{}
	apply := func(col string, val any) {
		setClauses = append(setClauses, col+" = ?")
		args = append(args, val)
	}

	if req.StatusID != nil {
		apply("status_id", *req.StatusID)
		st, err := s.q.GetFiscalOrderStatusByID(ctx, *req.StatusID)
		if err == nil {
			prevSt, _ := s.q.GetFiscalOrderStatusByID(ctx, order.StatusID)
			if st.IsClosed && !prevSt.IsClosed {
				apply("closed_at", time.Now())
			} else if !st.IsClosed && prevSt.IsClosed {
				apply("closed_at", nil)
			}
		}
	}
	if req.PriorityID != nil {
		apply("priority_id", *req.PriorityID)
	}
	if req.ReasonID != nil {
		apply("reason_id", *req.ReasonID)
	}
	if req.Description != nil {
		apply("description", *req.Description)
	}
	if req.SalespersonName != nil {
		apply("salesperson_name", *req.SalespersonName)
	}
	if req.ClientCode != nil {
		apply("client_code", *req.ClientCode)
	}
	if req.ClientName != nil {
		apply("client_name", *req.ClientName)
	}
	if req.AliasName != nil {
		apply("client_alias_name", *req.AliasName)
	}
	if req.Resolution != nil {
		apply("resolution", *req.Resolution)
	}
	if req.ClosedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ClosedAt)
		if err != nil {
			t, _ = time.Parse("2006-01-02T15:04:05", *req.ClosedAt)
		}
		apply("closed_at", t)
	}

	if len(setClauses) == 0 {
		return s.loadOrder(ctx, id)
	}

	query := fmt.Sprintf("UPDATE fiscal_order SET %s WHERE fiscal_order_id = ? AND deleted_at IS NULL",
		strings.Join(setClauses, ", "))
	args = append(args, id)
	if _, err := s.sqlDB.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}
	return s.loadOrder(ctx, id)
}

// ── Assign ────────────────────────────────────────────────────────────────────

func (s *OrderService) Assign(ctx context.Context, id int32, userID int64) (*OrderResponse, error) {
	if _, err := s.q.GetFiscalOrderByID(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem fiscal não encontrada")
	} else if err != nil {
		return nil, err
	}

	user, err := s.q.GetUserByID(ctx, int32(userID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("usuário não encontrado")
	}
	if err != nil {
		return nil, err
	}

	name := user.Name
	if user.LastName.Valid {
		name += " " + user.LastName.String
	}

	if err := s.q.AssignFiscalOrder(ctx, db.AssignFiscalOrderParams{
		AssignedToID:   sql.NullInt32{Int32: int32(userID), Valid: true},
		AssignedToName: sql.NullString{String: name, Valid: true},
		FiscalOrderID:  id,
	}); err != nil {
		return nil, err
	}
	return s.loadOrder(ctx, id)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (s *OrderService) Delete(ctx context.Context, id int32) error {
	if _, err := s.q.GetFiscalOrderByID(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return errors.New("ordem fiscal não encontrada")
	} else if err != nil {
		return err
	}
	return s.q.DeleteFiscalOrder(ctx, id)
}

// ── FindPaginated ─────────────────────────────────────────────────────────────

func (s *OrderService) FindPaginated(ctx context.Context, p FindPaginatedParams) ([]OrderResponse, int64, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	offset := (p.Page - 1) * p.Limit

	sortDir := "DESC"
	if strings.ToUpper(p.SortOrder) == "ASC" {
		sortDir = "ASC"
	}

	baseSelect := `
SELECT fo.fiscal_order_id, fo.requester_id, fo.requester_name,
    fo.assigned_to_id, fo.assigned_to_name, fo.salesperson_name,
    fo.client_code, fo.client_name, fo.client_alias_name,
    fo.status_id, fo.priority_id, fo.reason_id, fo.description,
    fo.captured_at, fo.resolution, fo.closed_at, fo.created_at, fo.updated_at,
    s.name AS status_name, s.background_color AS status_bg, s.border_color AS status_border,
    s.text_color AS status_text, s.is_closed AS status_is_closed,
    p.name AS priority_name, p.background_color AS priority_bg,
    r.name AS reason_name`

	baseFrom := `
FROM fiscal_order fo
JOIN fiscal_order_status s ON s.fiscal_order_status_id = fo.status_id
JOIN fiscal_order_priorities p ON p.fiscal_order_priority_id = fo.priority_id
JOIN fiscal_order_reason r ON r.fiscal_order_reason_id = fo.reason_id`

	where := []string{"fo.deleted_at IS NULL"}
	args := []any{}

	if p.Search != "" {
		where = append(where, "(fo.client_name LIKE ? OR fo.requester_name LIKE ? OR fo.assigned_to_name LIKE ? OR r.name LIKE ?)")
		like := "%" + p.Search + "%"
		args = append(args, like, like, like, like)
	}
	if p.StatusID != nil {
		where = append(where, "fo.status_id = ?")
		args = append(args, *p.StatusID)
	}
	if p.PriorityID != nil {
		where = append(where, "fo.priority_id = ?")
		args = append(args, *p.PriorityID)
	}
	if p.ReasonID != nil {
		where = append(where, "fo.reason_id = ?")
		args = append(args, *p.ReasonID)
	}
	if p.ResponsibleID != nil {
		where = append(where, "fo.assigned_to_id = ?")
		args = append(args, *p.ResponsibleID)
	}
	if p.StartDate != nil {
		where = append(where, "fo.created_at >= ?")
		args = append(args, *p.StartDate)
	}
	if p.EndDate != nil {
		where = append(where, "fo.created_at <= ?")
		args = append(args, *p.EndDate)
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int64
	countQuery := "SELECT COUNT(*) " + baseFrom + " " + whereClause
	if err := s.sqlDB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := baseSelect + baseFrom + " " + whereClause +
		" ORDER BY fo.created_at " + sortDir +
		fmt.Sprintf(" LIMIT %d OFFSET %d", p.Limit, offset)

	rows, err := s.sqlDB.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []OrderResponse
	for rows.Next() {
		var o db.GetFiscalOrderByIDRow
		if err := rows.Scan(
			&o.FiscalOrderID, &o.RequesterID, &o.RequesterName,
			&o.AssignedToID, &o.AssignedToName, &o.SalespersonName,
			&o.ClientCode, &o.ClientName, &o.ClientAliasName,
			&o.StatusID, &o.PriorityID, &o.ReasonID, &o.Description,
			&o.CapturedAt, &o.Resolution, &o.ClosedAt, &o.CreatedAt, &o.UpdatedAt,
			&o.StatusName, &o.StatusBg, &o.StatusBorder, &o.StatusText, &o.StatusIsClosed,
			&o.PriorityName, &o.PriorityBg, &o.ReasonName,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, *rowToOrderResponse(o))
	}
	if out == nil {
		out = []OrderResponse{}
	}
	return out, total, nil
}

// ── FindPendingApproval ───────────────────────────────────────────────────────

func (s *OrderService) FindPendingApproval(ctx context.Context, p FindPaginatedParams) ([]OrderResponse, int64, error) {
	p2 := p
	// Filter by statuses with is_pending_approval = true
	statuses, err := s.q.ListFiscalOrderStatuses(ctx)
	if err != nil {
		return nil, 0, err
	}
	var pendingIDs []string
	var pendingArgs []any
	for _, st := range statuses {
		if st.IsPendingApproval {
			pendingIDs = append(pendingIDs, "?")
			pendingArgs = append(pendingArgs, st.FiscalOrderStatusID)
		}
	}
	if len(pendingIDs) == 0 {
		return []OrderResponse{}, 0, nil
	}
	_ = p2
	// Build query using raw SQL with pending status filter
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	offset := (p.Page - 1) * p.Limit
	sortDir := "DESC"
	if strings.ToUpper(p.SortOrder) == "ASC" {
		sortDir = "ASC"
	}

	baseSelect := `SELECT fo.fiscal_order_id, fo.requester_id, fo.requester_name,
    fo.assigned_to_id, fo.assigned_to_name, fo.salesperson_name,
    fo.client_code, fo.client_name, fo.client_alias_name,
    fo.status_id, fo.priority_id, fo.reason_id, fo.description,
    fo.captured_at, fo.resolution, fo.closed_at, fo.created_at, fo.updated_at,
    s.name, s.background_color, s.border_color, s.text_color, s.is_closed,
    p.name, p.background_color, r.name`
	baseFrom := ` FROM fiscal_order fo
JOIN fiscal_order_status s ON s.fiscal_order_status_id = fo.status_id
JOIN fiscal_order_priorities p ON p.fiscal_order_priority_id = fo.priority_id
JOIN fiscal_order_reason r ON r.fiscal_order_reason_id = fo.reason_id`

	where := []string{"fo.deleted_at IS NULL", "fo.status_id IN (" + strings.Join(pendingIDs, ",") + ")"}
	args := append([]any{}, pendingArgs...)

	if p.Search != "" {
		where = append(where, "(fo.client_name LIKE ? OR fo.requester_name LIKE ?)")
		like := "%" + p.Search + "%"
		args = append(args, like, like)
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int64
	countArgs := append([]any{}, args...)
	if err := s.sqlDB.QueryRowContext(ctx, "SELECT COUNT(*) "+baseFrom+" "+whereClause, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := baseSelect + baseFrom + " " + whereClause +
		" ORDER BY fo.created_at " + sortDir +
		fmt.Sprintf(" LIMIT %d OFFSET %d", p.Limit, offset)

	rows, err := s.sqlDB.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []OrderResponse
	for rows.Next() {
		var o db.GetFiscalOrderByIDRow
		if err := rows.Scan(
			&o.FiscalOrderID, &o.RequesterID, &o.RequesterName,
			&o.AssignedToID, &o.AssignedToName, &o.SalespersonName,
			&o.ClientCode, &o.ClientName, &o.ClientAliasName,
			&o.StatusID, &o.PriorityID, &o.ReasonID, &o.Description,
			&o.CapturedAt, &o.Resolution, &o.ClosedAt, &o.CreatedAt, &o.UpdatedAt,
			&o.StatusName, &o.StatusBg, &o.StatusBorder, &o.StatusText, &o.StatusIsClosed,
			&o.PriorityName, &o.PriorityBg, &o.ReasonName,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, *rowToOrderResponse(o))
	}
	if out == nil {
		out = []OrderResponse{}
	}
	return out, total, nil
}

// ── Converters ────────────────────────────────────────────────────────────────

func rowToOrderResponse(o db.GetFiscalOrderByIDRow) *OrderResponse {
	r := &OrderResponse{
		ID:              o.FiscalOrderID,
		RequesterID:     o.RequesterID,
		RequesterName:   o.RequesterName,
		ClientCode:      o.ClientCode,
		ClientName:      o.ClientName,
		ClientAliasName: o.ClientAliasName,
		StatusID:        o.StatusID,
		PriorityID:      o.PriorityID,
		ReasonID:        o.ReasonID,
		Description:     o.Description,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
		ReasonName:      o.ReasonName,
		Status: StatusSummary{
			ID:              o.StatusID,
			Name:            o.StatusName,
			BackgroundColor: o.StatusBg,
			BorderColor:     o.StatusBorder,
			TextColor:       o.StatusText,
			IsClosed:        o.StatusIsClosed,
		},
		Priority: PrioritySummary{
			ID:              o.PriorityID,
			Name:            o.PriorityName,
			BackgroundColor: o.PriorityBg,
		},
	}
	if o.AssignedToID.Valid {
		v := o.AssignedToID.Int32
		r.AssignedToID = &v
	}
	if o.AssignedToName.Valid {
		r.AssignedToName = &o.AssignedToName.String
	}
	if o.SalespersonName.Valid {
		r.SalespersonName = &o.SalespersonName.String
	}
	if o.CapturedAt.Valid {
		t := o.CapturedAt.Time
		r.CapturedAt = &t
	}
	if o.Resolution.Valid {
		r.Resolution = &o.Resolution.String
	}
	if o.ClosedAt.Valid {
		t := o.ClosedAt.Time
		r.ClosedAt = &t
	}
	return r
}

func nullStr(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}
