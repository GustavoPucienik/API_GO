package logistic

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	"systemapi/internal/db"
)

type AttachmentService struct {
	sqlDB     *sql.DB
	q         *db.Queries
	uploadDir string
}

func NewAttachmentService(sqlDB *sql.DB, q *db.Queries, uploadDir string) *AttachmentService {
	return &AttachmentService{sqlDB: sqlDB, q: q, uploadDir: uploadDir}
}

func (s *AttachmentService) Upload(ctx context.Context, orderID int32, files []UploadedFile) ([]AttachmentResponse, error) {
	if _, err := s.q.GetLogisticOrderByID(ctx, orderID); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem logística não encontrada")
	} else if err != nil {
		return nil, err
	}

	var out []AttachmentResponse
	for _, f := range files {
		res, err := s.q.CreateLogisticOrderAttachment(ctx, db.CreateLogisticOrderAttachmentParams{
			LogisticOrderID: orderID,
			FileName:        f.OriginalName,
			FilePath:        f.SavedPath,
			FileSize:        int32(f.Size),
		})
		if err != nil {
			return nil, err
		}
		id64, _ := res.LastInsertId()
		att, err := s.q.GetLogisticOrderAttachmentByID(ctx, int32(id64))
		if err != nil {
			return nil, err
		}
		out = append(out, attToResp(att))
	}
	return out, nil
}

func (s *AttachmentService) FindByOrder(ctx context.Context, orderID int32) ([]AttachmentResponse, error) {
	if _, err := s.q.GetLogisticOrderByID(ctx, orderID); errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ordem logística não encontrada")
	} else if err != nil {
		return nil, err
	}
	rows, err := s.q.ListLogisticOrderAttachments(ctx, orderID)
	if err != nil {
		return nil, err
	}
	out := make([]AttachmentResponse, len(rows))
	for i, r := range rows {
		out[i] = AttachmentResponse{
			ID:        r.LogisticOrderAttachmentID,
			FileName:  r.FileName,
			FilePath:  r.FilePath,
			FileSize:  r.FileSize,
			CreatedAt: r.CreatedAt,
		}
	}
	return out, nil
}

func (s *AttachmentService) GetByID(ctx context.Context, id int32) (db.GetLogisticOrderAttachmentByIDRow, error) {
	att, err := s.q.GetLogisticOrderAttachmentByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.GetLogisticOrderAttachmentByIDRow{}, errors.New("arquivo não encontrado")
	}
	return att, err
}

func (s *AttachmentService) Delete(ctx context.Context, id int32) error {
	att, err := s.q.GetLogisticOrderAttachmentByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("arquivo não encontrado")
	}
	if err != nil {
		return err
	}
	fullPath := filepath.Join(s.uploadDir, att.FilePath)
	if _, statErr := os.Stat(fullPath); statErr == nil {
		os.Remove(fullPath)
	}
	return s.q.DeleteLogisticOrderAttachment(ctx, id)
}

func attToResp(a db.GetLogisticOrderAttachmentByIDRow) AttachmentResponse {
	return AttachmentResponse{
		ID:        a.LogisticOrderAttachmentID,
		FileName:  a.FileName,
		FilePath:  a.FilePath,
		FileSize:  a.FileSize,
		CreatedAt: a.CreatedAt,
	}
}
