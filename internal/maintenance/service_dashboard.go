package maintenance

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type DashboardService struct {
	sqlDB *sql.DB
}

func NewDashboardService(sqlDB *sql.DB) *DashboardService {
	return &DashboardService{sqlDB: sqlDB}
}

func (s *DashboardService) applyFilters(baseWhere []string, args []any, f DashboardFilter, alias string) ([]string, []any) {
	a := alias
	if f.ClientCode != nil {
		baseWhere = append(baseWhere, a+".client_code = ?")
		args = append(args, *f.ClientCode)
	}
	if f.AssignedToID != nil {
		baseWhere = append(baseWhere, `EXISTS (
          SELECT 1 FROM maintenance_order_schedule_date_technician _ft
          INNER JOIN maintenance_order_schedule_date _sd
            ON _sd.maintenance_order_schedule_date_id = _ft.maintenance_order_schedule_date_id
          WHERE _sd.maintenance_order_id = `+a+`.maintenance_order_id
            AND _ft.technician_id = ?
            AND _ft.deleted_at IS NULL AND _sd.deleted_at IS NULL
        )`)
		args = append(args, *f.AssignedToID)
	}
	if f.StatusID != nil {
		baseWhere = append(baseWhere, a+".status_id = ?")
		args = append(args, *f.StatusID)
	}
	if f.PriorityID != nil {
		baseWhere = append(baseWhere, a+".priority_id = ?")
		args = append(args, *f.PriorityID)
	}
	if f.ReasonID != nil {
		baseWhere = append(baseWhere, a+".reason_id = ?")
		args = append(args, *f.ReasonID)
	}
	if f.DateFrom != nil {
		baseWhere = append(baseWhere, a+".created_at >= ?")
		args = append(args, *f.DateFrom)
	}
	if f.DateTo != nil {
		baseWhere = append(baseWhere, a+".created_at <= ?")
		args = append(args, *f.DateTo)
	}
	if f.City != nil {
		baseWhere = append(baseWhere, a+".city = ?")
		args = append(args, *f.City)
	}
	if f.State != nil {
		baseWhere = append(baseWhere, a+".state = ?")
		args = append(args, *f.State)
	}
	return baseWhere, args
}

type SummaryResult struct {
	Total      int64 `json:"total"`
	Open       int64 `json:"open"`
	Concluded  int64 `json:"concluded"`
	Unassigned int64 `json:"unassigned"`
	Overdue    int64 `json:"overdue"`
}

func (s *DashboardService) GetSummary(ctx context.Context, f DashboardFilter) (*SummaryResult, error) {
	base := []string{"o.deleted_at IS NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	run := func(extra string, extraArgs ...any) (int64, error) {
		q := "SELECT COUNT(*) FROM maintenance_order o" + whereSQL + extra
		allArgs := append(args, extraArgs...)
		var n int64
		return n, s.sqlDB.QueryRowContext(ctx, q, allArgs...).Scan(&n)
	}

	total, err := run("")
	if err != nil {
		return nil, err
	}
	open, err := run(" AND o.status_id IN (SELECT maintenance_order_status_id FROM maintenance_order_status WHERE is_closed = 0)")
	if err != nil {
		return nil, err
	}
	concluded, err := run(" AND o.status_id IN (SELECT maintenance_order_status_id FROM maintenance_order_status WHERE is_closed = 1)")
	if err != nil {
		return nil, err
	}
	unassigned, err := run(` AND NOT EXISTS (
		SELECT 1 FROM maintenance_order_schedule_date_technician _ft
		INNER JOIN maintenance_order_schedule_date _sd ON _sd.maintenance_order_schedule_date_id = _ft.maintenance_order_schedule_date_id
		WHERE _sd.maintenance_order_id = o.maintenance_order_id AND _ft.deleted_at IS NULL AND _sd.deleted_at IS NULL
	)`)
	if err != nil {
		return nil, err
	}
	overdue, err := run(` AND o.status_id IN (SELECT maintenance_order_status_id FROM maintenance_order_status WHERE is_closed = 0)
	  AND EXISTS (SELECT 1 FROM maintenance_order_schedule_date _sd WHERE _sd.maintenance_order_id = o.maintenance_order_id AND _sd.date < NOW() AND _sd.deleted_at IS NULL)`)
	if err != nil {
		return nil, err
	}

	return &SummaryResult{Total: total, Open: open, Concluded: concluded, Unassigned: unassigned, Overdue: overdue}, nil
}

type NameCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type StatusCount struct {
	Name        string `json:"name"`
	BorderColor string `json:"borderColor"`
	Count       int64  `json:"count"`
}

func (s *DashboardService) GetByStatus(ctx context.Context, f DashboardFilter) ([]StatusCount, error) {
	base := []string{"o.deleted_at IS NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := "SELECT st.name, st.border_color, COUNT(o.maintenance_order_id) FROM maintenance_order o JOIN maintenance_order_status st ON st.maintenance_order_status_id = o.status_id" + whereSQL + " GROUP BY st.maintenance_order_status_id ORDER BY st.sort_index ASC"
	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StatusCount
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Name, &sc.BorderColor, &sc.Count); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *DashboardService) GetByPriority(ctx context.Context, f DashboardFilter) ([]StatusCount, error) {
	base := []string{"o.deleted_at IS NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := "SELECT p.name, p.border_color, COUNT(o.maintenance_order_id) FROM maintenance_order o JOIN maintenance_order_priorities p ON p.maintenance_order_priorities_id = o.priority_id" + whereSQL + " GROUP BY p.maintenance_order_priorities_id ORDER BY p.sort_index ASC"
	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StatusCount
	for rows.Next() {
		var sc StatusCount
		if err := rows.Scan(&sc.Name, &sc.BorderColor, &sc.Count); err != nil {
			return nil, err
		}
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *DashboardService) GetByReason(ctx context.Context, f DashboardFilter) ([]NameCount, error) {
	base := []string{"o.deleted_at IS NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := "SELECT r.name, COUNT(o.maintenance_order_id) FROM maintenance_order o JOIN maintenance_order_reason r ON r.maintenance_order_reason_id = o.reason_id" + whereSQL + " GROUP BY r.maintenance_order_reason_id ORDER BY COUNT(o.maintenance_order_id) DESC"
	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NameCount
	for rows.Next() {
		var nc NameCount
		if err := rows.Scan(&nc.Name, &nc.Count); err != nil {
			return nil, err
		}
		out = append(out, nc)
	}
	return out, rows.Err()
}

type MonthCount struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

func (s *DashboardService) GetByMonth(ctx context.Context, f DashboardFilter) ([]MonthCount, error) {
	base := []string{"o.deleted_at IS NULL"}
	args := []any{}
	if f.DateFrom == nil && f.DateTo == nil {
		base = append(base, "o.created_at >= DATE_SUB(NOW(), INTERVAL 6 MONTH)")
	}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := "SELECT DATE_FORMAT(o.created_at, '%Y-%m'), COUNT(o.maintenance_order_id) FROM maintenance_order o" + whereSQL + " GROUP BY DATE_FORMAT(o.created_at, '%Y-%m') ORDER BY DATE_FORMAT(o.created_at, '%Y-%m') ASC"
	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MonthCount
	for rows.Next() {
		var mc MonthCount
		if err := rows.Scan(&mc.Month, &mc.Count); err != nil {
			return nil, err
		}
		out = append(out, mc)
	}
	return out, rows.Err()
}

type ClientCount struct {
	ClientCode string `json:"clientCode"`
	ClientName string `json:"clientName"`
	Count      int64  `json:"count"`
}

func (s *DashboardService) GetByClient(ctx context.Context, f DashboardFilter) ([]ClientCount, error) {
	base := []string{"o.deleted_at IS NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := "SELECT o.client_code, o.client_name, COUNT(o.maintenance_order_id) FROM maintenance_order o" + whereSQL + " GROUP BY o.client_code, o.client_name ORDER BY COUNT(o.maintenance_order_id) DESC LIMIT 10"
	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClientCount
	for rows.Next() {
		var cc ClientCount
		if err := rows.Scan(&cc.ClientCode, &cc.ClientName, &cc.Count); err != nil {
			return nil, err
		}
		out = append(out, cc)
	}
	return out, rows.Err()
}

func (s *DashboardService) GetByRequester(ctx context.Context, f DashboardFilter) ([]NameCount, error) {
	base := []string{"o.deleted_at IS NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := "SELECT o.requester_name, COUNT(o.maintenance_order_id) FROM maintenance_order o" + whereSQL + " GROUP BY o.requester_name ORDER BY COUNT(o.maintenance_order_id) DESC"
	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []NameCount
	for rows.Next() {
		var nc NameCount
		if err := rows.Scan(&nc.Name, &nc.Count); err != nil {
			return nil, err
		}
		out = append(out, nc)
	}
	return out, rows.Err()
}

type CityCount struct {
	City  string `json:"city"`
	State string `json:"state"`
	Count int64  `json:"count"`
}

func (s *DashboardService) GetByCity(ctx context.Context, f DashboardFilter) ([]CityCount, error) {
	base := []string{"o.deleted_at IS NULL", "o.city IS NOT NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := "SELECT o.city, o.state, COUNT(o.maintenance_order_id) FROM maintenance_order o" + whereSQL + " GROUP BY o.city, o.state ORDER BY COUNT(o.maintenance_order_id) DESC"
	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CityCount
	for rows.Next() {
		var cc CityCount
		var state sql.NullString
		if err := rows.Scan(&cc.City, &state, &cc.Count); err != nil {
			return nil, err
		}
		if state.Valid {
			cc.State = state.String
		}
		out = append(out, cc)
	}
	return out, rows.Err()
}

type TimeEntrySummary struct {
	Type         string  `json:"type"`
	TotalMinutes float64 `json:"totalMinutes"`
	OrderCount   int64   `json:"orderCount"`
}

func (s *DashboardService) GetTimeEntriesSummary(ctx context.Context, f DashboardFilter) ([]TimeEntrySummary, error) {
	base := []string{"o.deleted_at IS NULL", "te.ended_at IS NOT NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := "SELECT t.name, SUM(TIMESTAMPDIFF(MINUTE, te.started_at, te.ended_at)), COUNT(DISTINCT o.maintenance_order_id) FROM maintenance_order_time_entry te JOIN maintenance_order_time_entry_type t ON t.maintenance_order_time_entry_type_id = te.type_id JOIN maintenance_order o ON o.maintenance_order_id = te.maintenance_order_id" + whereSQL + " GROUP BY t.name"
	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TimeEntrySummary
	for rows.Next() {
		var te TimeEntrySummary
		var totalMin sql.NullFloat64
		if err := rows.Scan(&te.Type, &totalMin, &te.OrderCount); err != nil {
			return nil, err
		}
		if totalMin.Valid {
			te.TotalMinutes = totalMin.Float64
		}
		out = append(out, te)
	}
	return out, rows.Err()
}

type TrendResult struct {
	CurrentMonth  TrendPeriod `json:"currentMonth"`
	PreviousMonth TrendPeriod `json:"previousMonth"`
	VariationPct  *float64    `json:"variationPercent"`
}

type TrendPeriod struct {
	Period string `json:"period"`
	Count  int64  `json:"count"`
}

func (s *DashboardService) GetTrend(ctx context.Context, f DashboardFilter) (*TrendResult, error) {
	now := time.Now()
	startCurrent := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startPrev := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())

	run := func(from, to *time.Time) (int64, error) {
		base := []string{"o.deleted_at IS NULL"}
		args := []any{}
		if from != nil {
			base = append(base, "o.created_at >= ?")
			args = append(args, *from)
		}
		if to != nil {
			base = append(base, "o.created_at < ?")
			args = append(args, *to)
		}
		base, args = s.applyFilters(base, args, f, "o")
		var n int64
		return n, s.sqlDB.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM maintenance_order o WHERE "+strings.Join(base, " AND "),
			args...).Scan(&n)
	}

	current, err := run(&startCurrent, nil)
	if err != nil {
		return nil, err
	}
	previous, err := run(&startPrev, &startCurrent)
	if err != nil {
		return nil, err
	}

	fmtPeriod := func(t time.Time) string {
		return t.Format("2006-01")
	}

	result := &TrendResult{
		CurrentMonth:  TrendPeriod{Period: fmtPeriod(now), Count: current},
		PreviousMonth: TrendPeriod{Period: fmtPeriod(startPrev), Count: previous},
	}
	if previous > 0 {
		v := float64(current-previous) / float64(previous) * 100
		vRounded := float64(int(v*10)) / 10
		result.VariationPct = &vRounded
	}
	return result, nil
}

type TechnicianSummary struct {
	Name      string `json:"name"`
	Total     int64  `json:"total"`
	Concluded int64  `json:"concluded"`
	Open      int64  `json:"open"`
}

func (s *DashboardService) GetSummaryByTechnician(ctx context.Context, f DashboardFilter) ([]TechnicianSummary, error) {
	base := []string{"o.deleted_at IS NULL"}
	args := []any{}
	base, args = s.applyFilters(base, args, f, "o")
	whereSQL := " WHERE " + strings.Join(base, " AND ")

	q := `SELECT tech.technician_name,
	  COUNT(DISTINCT o.maintenance_order_id),
	  SUM(CASE WHEN st.is_closed = 1 THEN 1 ELSE 0 END),
	  SUM(CASE WHEN st.is_closed = 0 THEN 1 ELSE 0 END)
	FROM maintenance_order o
	JOIN maintenance_order_status st ON st.maintenance_order_status_id = o.status_id
	JOIN maintenance_order_schedule_date sd ON sd.maintenance_order_id = o.maintenance_order_id AND sd.deleted_at IS NULL
	JOIN maintenance_order_schedule_date_technician tech ON tech.maintenance_order_schedule_date_id = sd.maintenance_order_schedule_date_id AND tech.deleted_at IS NULL` +
		whereSQL + " GROUP BY tech.technician_name ORDER BY COUNT(DISTINCT o.maintenance_order_id) DESC"

	rows, err := s.sqlDB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TechnicianSummary
	for rows.Next() {
		var ts TechnicianSummary
		if err := rows.Scan(&ts.Name, &ts.Total, &ts.Concluded, &ts.Open); err != nil {
			return nil, err
		}
		out = append(out, ts)
	}
	return out, rows.Err()
}
