package user_dto

type Create struct {
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
	Phone    string `json:"phone"     validate:"required,min=7,max=50"`
	Password string `json:"password"  validate:"required,min=6"`
	Photo    string `json:"photo"     validate:"omitempty,max=255"`
	Role     string `json:"role"      validate:"omitempty,oneof=user admin"`
	IsActive *bool  `json:"is_active"`
}

type Update struct {
	FullName *string `json:"full_name" validate:"omitempty,min=2,max=255"`
	Phone    *string `json:"phone"     validate:"omitempty,min=7,max=50"`
	Password *string `json:"password"  validate:"omitempty,min=6"`
	Photo    *string `json:"photo"     validate:"omitempty,max=255"`
	Role     *string `json:"role"      validate:"omitempty,oneof=user admin"`
	IsActive *bool   `json:"is_active"`
}

type Filter struct {
	FullName string
	Phone    string
	Role     string
	IsActive *bool
}

type AdminFilter struct {
	Filter
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
}

type CursorFilter struct {
	Filter
	Cursor  *int64
	Limit   int
	SortBy  string
	SortDir string
}
