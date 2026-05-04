package language_dto

type Create struct {
	Name        string  `json:"name"        validate:"required,min=2,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active"`
}

type Update struct {
	Name        *string `json:"name"        validate:"omitempty,min=2,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active"`
}

type Filter struct {
	Name        string
	Description string
	IsActive    *bool
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
