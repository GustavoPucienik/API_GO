package fiscal

import "time"

// ── Lookup requests ───────────────────────────────────────────────────────────

type CreateStatusRequest struct {
	Name              string `json:"name"`
	BackgroundColor   string `json:"backgroundColor"`
	BorderColor       string `json:"borderColor"`
	TextColor         string `json:"textColor"`
	SortIndex         int32  `json:"sortIndex"`
	IsClosed          bool   `json:"isClosed"`
	IsDefault         bool   `json:"isDefault"`
	IsPendingApproval bool   `json:"isPendingApproval"`
}

type UpdateStatusRequest = CreateStatusRequest

type CreatePriorityRequest struct {
	Name            string `json:"name"`
	BackgroundColor string `json:"backgroundColor"`
	BorderColor     string `json:"borderColor"`
	TextColor       string `json:"textColor"`
	SortIndex       int32  `json:"sortIndex"`
}

type UpdatePriorityRequest = CreatePriorityRequest

type CreateReasonRequest struct {
	Name            string `json:"name"`
	InitialStatusID int32  `json:"initialStatusId"`
}

type UpdateReasonRequest = CreateReasonRequest

type CreateTransitionRequest struct {
	FromStatusID  *int32 `json:"fromStatusId"`
	ToStatusID    int32  `json:"toStatusId"`
	AccessLevelID int32  `json:"accessLevelId"`
}

// ── Lookup responses ──────────────────────────────────────────────────────────

type StatusResponse struct {
	ID                int32  `json:"id"`
	Name              string `json:"name"`
	BackgroundColor   string `json:"backgroundColor"`
	BorderColor       string `json:"borderColor"`
	TextColor         string `json:"textColor"`
	SortIndex         int32  `json:"sortIndex"`
	IsClosed          bool   `json:"isClosed"`
	IsDefault         bool   `json:"isDefault"`
	IsPendingApproval bool   `json:"isPendingApproval"`
}

type PriorityResponse struct {
	ID              int32  `json:"id"`
	Name            string `json:"name"`
	BackgroundColor string `json:"backgroundColor"`
	BorderColor     string `json:"borderColor"`
	TextColor       string `json:"textColor"`
	SortIndex       int32  `json:"sortIndex"`
}

type ReasonResponse struct {
	ID                int32  `json:"id"`
	Name              string `json:"name"`
	InitialStatusID   int32  `json:"initialStatusId"`
	InitialStatusName string `json:"initialStatusName"`
}

type TransitionResponse struct {
	ID            int32  `json:"id"`
	FromStatusID  *int32 `json:"fromStatusId"`
	ToStatusID    int32  `json:"toStatusId"`
	AccessLevelID int32  `json:"accessLevelId"`
}

type LookupsResponse struct {
	Status      []StatusResponse     `json:"status"`
	Priorities  []PriorityResponse   `json:"priorities"`
	Reasons     []ReasonResponse     `json:"reasons"`
	Transitions []TransitionResponse `json:"transitions"`
}

// ── Order requests ────────────────────────────────────────────────────────────

type CreateOrderRequest struct {
	RequesterID     int32   `json:"requesterId"`
	StatusID        *int32  `json:"statusId"`
	PriorityID      int32   `json:"priorityId"`
	ReasonID        int32   `json:"reasonId"`
	Description     string  `json:"description"`
	SalespersonName *string `json:"salespersonName"`
	ClientCode      string  `json:"clientCode"`
	ClientName      string  `json:"clientName"`
	AliasName       string  `json:"aliasName"`
}

type UpdateOrderRequest struct {
	AssignedToID    *int64  `json:"assignedToId"`
	StatusID        *int32  `json:"statusId"`
	PriorityID      *int32  `json:"priorityId"`
	ReasonID        *int32  `json:"reasonId"`
	Description     *string `json:"description"`
	SalespersonName *string `json:"salespersonName"`
	ClientCode      *string `json:"clientCode"`
	ClientName      *string `json:"clientName"`
	AliasName       *string `json:"aliasName"`
	Resolution      *string `json:"resolution"`
	ClosedAt        *string `json:"closedAt"`
}

type AssignRequest struct {
	UserID int64 `json:"userId"`
}

type ChangeStatusRequest struct {
	StatusID int32 `json:"statusId"`
}

// ── Order response ────────────────────────────────────────────────────────────

type StatusSummary struct {
	ID              int32  `json:"id"`
	Name            string `json:"name"`
	BackgroundColor string `json:"backgroundColor"`
	BorderColor     string `json:"borderColor"`
	TextColor       string `json:"textColor"`
	IsClosed        bool   `json:"isClosed"`
}

type PrioritySummary struct {
	ID              int32  `json:"id"`
	Name            string `json:"name"`
	BackgroundColor string `json:"backgroundColor"`
}

type OrderResponse struct {
	ID              int32            `json:"id"`
	RequesterID     int32            `json:"requesterId"`
	RequesterName   string           `json:"requesterName"`
	AssignedToID    *int32           `json:"assignedToId"`
	AssignedToName  *string          `json:"assignedToName"`
	SalespersonName *string          `json:"salespersonName"`
	ClientCode      string           `json:"clientCode"`
	ClientName      string           `json:"clientName"`
	ClientAliasName string           `json:"clientAliasName"`
	StatusID        int32            `json:"statusId"`
	PriorityID      int32            `json:"priorityId"`
	ReasonID        int32            `json:"reasonId"`
	Description     string           `json:"description"`
	CapturedAt      *time.Time       `json:"capturedAt"`
	Resolution      *string          `json:"resolution"`
	ClosedAt        *time.Time       `json:"closedAt"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
	Status          StatusSummary    `json:"status"`
	Priority        PrioritySummary  `json:"priority"`
	ReasonName      string           `json:"reasonName"`
}

type PaginatedResponse struct {
	Data       []OrderResponse `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	TotalPages int             `json:"totalPages"`
	Limit      int             `json:"limit"`
}

// ── Attachment ────────────────────────────────────────────────────────────────

type UploadedFile struct {
	OriginalName string
	SavedPath    string
	Size         int64
}

type AttachmentResponse struct {
	ID        int32     `json:"id"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"`
	FileSize  int32     `json:"fileSize"`
	CreatedAt time.Time `json:"createdAt"`
}

// ── Pagination input ──────────────────────────────────────────────────────────

type FindPaginatedParams struct {
	Page        int
	Limit       int
	Search      string
	SortOrder   string
	StatusID    *int32
	PriorityID  *int32
	ReasonID    *int32
	ResponsibleID *int32
	StartDate   *string
	EndDate     *string
}
