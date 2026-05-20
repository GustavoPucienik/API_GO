package maintenance

import "time"

// ── Shared inputs ─────────────────────────────────────────────────────────────

type ItemInput struct {
	ItemCode string `json:"itemCode"`
	Quantity int32  `json:"quantity"`
}

type TechnicianInput struct {
	TechnicianID   int32  `json:"technicianId"`
	TechnicianName string `json:"technicianName"`
}

type ScheduleDateInput struct {
	Date        time.Time        `json:"date"`
	Technicians []TechnicianInput `json:"technicians"`
}

type TimeEntryInput struct {
	TypeID    int32   `json:"typeId"`
	StartedAt string  `json:"startedAt"`
	EndedAt   *string `json:"endedAt"`
}

// ── Order requests ────────────────────────────────────────────────────────────

type CreateOrderRequest struct {
	RequesterID  int32  `json:"requesterId"`
	StatusID     *int32 `json:"statusId"`
	PriorityID   int32  `json:"priorityId"`
	ReasonID     int32  `json:"reasonId"`
	Description  string `json:"description"`

	SalespersonName *string `json:"salespersonName"`

	ClientCode string `json:"clientCode"`
	ClientName string `json:"clientName"`
	AliasName  string `json:"aliasName"`

	ContactType   *string `json:"contactType"`
	ContactName   *string `json:"contactName"`
	ContactPhone  *string `json:"contactPhone"`
	ContactMobile *string `json:"contactMobile"`
	ContactEmail  *string `json:"contactEmail"`

	AddressName       *string `json:"addressName"`
	AddressType       *string `json:"addressType"`
	TypeOfAddress     *string `json:"typeOfAddress"`
	Street            *string `json:"street"`
	StreetNo          *string `json:"streetNo"`
	Block             *string `json:"block"`
	City              *string `json:"city"`
	County            *string `json:"county"`
	State             *string `json:"state"`
	Country           *string `json:"country"`
	ZipCode           *string `json:"zipCode"`
	BuildingFloorRoom *string `json:"buildingFloorRoom"`

	Items         []ItemInput         `json:"items"`
	ScheduleDates []ScheduleDateInput `json:"scheduleDates"`
}

type UpdateOrderRequest struct {
	StatusID   *int32 `json:"statusId"`
	PriorityID *int32 `json:"priorityId"`
	ReasonID   *int32 `json:"reasonId"`
	Description *string `json:"description"`

	SalespersonName *string `json:"salespersonName"`

	ClientCode *string `json:"clientCode"`
	ClientName *string `json:"clientName"`
	AliasName  *string `json:"aliasName"`

	ContactType   *string `json:"contactType"`
	ContactName   *string `json:"contactName"`
	ContactPhone  *string `json:"contactPhone"`
	ContactMobile *string `json:"contactMobile"`
	ContactEmail  *string `json:"contactEmail"`

	AddressName       *string `json:"addressName"`
	AddressType       *string `json:"addressType"`
	TypeOfAddress     *string `json:"typeOfAddress"`
	Street            *string `json:"street"`
	StreetNo          *string `json:"streetNo"`
	Block             *string `json:"block"`
	City              *string `json:"city"`
	County            *string `json:"county"`
	State             *string `json:"state"`
	Country           *string `json:"country"`
	ZipCode           *string `json:"zipCode"`
	BuildingFloorRoom *string `json:"buildingFloorRoom"`

	Resolution       *string `json:"resolution"`
	MotivoResolution *string `json:"motivoResolution"`
	ClosedAt         *string `json:"closedAt"`

	Items           *[]ItemInput         `json:"items"`
	ResolutionItems *[]ItemInput         `json:"resolutionItems"`
	TimeEntries     *[]TimeEntryInput    `json:"timeEntries"`
	ScheduleDates   *[]ScheduleDateInput `json:"scheduleDates"`
}

type AssignRequest struct {
	UserID int64 `json:"userId"`
}

type SignatureRequest struct {
	ClientSignature string `json:"clientSignature"`
}

type ChangeStatusRequest struct {
	StatusID int32 `json:"statusId"`
}

// ── Lookup requests ───────────────────────────────────────────────────────────

type CreateStatusRequest struct {
	Name            string `json:"name"`
	BackgroundColor string `json:"backgroundColor"`
	BorderColor     string `json:"borderColor"`
	TextColor       string `json:"textColor"`
	SortIndex       int32  `json:"sortIndex"`
	IsClosed        bool   `json:"isClosed"`
	IsDefault       bool   `json:"isDefault"`
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
	Name string `json:"name"`
}

type UpdateReasonRequest = CreateReasonRequest

type CreateNoteTemplateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateNoteTemplateRequest = CreateNoteTemplateRequest

type CreateTimeEntryTypeRequest struct {
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

type UpdateTimeEntryTypeRequest = CreateTimeEntryTypeRequest

// ── Attachment requests ───────────────────────────────────────────────────────

type UploadedFile struct {
	OriginalName string
	SavedPath    string
	Size         int64
}

// ── Checklist template requests ───────────────────────────────────────────────

type CreateChecklistTemplateRequest struct {
	Name     string `json:"name"`
	ReasonID *int32 `json:"reasonId"`
	Active   bool   `json:"active"`
}

type UpdateChecklistTemplateRequest = CreateChecklistTemplateRequest

type CreateChecklistItemRequest struct {
	Section        *string  `json:"section"`
	Label          string   `json:"label"`
	Type           string   `json:"type"`
	Options        *string  `json:"options"`
	Metadata       *string  `json:"metadata"`
	Required       bool     `json:"required"`
	SortOrder      int32    `json:"sortOrder"`
	ParentItemID   *int32   `json:"parentItemId"`
	ShowWhenValues *string  `json:"showWhenValues"`
}

type UpdateChecklistItemRequest = CreateChecklistItemRequest

// ── Order checklist requests ──────────────────────────────────────────────────

type SubmitChecklistRequest struct {
	TemplateID  int32            `json:"templateId"`
	CompletedAt *string          `json:"completedAt"`
	Answers     []AnswerInput    `json:"answers"`
}

type AnswerInput struct {
	ItemID int32   `json:"itemId"`
	Value  *string `json:"value"`
}

// ── Response types ────────────────────────────────────────────────────────────

type StatusResponse struct {
	ID              int32  `json:"id"`
	Name            string `json:"name"`
	BackgroundColor string `json:"backgroundColor"`
	BorderColor     string `json:"borderColor"`
	TextColor       string `json:"textColor"`
	SortIndex       int32  `json:"sortIndex"`
	IsClosed        bool   `json:"isClosed"`
	IsDefault       bool   `json:"isDefault"`
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
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type NoteTemplateResponse struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TimeEntryTypeResponse struct {
	ID    int32   `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

type OrderItemResponse struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	ItemCode string `json:"itemCode"`
	Quantity int32  `json:"quantity"`
}

type TechnicianResponse struct {
	TechnicianID   int32  `json:"technicianId"`
	TechnicianName string `json:"technicianName"`
}

type ScheduleDateResponse struct {
	ID          int32                `json:"id"`
	Date        time.Time            `json:"date"`
	Technicians []TechnicianResponse `json:"technicians"`
}

type TimeEntryResponse struct {
	ID        int32      `json:"id"`
	TypeID    int32      `json:"typeId"`
	TypeName  string     `json:"typeName"`
	TypeColor *string    `json:"typeColor"`
	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt"`
}

type AttachmentResponse struct {
	ID        int32     `json:"id"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"`
	FileSize  int32     `json:"fileSize"`
	CreatedAt time.Time `json:"createdAt"`
}

type OrderResponse struct {
	ID                int32      `json:"id"`
	RequesterID       int32      `json:"requesterId"`
	RequesterName     string     `json:"requesterName"`
	SalespersonName   *string    `json:"salespersonName"`
	ClientCode        string     `json:"clientCode"`
	ClientName        string     `json:"clientName"`
	AliasName         string     `json:"aliasName"`
	ContactType       *string    `json:"contactType"`
	ContactName       *string    `json:"contactName"`
	ContactPhone      *string    `json:"contactPhone"`
	ContactMobile     *string    `json:"contactMobile"`
	ContactEmail      *string    `json:"contactEmail"`
	AddressName       *string    `json:"addressName"`
	AddressType       *string    `json:"addressType"`
	TypeOfAddress     *string    `json:"typeOfAddress"`
	Street            *string    `json:"street"`
	StreetNo          *string    `json:"streetNo"`
	Block             *string    `json:"block"`
	City              *string    `json:"city"`
	County            *string    `json:"county"`
	State             *string    `json:"state"`
	Country           *string    `json:"country"`
	ZipCode           *string    `json:"zipCode"`
	BuildingFloorRoom *string    `json:"buildingFloorRoom"`
	Status            StatusResponse   `json:"status"`
	Priority          PriorityResponse `json:"priority"`
	Reason            ReasonResponse   `json:"reason"`
	Description       string           `json:"description"`
	CapturedAt        *time.Time       `json:"capturedAt"`
	Resolution        *string          `json:"resolution"`
	MotivoResolution  *string          `json:"motivoResolution"`
	ClientSignature   *string          `json:"clientSignature"`
	ClosedAt          *time.Time       `json:"closedAt"`
	Items             []OrderItemResponse    `json:"items"`
	ResolutionItems   []OrderItemResponse    `json:"resolutionItems"`
	TimeEntries       []TimeEntryResponse    `json:"timeEntries"`
	ScheduleDates     []ScheduleDateResponse `json:"scheduleDates"`
	CreatedAt         time.Time              `json:"createdAt"`
	UpdatedAt         time.Time              `json:"updatedAt"`
}

type LookupsResponse struct {
	Status         []StatusResponse       `json:"status"`
	Priorities     []PriorityResponse     `json:"priorities"`
	Reasons        []ReasonResponse       `json:"reasons"`
	NoteTemplates  []NoteTemplateResponse `json:"noteTemplates"`
	TimeEntryTypes []TimeEntryTypeResponse `json:"timeEntryTypes"`
}

type ChecklistTemplateResponse struct {
	ID        int32                      `json:"id"`
	Name      string                     `json:"name"`
	ReasonID  *int32                     `json:"reasonId"`
	Active    bool                       `json:"active"`
	Items     []ChecklistItemResponse    `json:"items"`
	CreatedAt time.Time                  `json:"createdAt"`
}

type ChecklistItemResponse struct {
	ID             int32   `json:"id"`
	TemplateID     int32   `json:"templateId"`
	Section        *string `json:"section"`
	Label          string  `json:"label"`
	Type           string  `json:"type"`
	Options        *string `json:"options"`
	Metadata       *string `json:"metadata"`
	Required       bool    `json:"required"`
	SortOrder      int32   `json:"sortOrder"`
	ParentItemID   *int32  `json:"parentItemId"`
	ShowWhenValues *string `json:"showWhenValues"`
}

type OrderChecklistResponse struct {
	ID          int32                    `json:"id"`
	OrderID     int32                    `json:"orderId"`
	TemplateID  int32                    `json:"templateId"`
	Template    *ChecklistTemplateResponse `json:"template,omitempty"`
	CompletedAt *time.Time               `json:"completedAt"`
	Answers     []ChecklistAnswerResponse `json:"answers"`
	CreatedAt   time.Time                `json:"createdAt"`
}

type ChecklistAnswerResponse struct {
	ID          int32   `json:"id"`
	ChecklistID int32   `json:"checklistId"`
	ItemID      int32   `json:"itemId"`
	Value       *string `json:"value"`
}

// ── Dashboard filter ──────────────────────────────────────────────────────────

type DashboardFilter struct {
	ClientCode   *string `json:"clientCode"`
	AssignedToID *int64  `json:"assignedToId"`
	StatusID     *int32  `json:"statusId"`
	PriorityID   *int32  `json:"priorityId"`
	ReasonID     *int32  `json:"reasonId"`
	DateFrom     *string `json:"dateFrom"`
	DateTo       *string `json:"dateTo"`
	City         *string `json:"city"`
	State        *string `json:"state"`
}

// ── Config ────────────────────────────────────────────────────────────────────

type UpdateAssignableAccessLevelsRequest struct {
	AssignableAccessLevelIDs []int32 `json:"assignableAccessLevelIds"`
}

type MaintenanceConfigResponse struct {
	AssignableAccessLevelIDs []int64 `json:"assignableAccessLevelIds"`
}

type AssignableUserResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Lastname *string `json:"lastname"`
	Email    string  `json:"email"`
}
