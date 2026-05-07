package product_dto

type Create struct {
	Name        string  `json:"name"        validate:"required,min=2,max=255"`
	Slug        *string `json:"slug"        validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Price       float64 `json:"price"       validate:"required,gt=0"`
	SellPrice   float64 `json:"sell_price"  validate:"required,gt=0"`
	CategoryID  *int64  `json:"category_id" validate:"omitempty,gt=0"`
	IsActive    *bool   `json:"is_active"`
	Photo       *string `json:"photo"`
}

type Update struct {
	Name        *string  `json:"name"        validate:"omitempty,min=2,max=255"`
	Description *string  `json:"description" validate:"omitempty,max=1000"`
	Price       *float64 `json:"price"       validate:"omitempty,gt=0"`
	SellPrice   *float64 `json:"sell_price"  validate:"omitempty,gt=0"`
	CategoryID  *int64   `json:"category_id" validate:"omitempty,gt=0"`
	IsActive    *bool    `json:"is_active"`
	Photo       *string  `json:"photo"`
}

type Filter struct {
	Name       string
	IsActive   *bool
	CategoryID *int64
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
