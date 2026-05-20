package admin

// ─── Auth ────────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PermissionDTO struct {
	Module     string `json:"module"`
	Permission int    `json:"permission"`
	DataScope  string `json:"dataScope"`
}

type MenuItemDTO struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Paths     string `json:"paths"`
	SortIndex int32  `json:"sortIndex"`
	ModuleID  int64  `json:"moduleId"`
	ParentID  *int64 `json:"parentId"`
}

type UserPermissionsDTO struct {
	Permissions []PermissionDTO `json:"permissions"`
	Menus       []MenuItemDTO   `json:"menus"`
}

type AuthResponse struct {
	Token       string             `json:"token"`
	User        UserResponse       `json:"user"`
	Permissions UserPermissionsDTO `json:"permissions"`
}

// ─── User ─────────────────────────────────────────────────────────────────────

type CreateUserRequest struct {
	Name     string  `json:"name"`
	Lastname *string `json:"lastname"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
}

type UpdateUserRequest struct {
	Name         string  `json:"name"`
	Lastname     *string `json:"lastname"`
	Email        string  `json:"email"`
	IsActive     *bool   `json:"isActive"`
	Groups       []int64 `json:"groups"`
	AccessLevels []int64 `json:"accessLevels"`
}

type ResetPasswordSelfRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

type ResetPasswordAdminRequest struct {
	NewPassword string `json:"newPassword"`
}

type UserResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Lastname *string `json:"lastname"`
	Email    string  `json:"email"`
	IsActive bool    `json:"isActive"`
}

type GroupBasic struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type AccessLevelBasic struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type UserWithRelationsResponse struct {
	UserResponse
	Groups       []GroupBasic       `json:"groups"`
	AccessLevels []AccessLevelBasic `json:"accessLevels"`
}

// ─── Group ────────────────────────────────────────────────────────────────────

type CreateGroupRequest struct {
	Name           string  `json:"name"`
	AccessLevelIDs []int64 `json:"accessLevelIds"`
}

type UpdateGroupRequest struct {
	Name           *string `json:"name"`
	AccessLevelIDs []int64 `json:"accessLevelIds"`
}

type GroupResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	ParentID *int64  `json:"parentId"`
}

// ─── Access Level ─────────────────────────────────────────────────────────────

type AccessLevelModuleInput struct {
	ModuleID      int64  `json:"moduleId"`
	Read          bool   `json:"read"`
	Create        bool   `json:"create"`
	Update        bool   `json:"update"`
	Delete        bool   `json:"delete"`
	DataScope     string `json:"dataScope"`
	CanBeAssigned bool   `json:"canBeAssigned"`
}

type CreateAccessLevelRequest struct {
	Name    string                   `json:"name"`
	Modules []AccessLevelModuleInput `json:"modules"`
}

type UpdateAccessLevelRequest struct {
	Name    *string                  `json:"name"`
	Modules []AccessLevelModuleInput `json:"modules"`
}

type AccessLevelResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ─── Module ───────────────────────────────────────────────────────────────────

type CreateModuleRequest struct {
	Name string `json:"name"`
}

type UpdateModuleRequest struct {
	Name string `json:"name"`
}

type ModuleResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ─── Menu ─────────────────────────────────────────────────────────────────────

type CreateMenuRequest struct {
	Name          string  `json:"name"`
	SortIndex     int32   `json:"sortIndex"`
	Paths         string  `json:"paths"`
	ModuleID      int64   `json:"moduleId"`
	ParentID      *int64  `json:"parentId"`
	RequiredScope *string `json:"requiredScope"`
}

type UpdateMenuRequest struct {
	Name          string  `json:"name"`
	SortIndex     int32   `json:"sortIndex"`
	Paths         string  `json:"paths"`
	ModuleID      int64   `json:"moduleId"`
	ParentID      *int64  `json:"parentId"`
	RequiredScope *string `json:"requiredScope"`
}

type MenuResponse struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	SortIndex     int32   `json:"sortIndex"`
	Paths         string  `json:"paths"`
	ModuleID      int64   `json:"moduleId"`
	ParentID      *int64  `json:"parentId"`
	RequiredScope *string `json:"requiredScope"`
}

// ─── User Group ───────────────────────────────────────────────────────────────

type SyncUsersInGroupRequest struct {
	UserIDs []int32 `json:"userIds"`
}

type UserGroupUserResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Lastname *string `json:"lastname"`
	Email    string  `json:"email"`
	IsActive bool    `json:"isActive"`
}
