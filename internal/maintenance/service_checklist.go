package maintenance

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"systemapi/internal/db"
)

type ChecklistService struct {
	sqlDB *sql.DB
	q     *db.Queries
}

func NewChecklistService(sqlDB *sql.DB, q *db.Queries) *ChecklistService {
	return &ChecklistService{sqlDB: sqlDB, q: q}
}

func (s *ChecklistService) GetByOrderID(ctx context.Context, orderID int32) ([]OrderChecklistResponse, error) {
	rows, err := s.q.ListOrderChecklists(ctx, orderID)
	if err != nil {
		return nil, err
	}
	out := make([]OrderChecklistResponse, len(rows))
	for i, r := range rows {
		cl, err := s.buildChecklistResp(ctx, r)
		if err != nil {
			return nil, err
		}
		out[i] = cl
	}
	return out, nil
}

func (s *ChecklistService) GetByID(ctx context.Context, orderID, checklistID int32) (*OrderChecklistResponse, error) {
	row, err := s.q.GetOrderChecklistByID(ctx, checklistID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("checklist não encontrado")
	}
	if err != nil {
		return nil, err
	}
	if row.MaintenanceOrderID != orderID {
		return nil, errors.New("checklist não pertence a esta ordem")
	}
	resp, err := s.buildChecklistResp(ctx, row)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *ChecklistService) Submit(ctx context.Context, orderID int32, req SubmitChecklistRequest) (*OrderChecklistResponse, error) {
	if _, err := s.q.GetMaintenanceOrderByID(ctx, orderID); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem de serviço não encontrada")
	} else if err != nil {
		return nil, err
	}

	tx, err := s.sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	q := s.q.WithTx(tx)

	res, err := q.CreateOrderChecklist(ctx, db.CreateOrderChecklistParams{
		MaintenanceOrderID: orderID,
		TemplateID:         req.TemplateID,
	})
	if err != nil {
		return nil, err
	}
	id64, _ := res.LastInsertId()
	checklistID := int32(id64)

	if req.CompletedAt != nil {
		t, err := time.Parse(time.RFC3339, *req.CompletedAt)
		if err != nil {
			t, _ = time.Parse("2006-01-02T15:04:05", *req.CompletedAt)
		}
		if err := q.UpdateOrderChecklistCompletedAt(ctx, db.UpdateOrderChecklistCompletedAtParams{
			CompletedAt:                 sql.NullTime{Time: t, Valid: true},
			MaintenanceOrderChecklistID: checklistID,
		}); err != nil {
			return nil, err
		}
	}

	for _, answer := range req.Answers {
		val := sql.NullString{}
		if answer.Value != nil {
			val = sql.NullString{String: *answer.Value, Valid: true}
		}
		if err := q.CreateChecklistAnswer(ctx, db.CreateChecklistAnswerParams{
			ChecklistID: checklistID,
			ItemID:      answer.ItemID,
			Value:       val,
		}); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, orderID, checklistID)
}

func (s *ChecklistService) Delete(ctx context.Context, orderID, checklistID int32) error {
	if _, err := s.GetByID(ctx, orderID, checklistID); err != nil {
		return err
	}
	return s.q.SoftDeleteOrderChecklist(ctx, checklistID)
}

func (s *ChecklistService) buildChecklistResp(ctx context.Context, row db.MaintenanceOrderChecklist) (OrderChecklistResponse, error) {
	answers, err := s.q.ListChecklistAnswers(ctx, row.MaintenanceOrderChecklistID)
	if err != nil {
		return OrderChecklistResponse{}, err
	}
	answerResps := make([]ChecklistAnswerResponse, len(answers))
	for i, a := range answers {
		ar := ChecklistAnswerResponse{
			ID:          a.MaintenanceOrderChecklistAnswerID,
			ChecklistID: a.ChecklistID,
			ItemID:      a.ItemID,
		}
		if a.Value.Valid {
			ar.Value = &a.Value.String
		}
		answerResps[i] = ar
	}

	resp := OrderChecklistResponse{
		ID:         row.MaintenanceOrderChecklistID,
		OrderID:    row.MaintenanceOrderID,
		TemplateID: row.TemplateID,
		Answers:    answerResps,
		CreatedAt:  row.CreatedAt,
	}
	if row.CompletedAt.Valid {
		resp.CompletedAt = &row.CompletedAt.Time
	}
	return resp, nil
}
