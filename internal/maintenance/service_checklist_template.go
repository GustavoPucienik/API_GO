package maintenance

import (
	"context"
	"database/sql"
	"errors"

	"systemapi/internal/db"
)

type ChecklistTemplateService struct {
	q *db.Queries
}

func NewChecklistTemplateService(q *db.Queries) *ChecklistTemplateService {
	return &ChecklistTemplateService{q: q}
}

func (s *ChecklistTemplateService) GetAll(ctx context.Context, reasonID *int32) ([]ChecklistTemplateResponse, error) {
	var rows []db.MaintenanceOrderChecklistTemplate
	var err error
	if reasonID != nil {
		rows, err = s.q.ListChecklistTemplatesByReason(ctx, sql.NullInt32{Int32: *reasonID, Valid: true})
	} else {
		rows, err = s.q.ListChecklistTemplates(ctx)
	}
	if err != nil {
		return nil, err
	}
	out := make([]ChecklistTemplateResponse, len(rows))
	for i, r := range rows {
		items, _ := s.q.ListChecklistItemsByTemplate(ctx, r.MaintenanceOrderChecklistTemplateID)
		out[i] = templateToResp(r, items)
	}
	return out, nil
}

func (s *ChecklistTemplateService) GetByID(ctx context.Context, id int32) (*ChecklistTemplateResponse, error) {
	r, err := s.q.GetChecklistTemplateByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("template de checklist não encontrado")
	}
	if err != nil {
		return nil, err
	}
	items, _ := s.q.ListChecklistItemsByTemplate(ctx, id)
	resp := templateToResp(r, items)
	return &resp, nil
}

func (s *ChecklistTemplateService) Create(ctx context.Context, req CreateChecklistTemplateRequest) (*ChecklistTemplateResponse, error) {
	reasonID := sql.NullInt32{}
	if req.ReasonID != nil {
		reasonID = sql.NullInt32{Int32: *req.ReasonID, Valid: true}
	}
	res, err := s.q.CreateChecklistTemplate(ctx, db.CreateChecklistTemplateParams{
		Name:     req.Name,
		ReasonID: reasonID,
		Active:   req.Active,
	})
	if err != nil {
		return nil, err
	}
	id64, _ := res.LastInsertId()
	return s.GetByID(ctx, int32(id64))
}

func (s *ChecklistTemplateService) Update(ctx context.Context, id int32, req UpdateChecklistTemplateRequest) (*ChecklistTemplateResponse, error) {
	if _, err := s.GetByID(ctx, id); err != nil {
		return nil, err
	}
	reasonID := sql.NullInt32{}
	if req.ReasonID != nil {
		reasonID = sql.NullInt32{Int32: *req.ReasonID, Valid: true}
	}
	if err := s.q.UpdateChecklistTemplate(ctx, db.UpdateChecklistTemplateParams{
		Name:                                req.Name,
		ReasonID:                            reasonID,
		Active:                              req.Active,
		MaintenanceOrderChecklistTemplateID: id,
	}); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *ChecklistTemplateService) Delete(ctx context.Context, id int32) error {
	if _, err := s.GetByID(ctx, id); err != nil {
		return err
	}
	return s.q.SoftDeleteChecklistTemplate(ctx, id)
}

// ── Items ─────────────────────────────────────────────────────────────────────

func (s *ChecklistTemplateService) CreateItem(ctx context.Context, templateID int32, req CreateChecklistItemRequest) (*ChecklistItemResponse, error) {
	if _, err := s.GetByID(ctx, templateID); err != nil {
		return nil, err
	}
	res, err := s.q.CreateChecklistItem(ctx, db.CreateChecklistItemParams{
		TemplateID:     templateID,
		Section:        nullStr(req.Section),
		Label:          req.Label,
		Type:           db.MaintenanceOrderChecklistItemType(req.Type),
		Options:        nullStr(req.Options),
		Metadata:       nullStr(req.Metadata),
		Required:       req.Required,
		SortOrder:      req.SortOrder,
		ParentItemID:   nullInt32(req.ParentItemID),
		ShowWhenValues: nullStr(req.ShowWhenValues),
	})
	if err != nil {
		return nil, err
	}
	id64, _ := res.LastInsertId()
	row, err := s.q.GetChecklistItemByID(ctx, int32(id64))
	if err != nil {
		return nil, err
	}
	resp := itemToResp(row)
	return &resp, nil
}

func (s *ChecklistTemplateService) UpdateItem(ctx context.Context, templateID, itemID int32, req UpdateChecklistItemRequest) (*ChecklistItemResponse, error) {
	if _, err := s.GetByID(ctx, templateID); err != nil {
		return nil, err
	}
	row, err := s.q.GetChecklistItemByID(ctx, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("item não encontrado")
	}
	if err != nil {
		return nil, err
	}
	if row.TemplateID != templateID {
		return nil, errors.New("item não pertence a este template")
	}
	if err := s.q.UpdateChecklistItem(ctx, db.UpdateChecklistItemParams{
		Section:                            nullStr(req.Section),
		Label:                              req.Label,
		Type:                               db.MaintenanceOrderChecklistItemType(req.Type),
		Options:                            nullStr(req.Options),
		Metadata:                           nullStr(req.Metadata),
		Required:                           req.Required,
		SortOrder:                          req.SortOrder,
		ParentItemID:                       nullInt32(req.ParentItemID),
		ShowWhenValues:                     nullStr(req.ShowWhenValues),
		MaintenanceOrderChecklistItemID:    itemID,
	}); err != nil {
		return nil, err
	}
	updated, err := s.q.GetChecklistItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	resp := itemToResp(updated)
	return &resp, nil
}

func (s *ChecklistTemplateService) DeleteItem(ctx context.Context, templateID, itemID int32) error {
	if _, err := s.GetByID(ctx, templateID); err != nil {
		return err
	}
	row, err := s.q.GetChecklistItemByID(ctx, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("item não encontrado")
	}
	if err != nil {
		return err
	}
	if row.TemplateID != templateID {
		return errors.New("item não pertence a este template")
	}
	return s.q.SoftDeleteChecklistItem(ctx, itemID)
}

// ── Converters ────────────────────────────────────────────────────────────────

func templateToResp(t db.MaintenanceOrderChecklistTemplate, items []db.MaintenanceOrderChecklistItem) ChecklistTemplateResponse {
	resp := ChecklistTemplateResponse{
		ID:        t.MaintenanceOrderChecklistTemplateID,
		Name:      t.Name,
		Active:    t.Active,
		CreatedAt: t.CreatedAt,
		Items:     make([]ChecklistItemResponse, len(items)),
	}
	if t.ReasonID.Valid {
		resp.ReasonID = &t.ReasonID.Int32
	}
	for i, it := range items {
		resp.Items[i] = itemToResp(it)
	}
	return resp
}

func itemToResp(it db.MaintenanceOrderChecklistItem) ChecklistItemResponse {
	resp := ChecklistItemResponse{
		ID:        it.MaintenanceOrderChecklistItemID,
		TemplateID: it.TemplateID,
		Label:     it.Label,
		Type:      string(it.Type),
		Required:  it.Required,
		SortOrder: it.SortOrder,
	}
	if it.Section.Valid {
		resp.Section = &it.Section.String
	}
	if it.Options.Valid {
		resp.Options = &it.Options.String
	}
	if it.Metadata.Valid {
		resp.Metadata = &it.Metadata.String
	}
	if it.ParentItemID.Valid {
		resp.ParentItemID = &it.ParentItemID.Int32
	}
	if it.ShowWhenValues.Valid {
		resp.ShowWhenValues = &it.ShowWhenValues.String
	}
	return resp
}

func nullInt32(p *int32) sql.NullInt32 {
	if p == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *p, Valid: true}
}
