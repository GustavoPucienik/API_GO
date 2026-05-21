package logistic

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
	requester, err := s.q.GetUserByID(ctx, req.RequesterID)
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
		reason, err := s.q.GetLogisticOrderReasonByID(ctx, req.ReasonID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("motivo não encontrado")
		}
		if err != nil {
			return nil, err
		}
		statusID = reason.InitialStatusID
	}

	var schedDate sql.NullTime
	if req.ScheduleDate != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduleDate)
		if err != nil {
			t, _ = time.Parse("2006-01-02T15:04:05", *req.ScheduleDate)
		}
		schedDate = sql.NullTime{Time: t, Valid: true}
	}

	res, err := s.q.CreateLogisticOrder(ctx, db.CreateLogisticOrderParams{
		RequesterID:       req.RequesterID,
		RequesterName:     requesterName,
		SalespersonName:   nullStr(req.SalespersonName),
		ClientCode:        req.ClientCode,
		ClientName:        req.ClientName,
		ClientAliasName:   req.AliasName,
		ContactType:       nullStr(req.ContactType),
		ContactName:       nullStr(req.ContactName),
		ContactPhone:      nullStr(req.ContactPhone),
		ContactMobile:     nullStr(req.ContactMobile),
		ContactEmail:      nullStr(req.ContactEmail),
		AddressName:       nullStr(req.AddressName),
		AddressType:       nullStr(req.AddressType),
		TypeOfAddress:     nullStr(req.TypeOfAddress),
		Street:            nullStr(req.Street),
		StreetNo:          nullStr(req.StreetNo),
		Block:             nullStr(req.Block),
		City:              nullStr(req.City),
		County:            nullStr(req.County),
		State:             nullStr(req.State),
		Country:           nullStr(req.Country),
		ZipCode:           nullStr(req.ZipCode),
		BuildingFloorRoom: nullStr(req.BuildingFloorRoom),
		StatusID:          statusID,
		PriorityID:        req.PriorityID,
		ReasonID:          req.ReasonID,
		ScheduleDate:      schedDate,
		Description:       req.Description,
	})
	if err != nil {
		return nil, err
	}
	id64, _ := res.LastInsertId()
	orderID := int32(id64)

	for _, item := range req.Items {
		if _, err := s.q.CreateLogisticOrderItem(ctx, db.CreateLogisticOrderItemParams{
			LogisticOrderID: orderID,
			Name:            item.ItemCode,
			ItemCode:        item.ItemCode,
			Quantity:        item.Quantity,
		}); err != nil {
			return nil, err
		}
	}

	return s.loadOrder(ctx, orderID)
}

// ── FindByID ──────────────────────────────────────────────────────────────────

func (s *OrderService) FindByID(ctx context.Context, id int32) (*OrderResponse, error) {
	return s.loadOrder(ctx, id)
}

func (s *OrderService) loadOrder(ctx context.Context, id int32) (*OrderResponse, error) {
	o, err := s.q.GetLogisticOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem logística não encontrada")
	}
	if err != nil {
		return nil, err
	}
	resp := rowToOrderResponse(o)

	itemRows, err := s.q.ListLogisticOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	resp.Items = make([]ItemResponse, len(itemRows))
	for i, it := range itemRows {
		resp.Items[i] = ItemResponse{
			ID:       it.LogisticOrderItemID,
			Name:     it.Name,
			ItemCode: it.ItemCode,
			Quantity: it.Quantity,
		}
	}
	return resp, nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *OrderService) Update(ctx context.Context, id int32, req UpdateOrderRequest) (*OrderResponse, error) {
	order, err := s.q.GetLogisticOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem logística não encontrada")
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

	if req.AssignedToID != nil {
		apply("assigned_to_id", int32(*req.AssignedToID))
		user, err := s.q.GetUserByID(ctx, int32(*req.AssignedToID))
		if err == nil {
			name := user.Name
			if user.LastName.Valid {
				name += " " + user.LastName.String
			}
			apply("assigned_to_name", name)
		}
	}
	if req.StatusID != nil {
		apply("status_id", *req.StatusID)
		st, err := s.q.GetLogisticOrderStatusByID(ctx, *req.StatusID)
		if err == nil {
			prevSt, _ := s.q.GetLogisticOrderStatusByID(ctx, order.StatusID)
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
	if req.ContactType != nil {
		apply("contact_type", *req.ContactType)
	}
	if req.ContactName != nil {
		apply("contact_name", *req.ContactName)
	}
	if req.ContactPhone != nil {
		apply("contact_phone", *req.ContactPhone)
	}
	if req.ContactMobile != nil {
		apply("contact_mobile", *req.ContactMobile)
	}
	if req.ContactEmail != nil {
		apply("contact_email", *req.ContactEmail)
	}
	if req.AddressName != nil {
		apply("address_name", *req.AddressName)
	}
	if req.AddressType != nil {
		apply("address_type", *req.AddressType)
	}
	if req.TypeOfAddress != nil {
		apply("type_of_address", *req.TypeOfAddress)
	}
	if req.Street != nil {
		apply("street", *req.Street)
	}
	if req.StreetNo != nil {
		apply("street_no", *req.StreetNo)
	}
	if req.Block != nil {
		apply("block", *req.Block)
	}
	if req.City != nil {
		apply("city", *req.City)
	}
	if req.County != nil {
		apply("county", *req.County)
	}
	if req.State != nil {
		apply("state", *req.State)
	}
	if req.Country != nil {
		apply("country", *req.Country)
	}
	if req.ZipCode != nil {
		apply("zip_code", *req.ZipCode)
	}
	if req.BuildingFloorRoom != nil {
		apply("building_floor_room", *req.BuildingFloorRoom)
	}
	if req.DriverName != nil {
		apply("driver_name", *req.DriverName)
	}
	if req.CheckerName != nil {
		apply("checker_name", *req.CheckerName)
	}
	if req.HasSalaryDiscount != nil {
		apply("has_salary_discount", *req.HasSalaryDiscount)
	}
	if req.SalaryDiscountReason != nil {
		apply("salary_discount_reason", *req.SalaryDiscountReason)
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
	if req.ScheduleDate != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduleDate)
		if err != nil {
			t, _ = time.Parse("2006-01-02T15:04:05", *req.ScheduleDate)
		}
		apply("schedule_date", t)
	}

	if req.Items != nil {
		tx, err := s.sqlDB.BeginTx(ctx, nil)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		if len(setClauses) > 0 {
			query := fmt.Sprintf("UPDATE logistic_order SET %s WHERE logistic_order_id = ? AND deleted_at IS NULL",
				strings.Join(setClauses, ", "))
			args = append(args, id)
			if _, err := tx.ExecContext(ctx, query, args...); err != nil {
				return nil, err
			}
		}

		qtx := s.q.WithTx(tx)
		if err := qtx.DeleteLogisticOrderItems(ctx, id); err != nil {
			return nil, err
		}
		for _, item := range *req.Items {
			if _, err := qtx.CreateLogisticOrderItem(ctx, db.CreateLogisticOrderItemParams{
				LogisticOrderID: id,
				Name:            item.ItemCode,
				ItemCode:        item.ItemCode,
				Quantity:        item.Quantity,
			}); err != nil {
				return nil, err
			}
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}
	} else if len(setClauses) > 0 {
		query := fmt.Sprintf("UPDATE logistic_order SET %s WHERE logistic_order_id = ? AND deleted_at IS NULL",
			strings.Join(setClauses, ", "))
		args = append(args, id)
		if _, err := s.sqlDB.ExecContext(ctx, query, args...); err != nil {
			return nil, err
		}
	}

	_ = order
	return s.loadOrder(ctx, id)
}

// ── Assign ────────────────────────────────────────────────────────────────────

func (s *OrderService) Assign(ctx context.Context, id int32, userID int64) (*OrderResponse, error) {
	if _, err := s.q.GetLogisticOrderByID(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem logística não encontrada")
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

	if _, err := s.sqlDB.ExecContext(ctx,
		`UPDATE logistic_order SET assigned_to_id=?, assigned_to_name=?, captured_at=NOW() WHERE logistic_order_id=? AND deleted_at IS NULL`,
		int32(userID), name, id,
	); err != nil {
		return nil, err
	}
	return s.loadOrder(ctx, id)
}

// ── ChangeStatus ──────────────────────────────────────────────────────────────

func (s *OrderService) ChangeStatus(ctx context.Context, id int32, userID int64, newStatusID int32) (*OrderResponse, error) {
	order, err := s.q.GetLogisticOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem logística não encontrada")
	}
	if err != nil {
		return nil, err
	}

	// collect user's access level IDs
	levelRows, err := s.sqlDB.QueryContext(ctx, `
		SELECT access_level_id FROM users_access_levels WHERE user_id = ?
		UNION
		SELECT gal.access_level_id FROM groups_access_levels gal
		JOIN users_groups ug ON ug.group_id = gal.group_id
		WHERE ug.user_id = ?`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer levelRows.Close()
	var levelIDs []any
	for levelRows.Next() {
		var alID int32
		if err := levelRows.Scan(&alID); err == nil {
			levelIDs = append(levelIDs, alID)
		}
	}
	if len(levelIDs) == 0 {
		return nil, errors.New("transição de status não permitida para o seu nível de acesso")
	}

	placeholders := strings.Repeat("?,", len(levelIDs))
	placeholders = placeholders[:len(placeholders)-1]
	checkArgs := []any{newStatusID}
	checkArgs = append(checkArgs, levelIDs...)
	checkArgs = append(checkArgs, order.StatusID)

	var count int
	checkQ := fmt.Sprintf(`SELECT COUNT(*) FROM logistic_order_status_transition
		WHERE to_status_id = ? AND access_level_id IN (%s)
		AND (from_status_id = ? OR from_status_id IS NULL)`, placeholders)
	if err := s.sqlDB.QueryRowContext(ctx, checkQ, checkArgs...).Scan(&count); err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, errors.New("transição de status não permitida para o seu nível de acesso")
	}

	newStatus, err := s.q.GetLogisticOrderStatusByID(ctx, newStatusID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("status não encontrado")
	}
	if err != nil {
		return nil, err
	}

	capturedAt := order.CapturedAt
	closedAt := order.ClosedAt
	if newStatus.IsClosed && !order.StatusIsClosed {
		closedAt = sql.NullTime{Time: time.Now(), Valid: true}
	} else if !newStatus.IsClosed && order.StatusIsClosed {
		closedAt = sql.NullTime{}
	}

	if err := s.q.UpdateLogisticOrderStatus(ctx, db.UpdateLogisticOrderStatusParams{
		StatusID:        newStatusID,
		CapturedAt:      capturedAt,
		ClosedAt:        closedAt,
		LogisticOrderID: id,
	}); err != nil {
		return nil, err
	}
	return s.loadOrder(ctx, id)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (s *OrderService) Delete(ctx context.Context, id int32) error {
	if _, err := s.q.GetLogisticOrderByID(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return errors.New("ordem logística não encontrada")
	} else if err != nil {
		return err
	}
	return s.q.DeleteLogisticOrder(ctx, id)
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
	sortCol := "o.created_at"
	if p.SortBy == "scheduleDate" {
		sortCol = "o.schedule_date"
	}

	baseSelect := `
SELECT o.logistic_order_id, o.requester_id, o.requester_name,
    o.assigned_to_id, o.assigned_to_name,
    o.driver_name, o.checker_name, o.checker_signature, o.salesperson_name,
    o.client_code, o.client_name, o.client_alias_name,
    o.contact_type, o.contact_name, o.contact_phone, o.contact_mobile, o.contact_email,
    o.address_name, o.address_type, o.type_of_address, o.street, o.street_no, o.block,
    o.city, o.county, o.state, o.country, o.zip_code, o.building_floor_room,
    o.status_id, o.priority_id, o.reason_id, o.schedule_date, o.description,
    o.has_salary_discount, o.salary_discount_reason,
    o.captured_at, o.resolution, o.closed_at, o.created_at, o.updated_at,
    s.name, s.background_color, s.border_color, s.text_color, s.is_closed, s.is_rh_step,
    p.name, p.background_color,
    r.name`

	baseFrom := `
FROM logistic_order o
JOIN logistic_order_status s ON s.logistic_order_status_id = o.status_id
JOIN logistic_order_priorities p ON p.logistic_order_priority_id = o.priority_id
JOIN logistic_order_reason r ON r.logistic_order_reason_id = o.reason_id`

	where := []string{"o.deleted_at IS NULL"}
	args := []any{}

	if p.Search != "" {
		where = append(where, "(o.client_name LIKE ? OR o.requester_name LIKE ? OR o.assigned_to_name LIKE ? OR r.name LIKE ?)")
		like := "%" + p.Search + "%"
		args = append(args, like, like, like, like)
	}
	if p.StatusID != nil {
		where = append(where, "o.status_id = ?")
		args = append(args, *p.StatusID)
	}
	if p.PriorityID != nil {
		where = append(where, "o.priority_id = ?")
		args = append(args, *p.PriorityID)
	}
	if p.ReasonID != nil {
		where = append(where, "o.reason_id = ?")
		args = append(args, *p.ReasonID)
	}
	if p.AssignedToID != nil {
		where = append(where, "o.assigned_to_id = ?")
		args = append(args, *p.AssignedToID)
	}
	if p.DriverName != nil {
		where = append(where, "o.driver_name LIKE ?")
		args = append(args, "%"+*p.DriverName+"%")
	}
	if p.OnlyWithScheduleDate {
		where = append(where, "o.schedule_date IS NOT NULL")
	}
	if p.StartDate != nil {
		where = append(where, "o.created_at >= ?")
		args = append(args, *p.StartDate)
	}
	if p.EndDate != nil {
		where = append(where, "o.created_at <= ?")
		args = append(args, *p.EndDate)
	}

	whereClause := "WHERE " + strings.Join(where, " AND ")

	var total int64
	countQuery := "SELECT COUNT(*) " + baseFrom + " " + whereClause
	if err := s.sqlDB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := baseSelect + baseFrom + " " + whereClause +
		" ORDER BY " + sortCol + " " + sortDir +
		fmt.Sprintf(" LIMIT %d OFFSET %d", p.Limit, offset)

	rows, err := s.sqlDB.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []OrderResponse
	for rows.Next() {
		var o db.GetLogisticOrderByIDRow
		if err := rows.Scan(
			&o.LogisticOrderID, &o.RequesterID, &o.RequesterName,
			&o.AssignedToID, &o.AssignedToName,
			&o.DriverName, &o.CheckerName, &o.CheckerSignature, &o.SalespersonName,
			&o.ClientCode, &o.ClientName, &o.ClientAliasName,
			&o.ContactType, &o.ContactName, &o.ContactPhone, &o.ContactMobile, &o.ContactEmail,
			&o.AddressName, &o.AddressType, &o.TypeOfAddress, &o.Street, &o.StreetNo, &o.Block,
			&o.City, &o.County, &o.State, &o.Country, &o.ZipCode, &o.BuildingFloorRoom,
			&o.StatusID, &o.PriorityID, &o.ReasonID, &o.ScheduleDate, &o.Description,
			&o.HasSalaryDiscount, &o.SalaryDiscountReason,
			&o.CapturedAt, &o.Resolution, &o.ClosedAt, &o.CreatedAt, &o.UpdatedAt,
			&o.StatusName, &o.StatusBg, &o.StatusBorder, &o.StatusText, &o.StatusIsClosed, &o.StatusIsRhStep,
			&o.PriorityName, &o.PriorityBg,
			&o.ReasonName,
		); err != nil {
			return nil, 0, err
		}
		resp := rowToOrderResponse(o)
		resp.Items = []ItemResponse{}
		out = append(out, *resp)
	}
	if out == nil {
		out = []OrderResponse{}
	}
	return out, total, nil
}

// ── FindPendingApproval ───────────────────────────────────────────────────────

func (s *OrderService) FindPendingApproval(ctx context.Context, p FindPaginatedParams) ([]OrderResponse, int64, error) {
	statuses, err := s.q.ListLogisticOrderStatuses(ctx)
	if err != nil {
		return nil, 0, err
	}
	var pendingIDs []string
	var pendingArgs []any
	for _, st := range statuses {
		if st.IsPendingApproval {
			pendingIDs = append(pendingIDs, "?")
			pendingArgs = append(pendingArgs, st.LogisticOrderStatusID)
		}
	}
	if len(pendingIDs) == 0 {
		return []OrderResponse{}, 0, nil
	}

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

	baseSelect := `SELECT o.logistic_order_id, o.requester_id, o.requester_name,
    o.assigned_to_id, o.assigned_to_name,
    o.driver_name, o.checker_name, o.checker_signature, o.salesperson_name,
    o.client_code, o.client_name, o.client_alias_name,
    o.contact_type, o.contact_name, o.contact_phone, o.contact_mobile, o.contact_email,
    o.address_name, o.address_type, o.type_of_address, o.street, o.street_no, o.block,
    o.city, o.county, o.state, o.country, o.zip_code, o.building_floor_room,
    o.status_id, o.priority_id, o.reason_id, o.schedule_date, o.description,
    o.has_salary_discount, o.salary_discount_reason,
    o.captured_at, o.resolution, o.closed_at, o.created_at, o.updated_at,
    s.name, s.background_color, s.border_color, s.text_color, s.is_closed, s.is_rh_step,
    p.name, p.background_color, r.name`

	baseFrom := ` FROM logistic_order o
JOIN logistic_order_status s ON s.logistic_order_status_id = o.status_id
JOIN logistic_order_priorities p ON p.logistic_order_priority_id = o.priority_id
JOIN logistic_order_reason r ON r.logistic_order_reason_id = o.reason_id`

	where := []string{"o.deleted_at IS NULL", "o.status_id IN (" + strings.Join(pendingIDs, ",") + ")"}
	args := append([]any{}, pendingArgs...)

	if p.Search != "" {
		where = append(where, "(o.client_name LIKE ? OR o.requester_name LIKE ?)")
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
		" ORDER BY o.created_at " + sortDir +
		fmt.Sprintf(" LIMIT %d OFFSET %d", p.Limit, offset)

	rows, err := s.sqlDB.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []OrderResponse
	for rows.Next() {
		var o db.GetLogisticOrderByIDRow
		if err := rows.Scan(
			&o.LogisticOrderID, &o.RequesterID, &o.RequesterName,
			&o.AssignedToID, &o.AssignedToName,
			&o.DriverName, &o.CheckerName, &o.CheckerSignature, &o.SalespersonName,
			&o.ClientCode, &o.ClientName, &o.ClientAliasName,
			&o.ContactType, &o.ContactName, &o.ContactPhone, &o.ContactMobile, &o.ContactEmail,
			&o.AddressName, &o.AddressType, &o.TypeOfAddress, &o.Street, &o.StreetNo, &o.Block,
			&o.City, &o.County, &o.State, &o.Country, &o.ZipCode, &o.BuildingFloorRoom,
			&o.StatusID, &o.PriorityID, &o.ReasonID, &o.ScheduleDate, &o.Description,
			&o.HasSalaryDiscount, &o.SalaryDiscountReason,
			&o.CapturedAt, &o.Resolution, &o.ClosedAt, &o.CreatedAt, &o.UpdatedAt,
			&o.StatusName, &o.StatusBg, &o.StatusBorder, &o.StatusText, &o.StatusIsClosed, &o.StatusIsRhStep,
			&o.PriorityName, &o.PriorityBg, &o.ReasonName,
		); err != nil {
			return nil, 0, err
		}
		resp := rowToOrderResponse(o)
		resp.Items = []ItemResponse{}
		out = append(out, *resp)
	}
	if out == nil {
		out = []OrderResponse{}
	}
	return out, total, nil
}

// ── Converters ────────────────────────────────────────────────────────────────

func rowToOrderResponse(o db.GetLogisticOrderByIDRow) *OrderResponse {
	r := &OrderResponse{
		ID:              o.LogisticOrderID,
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
		Items:           []ItemResponse{},
		Status: StatusSummary{
			ID:              o.StatusID,
			Name:            o.StatusName,
			BackgroundColor: o.StatusBg,
			BorderColor:     o.StatusBorder,
			TextColor:       o.StatusText,
			IsClosed:        o.StatusIsClosed,
			IsRhStep:        o.StatusIsRhStep,
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
	if o.DriverName.Valid {
		r.DriverName = &o.DriverName.String
	}
	if o.CheckerName.Valid {
		r.CheckerName = &o.CheckerName.String
	}
	if o.CheckerSignature.Valid {
		r.CheckerSignature = &o.CheckerSignature.String
	}
	if o.SalespersonName.Valid {
		r.SalespersonName = &o.SalespersonName.String
	}
	if o.ContactType.Valid {
		r.ContactType = &o.ContactType.String
	}
	if o.ContactName.Valid {
		r.ContactName = &o.ContactName.String
	}
	if o.ContactPhone.Valid {
		r.ContactPhone = &o.ContactPhone.String
	}
	if o.ContactMobile.Valid {
		r.ContactMobile = &o.ContactMobile.String
	}
	if o.ContactEmail.Valid {
		r.ContactEmail = &o.ContactEmail.String
	}
	if o.AddressName.Valid {
		r.AddressName = &o.AddressName.String
	}
	if o.AddressType.Valid {
		r.AddressType = &o.AddressType.String
	}
	if o.TypeOfAddress.Valid {
		r.TypeOfAddress = &o.TypeOfAddress.String
	}
	if o.Street.Valid {
		r.Street = &o.Street.String
	}
	if o.StreetNo.Valid {
		r.StreetNo = &o.StreetNo.String
	}
	if o.Block.Valid {
		r.Block = &o.Block.String
	}
	if o.City.Valid {
		r.City = &o.City.String
	}
	if o.County.Valid {
		r.County = &o.County.String
	}
	if o.State.Valid {
		r.State = &o.State.String
	}
	if o.Country.Valid {
		r.Country = &o.Country.String
	}
	if o.ZipCode.Valid {
		r.ZipCode = &o.ZipCode.String
	}
	if o.BuildingFloorRoom.Valid {
		r.BuildingFloorRoom = &o.BuildingFloorRoom.String
	}
	if o.ScheduleDate.Valid {
		t := o.ScheduleDate.Time
		r.ScheduleDate = &t
	}
	if o.HasSalaryDiscount.Valid {
		v := o.HasSalaryDiscount.Bool
		r.HasSalaryDiscount = &v
	}
	if o.SalaryDiscountReason.Valid {
		r.SalaryDiscountReason = &o.SalaryDiscountReason.String
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
