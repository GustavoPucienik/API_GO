package logistic

import "time"

// ── Lookup requests ───────────────────────────────────────────────────────────

type CreateStatusRequest struct {
	Name              string `json:"name"`
	BackgroundColor   string `json:"backgroundColor"`
	BorderColor       string `json:"borderColor"`
	TextColor         string `json:"textColor"`
	SortIndex         int32  `json:"sortIndex"`
	IsClosed          bool   `json:"isClosed"`
	IsPendingApproval bool   `json:"isPendingApproval"`
	IsRhStep          bool   `json:"isRhStep"`
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

type CreateOperatorRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type UpdateOperatorRequest = CreateOperatorRequest

// ── Lookup responses ──────────────────────────────────────────────────────────

type StatusResponse struct {
	ID                int32  `json:"id"`
	Name              string `json:"name"`
	BackgroundColor   string `json:"backgroundColor"`
	BorderColor       string `json:"borderColor"`
	TextColor         string `json:"textColor"`
	SortIndex         int32  `json:"sortIndex"`
	IsClosed          bool   `json:"isClosed"`
	IsPendingApproval bool   `json:"isPendingApproval"`
	IsRhStep          bool   `json:"isRhStep"`
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

type OperatorResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type LookupsResponse struct {
	Status      []StatusResponse    `json:"status"`
	Priorities  []PriorityResponse  `json:"priorities"`
	Reasons     []ReasonResponse    `json:"reasons"`
	Transitions []TransitionResponse `json:"transitions"`
	Operators   []OperatorResponse  `json:"operators"`
}

// ── Order requests ────────────────────────────────────────────────────────────

type ItemInput struct {
	ItemCode string `json:"itemCode"`
	Quantity int32  `json:"quantity"`
}

type CreateOrderRequest struct {
	RequesterID     int32       `json:"requesterId"`
	StatusID        *int32      `json:"statusId"`
	PriorityID      int32       `json:"priorityId"`
	ReasonID        int32       `json:"reasonId"`
	Description     string      `json:"description"`
	SalespersonName *string     `json:"salespersonName"`
	ClientCode      string      `json:"clientCode"`
	ClientName      string      `json:"clientName"`
	AliasName       string      `json:"aliasName"`
	ContactType     *string     `json:"contactType"`
	ContactName     *string     `json:"contactName"`
	ContactPhone    *string     `json:"contactPhone"`
	ContactMobile   *string     `json:"contactMobile"`
	ContactEmail    *string     `json:"contactEmail"`
	AddressName     *string     `json:"addressName"`
	AddressType     *string     `json:"addressType"`
	TypeOfAddress   *string     `json:"typeOfAddress"`
	Street          *string     `json:"street"`
	StreetNo        *string     `json:"streetNo"`
	Block           *string     `json:"block"`
	City            *string     `json:"city"`
	County          *string     `json:"county"`
	State           *string     `json:"state"`
	Country         *string     `json:"country"`
	ZipCode         *string     `json:"zipCode"`
	BuildingFloorRoom *string   `json:"buildingFloorRoom"`
	ScheduleDate    *string     `json:"scheduleDate"`
	Items           []ItemInput `json:"items"`
}

type UpdateOrderRequest struct {
	AssignedToID    *int64   `json:"assignedToId"`
	StatusID        *int32   `json:"statusId"`
	PriorityID      *int32   `json:"priorityId"`
	ReasonID        *int32   `json:"reasonId"`
	Description     *string  `json:"description"`
	SalespersonName *string  `json:"salespersonName"`
	ClientCode      *string  `json:"clientCode"`
	ClientName      *string  `json:"clientName"`
	AliasName       *string  `json:"aliasName"`
	ContactType     *string  `json:"contactType"`
	ContactName     *string  `json:"contactName"`
	ContactPhone    *string  `json:"contactPhone"`
	ContactMobile   *string  `json:"contactMobile"`
	ContactEmail    *string  `json:"contactEmail"`
	AddressName     *string  `json:"addressName"`
	AddressType     *string  `json:"addressType"`
	TypeOfAddress   *string  `json:"typeOfAddress"`
	Street          *string  `json:"street"`
	StreetNo        *string  `json:"streetNo"`
	Block           *string  `json:"block"`
	City            *string  `json:"city"`
	County          *string  `json:"county"`
	State           *string  `json:"state"`
	Country         *string  `json:"country"`
	ZipCode         *string  `json:"zipCode"`
	BuildingFloorRoom *string `json:"buildingFloorRoom"`
	ScheduleDate    *string  `json:"scheduleDate"`
	DriverName      *string  `json:"driverName"`
	CheckerName     *string  `json:"checkerName"`
	HasSalaryDiscount *bool  `json:"hasSalaryDiscount"`
	SalaryDiscountReason *string `json:"salaryDiscountReason"`
	Resolution      *string  `json:"resolution"`
	ClosedAt        *string  `json:"closedAt"`
	Items           *[]ItemInput `json:"items"`
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
	IsRhStep        bool   `json:"isRhStep"`
}

type PrioritySummary struct {
	ID              int32  `json:"id"`
	Name            string `json:"name"`
	BackgroundColor string `json:"backgroundColor"`
}

type ItemResponse struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	ItemCode string `json:"itemCode"`
	Quantity int32  `json:"quantity"`
}

type OrderResponse struct {
	ID                   int32            `json:"id"`
	RequesterID          int32            `json:"requesterId"`
	RequesterName        string           `json:"requesterName"`
	AssignedToID         *int32           `json:"assignedToId"`
	AssignedToName       *string          `json:"assignedToName"`
	DriverName           *string          `json:"driverName"`
	CheckerName          *string          `json:"checkerName"`
	CheckerSignature     *string          `json:"checkerSignature"`
	SalespersonName      *string          `json:"salespersonName"`
	ClientCode           string           `json:"clientCode"`
	ClientName           string           `json:"clientName"`
	ClientAliasName      string           `json:"clientAliasName"`
	ContactType          *string          `json:"contactType"`
	ContactName          *string          `json:"contactName"`
	ContactPhone         *string          `json:"contactPhone"`
	ContactMobile        *string          `json:"contactMobile"`
	ContactEmail         *string          `json:"contactEmail"`
	AddressName          *string          `json:"addressName"`
	AddressType          *string          `json:"addressType"`
	TypeOfAddress        *string          `json:"typeOfAddress"`
	Street               *string          `json:"street"`
	StreetNo             *string          `json:"streetNo"`
	Block                *string          `json:"block"`
	City                 *string          `json:"city"`
	County               *string          `json:"county"`
	State                *string          `json:"state"`
	Country              *string          `json:"country"`
	ZipCode              *string          `json:"zipCode"`
	BuildingFloorRoom    *string          `json:"buildingFloorRoom"`
	StatusID             int32            `json:"statusId"`
	PriorityID           int32            `json:"priorityId"`
	ReasonID             int32            `json:"reasonId"`
	ScheduleDate         *time.Time       `json:"scheduleDate"`
	Description          string           `json:"description"`
	HasSalaryDiscount    *bool            `json:"hasSalaryDiscount"`
	SalaryDiscountReason *string          `json:"salaryDiscountReason"`
	CapturedAt           *time.Time       `json:"capturedAt"`
	Resolution           *string          `json:"resolution"`
	ClosedAt             *time.Time       `json:"closedAt"`
	CreatedAt            time.Time        `json:"createdAt"`
	UpdatedAt            time.Time        `json:"updatedAt"`
	Status               StatusSummary    `json:"status"`
	Priority             PrioritySummary  `json:"priority"`
	ReasonName           string           `json:"reasonName"`
	Items                []ItemResponse   `json:"items"`
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

// ── Config ────────────────────────────────────────────────────────────────────

type AccessLevelItem struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type ConfigResponse struct {
	AssignableAccessLevels      []AccessLevelItem `json:"assignableAccessLevels"`
	SalaryDiscountAccessLevelID *int32            `json:"salaryDiscountAccessLevelId"`
}

type UpdateConfigRequest struct {
	AssignableAccessLevelIDs    []int32 `json:"assignableAccessLevelIds"`
	SalaryDiscountAccessLevelID *int32  `json:"salaryDiscountAccessLevelId"`
}

// ── Config ────────────────────────────────────────────────────────────────────

type AssignableUserResponse struct {
	ID       int32   `json:"id"`
	Name     string  `json:"name"`
	Lastname *string `json:"lastname"`
	Email    string  `json:"email"`
}

// ── Pagination ────────────────────────────────────────────────────────────────

type FindPaginatedParams struct {
	Page                int
	Limit               int
	Search              string
	SortBy              string
	SortOrder           string
	StatusID            *int32
	PriorityID          *int32
	ReasonID            *int32
	AssignedToID        *int32
	StartDate           *string
	EndDate             *string
	DriverName          *string
	OnlyWithScheduleDate bool
}
