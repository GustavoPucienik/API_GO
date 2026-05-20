package maintenance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"systemapi/internal/db"
	mw "systemapi/internal/middleware"
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
	// Validate requester
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

	// Validate SAP items
	type sapItem struct{ name, code string }
	items := make([]sapItem, len(req.Items))
	for i, item := range req.Items {
		name, err := s.getSapItemName(ctx, item.ItemCode)
		if err != nil {
			return nil, fmt.Errorf("item %s não encontrado", item.ItemCode)
		}
		items[i] = sapItem{name: name, code: item.ItemCode}
	}

	// Resolve status
	var statusID int32
	if req.StatusID != nil {
		statusID = *req.StatusID
	} else {
		st, err := s.q.GetDefaultMaintenanceOrderStatus(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			st, err = s.q.GetFirstMaintenanceOrderStatusBySort(ctx)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("nenhum status configurado para manutenção")
		}
		if err != nil {
			return nil, err
		}
		statusID = st.MaintenanceOrderStatusID
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	q := s.q.WithTx(tx)

	res, err := q.CreateMaintenanceOrder(ctx, db.CreateMaintenanceOrderParams{
		RequesterID:       int32(req.RequesterID),
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
		Description:       req.Description,
	})
	if err != nil {
		return nil, err
	}
	orderID64, _ := res.LastInsertId()
	orderID := int32(orderID64)

	// Insert items
	for i, item := range req.Items {
		if err := q.CreateMaintenanceOrderItem(ctx, db.CreateMaintenanceOrderItemParams{
			MaintenanceOrderID: orderID,
			Name:               items[i].name,
			ItemCode:           item.ItemCode,
			Quantity:           item.Quantity,
		}); err != nil {
			return nil, err
		}
	}

	// Insert schedule dates + technicians
	for _, sd := range req.ScheduleDates {
		sdRes, err := q.CreateMaintenanceOrderScheduleDate(ctx, db.CreateMaintenanceOrderScheduleDateParams{
			MaintenanceOrderID: orderID,
			Date:               sd.Date,
		})
		if err != nil {
			return nil, err
		}
		sdID64, _ := sdRes.LastInsertId()
		sdID := int32(sdID64)

		for _, t := range sd.Technicians {
			if err := q.CreateScheduleDateTechnician(ctx, db.CreateScheduleDateTechnicianParams{
				MaintenanceOrderScheduleDateID: sdID,
				TechnicianID:                  t.TechnicianID,
				TechnicianName:                t.TechnicianName,
			}); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.loadOrder(ctx, orderID)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *OrderService) Update(ctx context.Context, id int32, req UpdateOrderRequest) (*OrderResponse, error) {
	order, err := s.q.GetMaintenanceOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem de serviço não encontrada")
	}
	if err != nil {
		return nil, err
	}

	// Pre-validate SAP items outside transaction
	type sapItem struct{ name, code string }
	var newItems []sapItem
	if req.Items != nil {
		newItems = make([]sapItem, len(*req.Items))
		for i, item := range *req.Items {
			name, err := s.getSapItemName(ctx, item.ItemCode)
			if err != nil {
				return nil, fmt.Errorf("item %s não encontrado", item.ItemCode)
			}
			newItems[i] = sapItem{name: name, code: item.ItemCode}
		}
	}
	var newResItems []sapItem
	if req.ResolutionItems != nil {
		newResItems = make([]sapItem, len(*req.ResolutionItems))
		for i, item := range *req.ResolutionItems {
			name, err := s.getSapItemName(ctx, item.ItemCode)
			if err != nil {
				return nil, fmt.Errorf("item de resolução %s não encontrado", item.ItemCode)
			}
			newResItems[i] = sapItem{name: name, code: item.ItemCode}
		}
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	q := s.q.WithTx(tx)

	// Build the UPDATE for the main order row
	setClauses := []string{"updated_at = NOW()"}
	args := []any{}

	apply := func(col string, val any) {
		setClauses = append(setClauses, col+" = ?")
		args = append(args, val)
	}

	if req.StatusID != nil {
		apply("status_id", *req.StatusID)
		// Handle closedAt based on new status
		st, err := s.q.GetMaintenanceOrderStatusByID(ctx, *req.StatusID)
		if err == nil {
			prevSt, _ := s.q.GetMaintenanceOrderStatusByID(ctx, order.StatusID)
			prevClosed := prevSt.IsClosed
			if st.IsClosed && !prevClosed {
				apply("closed_at", time.Now())
			} else if !st.IsClosed && prevClosed {
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
	if req.Resolution != nil {
		apply("resolution", *req.Resolution)
	}
	if req.MotivoResolution != nil {
		apply("motivo_resolution", *req.MotivoResolution)
	}
	if req.ClosedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ClosedAt)
		if err != nil {
			t, _ = time.Parse("2006-01-02T15:04:05", *req.ClosedAt)
		}
		apply("closed_at", t)
	}

	args = append(args, id)
	query := "UPDATE maintenance_order SET " + strings.Join(setClauses, ", ") +
		" WHERE maintenance_order_id = ? AND deleted_at IS NULL"
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}

	// Items
	if req.Items != nil {
		if err := q.DeleteMaintenanceOrderItems(ctx, id); err != nil {
			return nil, err
		}
		for i, item := range *req.Items {
			if err := q.CreateMaintenanceOrderItem(ctx, db.CreateMaintenanceOrderItemParams{
				MaintenanceOrderID: id,
				Name:               newItems[i].name,
				ItemCode:           item.ItemCode,
				Quantity:           item.Quantity,
			}); err != nil {
				return nil, err
			}
		}
	}

	// Resolution items
	if req.ResolutionItems != nil {
		if err := q.DeleteMaintenanceOrderResolutionItems(ctx, id); err != nil {
			return nil, err
		}
		for i, item := range *req.ResolutionItems {
			if err := q.CreateMaintenanceOrderResolutionItem(ctx, db.CreateMaintenanceOrderResolutionItemParams{
				MaintenanceOrderID: id,
				Name:               newResItems[i].name,
				ItemCode:           item.ItemCode,
				Quantity:           item.Quantity,
			}); err != nil {
				return nil, err
			}
		}
	}

	// Time entries
	if req.TimeEntries != nil {
		if err := q.DeleteMaintenanceOrderTimeEntries(ctx, id); err != nil {
			return nil, err
		}
		for _, e := range *req.TimeEntries {
			started, err := time.Parse(time.RFC3339, e.StartedAt)
			if err != nil {
				started, _ = time.Parse("2006-01-02T15:04:05", e.StartedAt)
			}
			endedAt := sql.NullTime{}
			if e.EndedAt != nil {
				t, err := time.Parse(time.RFC3339, *e.EndedAt)
				if err != nil {
					t, _ = time.Parse("2006-01-02T15:04:05", *e.EndedAt)
				}
				endedAt = sql.NullTime{Time: t, Valid: true}
			}
			if err := q.CreateMaintenanceOrderTimeEntry(ctx, db.CreateMaintenanceOrderTimeEntryParams{
				MaintenanceOrderID: id,
				TypeID:             e.TypeID,
				StartedAt:          started,
				EndedAt:            endedAt,
			}); err != nil {
				return nil, err
			}
		}
	}

	// Schedule dates
	if req.ScheduleDates != nil {
		// Delete technicians for each existing schedule date
		existingDates, err := q.ListMaintenanceOrderScheduleDates(ctx, id)
		if err != nil {
			return nil, err
		}
		for _, d := range existingDates {
			if err := q.DeleteScheduleDateTechnicians(ctx, d.MaintenanceOrderScheduleDateID); err != nil {
				return nil, err
			}
		}
		if err := q.DeleteMaintenanceOrderScheduleDates(ctx, id); err != nil {
			return nil, err
		}

		for _, sd := range *req.ScheduleDates {
			sdRes, err := q.CreateMaintenanceOrderScheduleDate(ctx, db.CreateMaintenanceOrderScheduleDateParams{
				MaintenanceOrderID: id,
				Date:               sd.Date,
			})
			if err != nil {
				return nil, err
			}
			sdID64, _ := sdRes.LastInsertId()
			sdID := int32(sdID64)

			for _, t := range sd.Technicians {
				if err := q.CreateScheduleDateTechnician(ctx, db.CreateScheduleDateTechnicianParams{
					MaintenanceOrderScheduleDateID: sdID,
					TechnicianID:                  t.TechnicianID,
					TechnicianName:                t.TechnicianName,
				}); err != nil {
					return nil, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.loadOrder(ctx, id)
}

// ── FindPaginated ─────────────────────────────────────────────────────────────

type FindPaginatedParams struct {
	Page                int
	Limit               int
	Search              string
	SortBy              string
	SortOrder           string
	StatusID            *int32
	PriorityID          *int32
	ReasonID            *int32
	ClientCode          *string
	ResponsibleID       *int64
	StartDate           *string
	EndDate             *string
	OnlyWithScheduleDate bool
	Scope               *mw.ResolvedScope
}

func (s *OrderService) FindPaginated(ctx context.Context, p FindPaginatedParams) ([]OrderResponse, int64, error) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 10
	}
	offset := (p.Page - 1) * p.Limit

	allowedSort := map[string]string{
		"id":        "mo.maintenance_order_id",
		"createdAt": "mo.created_at",
	}
	sortField, ok := allowedSort[p.SortBy]
	if !ok {
		sortField = "mo.created_at"
	}
	sortDir := "DESC"
	if strings.ToUpper(p.SortOrder) == "ASC" {
		sortDir = "ASC"
	}

	baseSelect := `
SELECT
    mo.maintenance_order_id, mo.requester_id, mo.requester_name, mo.salesperson_name,
    mo.client_code, mo.client_name, mo.client_alias_name,
    mo.contact_type, mo.contact_name, mo.contact_phone, mo.contact_mobile, mo.contact_email,
    mo.address_name, mo.address_type, mo.type_of_address, mo.street, mo.street_no, mo.block,
    mo.city, mo.county, mo.state, mo.country, mo.zip_code, mo.building_floor_room,
    mo.status_id, mo.priority_id, mo.reason_id,
    mo.description, mo.captured_at, mo.resolution, mo.motivo_resolution, mo.client_signature, mo.closed_at,
    mo.created_at, mo.updated_at,
    s.maintenance_order_status_id, s.name AS status_name, s.background_color AS status_bg,
    s.border_color AS status_border, s.text_color AS status_text, s.sort_index AS status_sort,
    s.is_closed, s.is_default,
    p.maintenance_order_priorities_id, p.name AS priority_name,
    p.background_color AS priority_bg, p.border_color AS priority_border,
    p.text_color AS priority_text, p.sort_index AS priority_sort,
    r.maintenance_order_reason_id, r.name AS reason_name`

	baseFrom := `
FROM maintenance_order mo
JOIN maintenance_order_status s ON s.maintenance_order_status_id = mo.status_id
JOIN maintenance_order_priorities p ON p.maintenance_order_priorities_id = mo.priority_id
JOIN maintenance_order_reason r ON r.maintenance_order_reason_id = mo.reason_id`

	where := []string{"mo.deleted_at IS NULL"}
	args := []any{}

	if p.Search != "" {
		where = append(where, `(mo.client_name LIKE ? OR mo.requester_name LIKE ? OR r.name LIKE ?
          OR EXISTS (
            SELECT 1 FROM maintenance_order_schedule_date_technician mot
            INNER JOIN maintenance_order_schedule_date msd
              ON msd.maintenance_order_schedule_date_id = mot.maintenance_order_schedule_date_id
            WHERE msd.maintenance_order_id = mo.maintenance_order_id
              AND mot.technician_name LIKE ?
              AND mot.deleted_at IS NULL AND msd.deleted_at IS NULL
          ))`)
		like := "%" + p.Search + "%"
		args = append(args, like, like, like, like)
	}

	if p.StatusID != nil {
		where = append(where, "mo.status_id = ?")
		args = append(args, *p.StatusID)
	}
	if p.PriorityID != nil {
		where = append(where, "mo.priority_id = ?")
		args = append(args, *p.PriorityID)
	}
	if p.ReasonID != nil {
		where = append(where, "mo.reason_id = ?")
		args = append(args, *p.ReasonID)
	}
	if p.ClientCode != nil {
		where = append(where, "mo.client_code = ?")
		args = append(args, *p.ClientCode)
	}

	// responsibleId + date filters
	if p.ResponsibleID != nil && (p.StartDate != nil || p.EndDate != nil || p.OnlyWithScheduleDate) {
		sub := "EXISTS (SELECT 1 FROM maintenance_order_schedule_date sd JOIN maintenance_order_schedule_date_technician sdt ON sdt.maintenance_order_schedule_date_id = sd.maintenance_order_schedule_date_id WHERE sd.maintenance_order_id = mo.maintenance_order_id AND sdt.technician_id = ? AND sd.deleted_at IS NULL AND sdt.deleted_at IS NULL"
		args = append(args, *p.ResponsibleID)
		if p.StartDate != nil {
			sub += " AND sd.date >= ?"
			args = append(args, *p.StartDate)
		}
		if p.EndDate != nil {
			sub += " AND sd.date <= ?"
			args = append(args, *p.EndDate)
		}
		sub += ")"
		where = append(where, sub)
	} else if p.ResponsibleID != nil {
		where = append(where, "EXISTS (SELECT 1 FROM maintenance_order_schedule_date sd JOIN maintenance_order_schedule_date_technician sdt ON sdt.maintenance_order_schedule_date_id = sd.maintenance_order_schedule_date_id WHERE sd.maintenance_order_id = mo.maintenance_order_id AND sdt.technician_id = ? AND sd.deleted_at IS NULL AND sdt.deleted_at IS NULL)")
		args = append(args, *p.ResponsibleID)
	} else if p.StartDate != nil || p.EndDate != nil || p.OnlyWithScheduleDate {
		sub := "EXISTS (SELECT 1 FROM maintenance_order_schedule_date sd WHERE sd.maintenance_order_id = mo.maintenance_order_id AND sd.deleted_at IS NULL"
		if p.StartDate != nil {
			sub += " AND sd.date >= ?"
			args = append(args, *p.StartDate)
		}
		if p.EndDate != nil {
			sub += " AND sd.date <= ?"
			args = append(args, *p.EndDate)
		}
		sub += ")"
		where = append(where, sub)
	}

	// Scope filter
	if p.Scope != nil {
		switch p.Scope.Type {
		case mw.ScopeOWN:
			where = append(where, "mo.requester_id = ?")
			args = append(args, p.Scope.UserID)
		case mw.ScopeRESPONSIBLE:
			where = append(where, "EXISTS (SELECT 1 FROM maintenance_order_schedule_date sd JOIN maintenance_order_schedule_date_technician sdt ON sdt.maintenance_order_schedule_date_id = sd.maintenance_order_schedule_date_id WHERE sd.maintenance_order_id = mo.maintenance_order_id AND sdt.technician_id = ? AND sd.deleted_at IS NULL AND sdt.deleted_at IS NULL)")
			args = append(args, p.Scope.UserID)
		case mw.ScopeGROUP:
			if len(p.Scope.GroupIDs) == 0 {
				where = append(where, "mo.requester_id = ?")
				args = append(args, p.Scope.UserID)
			} else {
				placeholders := make([]string, len(p.Scope.GroupIDs))
				groupArgs := make([]any, len(p.Scope.GroupIDs))
				for i, gid := range p.Scope.GroupIDs {
					placeholders[i] = "?"
					groupArgs[i] = gid
				}
				where = append(where, fmt.Sprintf("(mo.requester_id = ? OR mo.requester_id IN (SELECT ug.user_id FROM users_groups ug WHERE ug.group_id IN (%s)))", strings.Join(placeholders, ",")))
				args = append(args, p.Scope.UserID)
				args = append(args, groupArgs...)
			}
		}
	}

	whereSQL := " WHERE " + strings.Join(where, " AND ")

	// Count query
	countQuery := "SELECT COUNT(*) FROM maintenance_order mo JOIN maintenance_order_status s ON s.maintenance_order_status_id = mo.status_id JOIN maintenance_order_priorities p ON p.maintenance_order_priorities_id = mo.priority_id JOIN maintenance_order_reason r ON r.maintenance_order_reason_id = mo.reason_id" + whereSQL
	var total int64
	if err := s.sqlDB.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Data query
	dataQuery := baseSelect + baseFrom + whereSQL +
		" ORDER BY " + sortField + " " + sortDir +
		" LIMIT ? OFFSET ?"
	dataArgs := append(args, p.Limit, offset)

	rows, err := s.sqlDB.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []OrderResponse
	var orderIDs []int32
	for rows.Next() {
		var o OrderResponse
		var salespersonName, contactType, contactName, contactPhone, contactMobile, contactEmail sql.NullString
		var addressName, addressType, typeOfAddress, street, streetNo, block, city, county, state, country, zipCode, buildingFloorRoom sql.NullString
		var resolution, motivoResolution, clientSignature sql.NullString
		var capturedAt, closedAt sql.NullTime

		if err := rows.Scan(
			&o.ID, &o.RequesterID, &o.RequesterName, &salespersonName,
			&o.ClientCode, &o.ClientName, &o.AliasName,
			&contactType, &contactName, &contactPhone, &contactMobile, &contactEmail,
			&addressName, &addressType, &typeOfAddress, &street, &streetNo, &block,
			&city, &county, &state, &country, &zipCode, &buildingFloorRoom,
			&o.Status.ID, &o.Priority.ID, &o.Reason.ID,
			&o.Description, &capturedAt, &resolution, &motivoResolution, &clientSignature, &closedAt,
			&o.CreatedAt, &o.UpdatedAt,
			&o.Status.ID, &o.Status.Name, &o.Status.BackgroundColor, &o.Status.BorderColor,
			&o.Status.TextColor, &o.Status.SortIndex, &o.Status.IsClosed, &o.Status.IsDefault,
			&o.Priority.ID, &o.Priority.Name, &o.Priority.BackgroundColor, &o.Priority.BorderColor,
			&o.Priority.TextColor, &o.Priority.SortIndex,
			&o.Reason.ID, &o.Reason.Name,
		); err != nil {
			return nil, 0, err
		}

		o.SalespersonName = nullStrPtr(salespersonName)
		o.ContactType = nullStrPtr(contactType)
		o.ContactName = nullStrPtr(contactName)
		o.ContactPhone = nullStrPtr(contactPhone)
		o.ContactMobile = nullStrPtr(contactMobile)
		o.ContactEmail = nullStrPtr(contactEmail)
		o.AddressName = nullStrPtr(addressName)
		o.AddressType = nullStrPtr(addressType)
		o.TypeOfAddress = nullStrPtr(typeOfAddress)
		o.Street = nullStrPtr(street)
		o.StreetNo = nullStrPtr(streetNo)
		o.Block = nullStrPtr(block)
		o.City = nullStrPtr(city)
		o.County = nullStrPtr(county)
		o.State = nullStrPtr(state)
		o.Country = nullStrPtr(country)
		o.ZipCode = nullStrPtr(zipCode)
		o.BuildingFloorRoom = nullStrPtr(buildingFloorRoom)
		o.Resolution = nullStrPtr(resolution)
		o.MotivoResolution = nullStrPtr(motivoResolution)
		o.ClientSignature = nullStrPtr(clientSignature)
		if capturedAt.Valid {
			o.CapturedAt = &capturedAt.Time
		}
		if closedAt.Valid {
			o.ClosedAt = &closedAt.Time
		}
		o.Items = []OrderItemResponse{}
		o.ResolutionItems = []OrderItemResponse{}
		o.TimeEntries = []TimeEntryResponse{}
		o.ScheduleDates = []ScheduleDateResponse{}

		orders = append(orders, o)
		orderIDs = append(orderIDs, o.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	if len(orders) == 0 {
		return []OrderResponse{}, total, nil
	}

	// Batch-load items and schedule dates
	if err := s.attachRelationsToPage(ctx, orders, orderIDs); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (s *OrderService) attachRelationsToPage(ctx context.Context, orders []OrderResponse, ids []int32) error {
	// Items
	itemMap, err := s.batchLoadItems(ctx, ids)
	if err != nil {
		return err
	}
	// Schedule dates + technicians
	dateMap, err := s.batchLoadScheduleDates(ctx, ids)
	if err != nil {
		return err
	}
	for i := range orders {
		orders[i].Items = itemMap[orders[i].ID]
		orders[i].ScheduleDates = dateMap[orders[i].ID]
	}
	return nil
}

func (s *OrderService) batchLoadItems(ctx context.Context, ids []int32) (map[int32][]OrderItemResponse, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	rows, err := s.sqlDB.QueryContext(ctx,
		"SELECT maintenance_order_item_id, maintenance_order_id, name, item_code, quantity FROM maintenance_order_items WHERE maintenance_order_id IN ("+strings.Join(placeholders, ",")+") AND deleted_at IS NULL",
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[int32][]OrderItemResponse{}
	for rows.Next() {
		var item OrderItemResponse
		var orderID int32
		if err := rows.Scan(&item.ID, &orderID, &item.Name, &item.ItemCode, &item.Quantity); err != nil {
			return nil, err
		}
		result[orderID] = append(result[orderID], item)
	}
	return result, rows.Err()
}

func (s *OrderService) batchLoadScheduleDates(ctx context.Context, ids []int32) (map[int32][]ScheduleDateResponse, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	inClause := strings.Join(placeholders, ",")
	rows, err := s.sqlDB.QueryContext(ctx,
		"SELECT maintenance_order_schedule_date_id, maintenance_order_id, date FROM maintenance_order_schedule_date WHERE maintenance_order_id IN ("+inClause+") AND deleted_at IS NULL ORDER BY date ASC",
		args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dateMap := map[int32][]ScheduleDateResponse{}
	var dateIDs []int32
	dateByID := map[int32]*ScheduleDateResponse{}

	for rows.Next() {
		var sd ScheduleDateResponse
		var orderID int32
		if err := rows.Scan(&sd.ID, &orderID, &sd.Date); err != nil {
			return nil, err
		}
		sd.Technicians = []TechnicianResponse{}
		dateMap[orderID] = append(dateMap[orderID], sd)
		dateIDs = append(dateIDs, sd.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Load technicians for all date IDs
	if len(dateIDs) > 0 {
		techPlaceholders := make([]string, len(dateIDs))
		techArgs := make([]any, len(dateIDs))
		for i, id := range dateIDs {
			techPlaceholders[i] = "?"
			techArgs[i] = id
		}
		techRows, err := s.sqlDB.QueryContext(ctx,
			"SELECT maintenance_order_schedule_date_id, technician_id, technician_name FROM maintenance_order_schedule_date_technician WHERE maintenance_order_schedule_date_id IN ("+strings.Join(techPlaceholders, ",")+") AND deleted_at IS NULL",
			techArgs...)
		if err != nil {
			return nil, err
		}
		defer techRows.Close()

		// Index dates for fast lookup
		for oid, dList := range dateMap {
			for i := range dList {
				dateByID[dList[i].ID] = &dateMap[oid][i]
			}
		}

		for techRows.Next() {
			var sdID, techID int32
			var techName string
			if err := techRows.Scan(&sdID, &techID, &techName); err != nil {
				return nil, err
			}
			if sd, ok := dateByID[sdID]; ok {
				sd.Technicians = append(sd.Technicians, TechnicianResponse{TechnicianID: techID, TechnicianName: techName})
			}
		}
		if err := techRows.Err(); err != nil {
			return nil, err
		}
	}

	return dateMap, nil
}

// ── FindByID ──────────────────────────────────────────────────────────────────

func (s *OrderService) FindByID(ctx context.Context, id int32, scope *mw.ResolvedScope) (*OrderResponse, error) {
	order, err := s.loadOrder(ctx, id)
	if err != nil {
		return nil, err
	}

	if scope != nil {
		switch scope.Type {
		case mw.ScopeOWN:
			if int64(order.RequesterID) != scope.UserID {
				return nil, errors.New("sem permissão para acessar esta ordem")
			}
		case mw.ScopeGROUP:
			if int64(order.RequesterID) != scope.UserID && len(scope.GroupIDs) > 0 {
				allowed, err := s.isUserInGroups(ctx, int64(order.RequesterID), scope.GroupIDs)
				if err != nil {
					return nil, err
				}
				if !allowed {
					return nil, errors.New("sem permissão para acessar esta ordem")
				}
			}
		}
	}

	return order, nil
}

// ── SaveSignature ─────────────────────────────────────────────────────────────

func (s *OrderService) SaveSignature(ctx context.Context, id int32, sig string) error {
	if _, err := s.q.GetMaintenanceOrderByID(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return errors.New("ordem de serviço não encontrada")
	} else if err != nil {
		return err
	}
	return s.q.UpdateMaintenanceOrderSignature(ctx, db.UpdateMaintenanceOrderSignatureParams{
		ClientSignature: sql.NullString{String: sig, Valid: true},
		MaintenanceOrderID: id,
	})
}

// ── Assign ────────────────────────────────────────────────────────────────────

func (s *OrderService) Assign(ctx context.Context, id int32, userID int64) (*OrderResponse, error) {
	order, err := s.q.GetMaintenanceOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem de serviço não encontrada")
	}
	if err != nil {
		return nil, err
	}

	user, err := s.q.GetUserByID(ctx, int32(userID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("usuário não encontrado")
	}
	if err != nil {
		return nil, err
	}

	technicianName := user.Name
	if user.LastName.Valid {
		technicianName += " " + user.LastName.String
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	q := s.q.WithTx(tx)

	dates, err := q.ListMaintenanceOrderScheduleDates(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, d := range dates {
		count, err := q.CountTechnicianOnDate(ctx, db.CountTechnicianOnDateParams{
			MaintenanceOrderScheduleDateID: d.MaintenanceOrderScheduleDateID,
			TechnicianID:                   int32(userID),
		})
		if err != nil {
			return nil, err
		}
		if count == 0 {
			if err := q.CreateScheduleDateTechnician(ctx, db.CreateScheduleDateTechnicianParams{
				MaintenanceOrderScheduleDateID: d.MaintenanceOrderScheduleDateID,
				TechnicianID:                   int32(userID),
				TechnicianName:                 technicianName,
			}); err != nil {
				return nil, err
			}
		}
	}

	if !order.CapturedAt.Valid {
		if err := q.UpdateMaintenanceOrderCapturedAt(ctx, db.UpdateMaintenanceOrderCapturedAtParams{
			CapturedAt:         sql.NullTime{Time: time.Now(), Valid: true},
			MaintenanceOrderID: id,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.loadOrder(ctx, id)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (s *OrderService) Delete(ctx context.Context, id int32) error {
	if _, err := s.q.GetMaintenanceOrderByID(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return errors.New("ordem de serviço não encontrada")
	} else if err != nil {
		return err
	}
	n, err := s.q.SoftDeleteMaintenanceOrder(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("não foi possível excluir a ordem de serviço")
	}
	return nil
}

// ── ChangeStatus ──────────────────────────────────────────────────────────────

func (s *OrderService) ChangeStatus(ctx context.Context, id int32, statusID int32) (*OrderResponse, error) {
	order, err := s.q.GetMaintenanceOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem de serviço não encontrada")
	}
	if err != nil {
		return nil, err
	}

	newStatus, err := s.q.GetMaintenanceOrderStatusByID(ctx, statusID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("status não encontrado")
	}
	if err != nil {
		return nil, err
	}

	prevStatus, _ := s.q.GetMaintenanceOrderStatusByID(ctx, order.StatusID)
	wasClosedBefore := prevStatus.IsClosed

	setClauses := []string{"status_id = ?", "updated_at = NOW()"}
	args := []any{statusID}

	if newStatus.IsClosed && !wasClosedBefore {
		setClauses = append(setClauses, "closed_at = NOW()")
	} else if !newStatus.IsClosed && wasClosedBefore {
		setClauses = append(setClauses, "closed_at = NULL")
	}

	args = append(args, id)
	query := "UPDATE maintenance_order SET " + strings.Join(setClauses, ", ") +
		" WHERE maintenance_order_id = ? AND deleted_at IS NULL"
	if _, err := s.sqlDB.ExecContext(ctx, query, args...); err != nil {
		return nil, err
	}

	return s.loadOrder(ctx, id)
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (s *OrderService) loadOrder(ctx context.Context, id int32) (*OrderResponse, error) {
	order, err := s.q.GetMaintenanceOrderByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem de serviço não encontrada")
	}
	if err != nil {
		return nil, err
	}

	status, err := s.q.GetMaintenanceOrderStatusByID(ctx, order.StatusID)
	if err != nil {
		return nil, err
	}
	priority, err := s.q.GetMaintenanceOrderPriorityByID(ctx, order.PriorityID)
	if err != nil {
		return nil, err
	}
	reason, err := s.q.GetMaintenanceOrderReasonByID(ctx, order.ReasonID)
	if err != nil {
		return nil, err
	}

	dbItems, _ := s.q.ListMaintenanceOrderItems(ctx, id)
	dbResItems, _ := s.q.ListMaintenanceOrderResolutionItems(ctx, id)
	dbTimeEntries, _ := s.q.ListMaintenanceOrderTimeEntries(ctx, id)
	dbDates, _ := s.q.ListMaintenanceOrderScheduleDates(ctx, id)

	items := make([]OrderItemResponse, len(dbItems))
	for i, it := range dbItems {
		items[i] = OrderItemResponse{ID: it.MaintenanceOrderItemID, Name: it.Name, ItemCode: it.ItemCode, Quantity: it.Quantity}
	}
	resItems := make([]OrderItemResponse, len(dbResItems))
	for i, it := range dbResItems {
		resItems[i] = OrderItemResponse{ID: it.MaintenanceOrderResolutionItemID, Name: it.Name, ItemCode: it.ItemCode, Quantity: it.Quantity}
	}
	timeEntries := make([]TimeEntryResponse, len(dbTimeEntries))
	for i, te := range dbTimeEntries {
		tr := TimeEntryResponse{
			ID:        te.MaintenanceOrderTimeEntryID,
			TypeID:    te.TypeID,
			TypeName:  te.TypeName,
			StartedAt: te.StartedAt,
		}
		if te.TypeColor.Valid {
			tr.TypeColor = &te.TypeColor.String
		}
		if te.EndedAt.Valid {
			tr.EndedAt = &te.EndedAt.Time
		}
		timeEntries[i] = tr
	}

	dates := make([]ScheduleDateResponse, len(dbDates))
	for i, d := range dbDates {
		techs, _ := s.q.ListScheduleDateTechnicians(ctx, d.MaintenanceOrderScheduleDateID)
		techResps := make([]TechnicianResponse, len(techs))
		for j, t := range techs {
			techResps[j] = TechnicianResponse{TechnicianID: t.TechnicianID, TechnicianName: t.TechnicianName}
		}
		dates[i] = ScheduleDateResponse{ID: d.MaintenanceOrderScheduleDateID, Date: d.Date, Technicians: techResps}
	}

	resp := &OrderResponse{
		ID:            order.MaintenanceOrderID,
		RequesterID:   order.RequesterID,
		RequesterName: order.RequesterName,
		ClientCode:    order.ClientCode,
		ClientName:    order.ClientName,
		AliasName:     order.ClientAliasName,
		Description:   order.Description,
		Status:        statusToResp(status),
		Priority:      priorityToResp(priority),
		Reason:        ReasonResponse{ID: reason.MaintenanceOrderReasonID, Name: reason.Name},
		Items:         items,
		ResolutionItems: resItems,
		TimeEntries:   timeEntries,
		ScheduleDates: dates,
		CreatedAt:     order.CreatedAt,
		UpdatedAt:     order.UpdatedAt,
	}
	resp.SalespersonName = nullStrPtr(order.SalespersonName)
	resp.ContactType = nullStrPtr(order.ContactType)
	resp.ContactName = nullStrPtr(order.ContactName)
	resp.ContactPhone = nullStrPtr(order.ContactPhone)
	resp.ContactMobile = nullStrPtr(order.ContactMobile)
	resp.ContactEmail = nullStrPtr(order.ContactEmail)
	resp.AddressName = nullStrPtr(order.AddressName)
	resp.AddressType = nullStrPtr(order.AddressType)
	resp.TypeOfAddress = nullStrPtr(order.TypeOfAddress)
	resp.Street = nullStrPtr(order.Street)
	resp.StreetNo = nullStrPtr(order.StreetNo)
	resp.Block = nullStrPtr(order.Block)
	resp.City = nullStrPtr(order.City)
	resp.County = nullStrPtr(order.County)
	resp.State = nullStrPtr(order.State)
	resp.Country = nullStrPtr(order.Country)
	resp.ZipCode = nullStrPtr(order.ZipCode)
	resp.BuildingFloorRoom = nullStrPtr(order.BuildingFloorRoom)
	resp.Resolution = nullStrPtr(order.Resolution)
	resp.MotivoResolution = nullStrPtr(order.MotivoResolution)
	resp.ClientSignature = nullStrPtr(order.ClientSignature)
	if order.CapturedAt.Valid {
		resp.CapturedAt = &order.CapturedAt.Time
	}
	if order.ClosedAt.Valid {
		resp.ClosedAt = &order.ClosedAt.Time
	}
	return resp, nil
}

func (s *OrderService) getSapItemName(ctx context.Context, itemCode string) (string, error) {
	var name string
	err := s.sqlDB.QueryRowContext(ctx, "SELECT item_name FROM sap_items WHERE item_code = ? LIMIT 1", itemCode).Scan(&name)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("item %s não encontrado", itemCode)
	}
	return name, err
}

func (s *OrderService) isUserInGroups(ctx context.Context, userID int64, groupIDs []int64) (bool, error) {
	if len(groupIDs) == 0 {
		return false, nil
	}
	placeholders := make([]string, len(groupIDs))
	args := make([]any, len(groupIDs)+1)
	args[0] = userID
	for i, gid := range groupIDs {
		placeholders[i] = "?"
		args[i+1] = gid
	}
	var count int64
	err := s.sqlDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users_groups WHERE user_id = ? AND group_id IN ("+strings.Join(placeholders, ",")+")",
		args...).Scan(&count)
	return count > 0, err
}

// ── Null helpers ──────────────────────────────────────────────────────────────

func nullStr(p *string) sql.NullString {
	if p == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *p, Valid: true}
}

func nullStrPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}
