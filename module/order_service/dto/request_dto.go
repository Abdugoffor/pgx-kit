package order_dto

type Create struct {
	Note *string `json:"note" validate:"omitempty,max=1000"`
}

type Update struct {
	Status *string `json:"status" validate:"omitempty,oneof=pending confirmed shipped delivered cancelled"`
	Note   *string `json:"note"   validate:"omitempty,max=1000"`
}

type Filter struct {
	Status string
	UserID *int64
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
