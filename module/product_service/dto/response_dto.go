package product_dto

import (
	"encoding/json"
	"time"
)

type Response struct {
	ID          int64           `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	Price       float64         `json:"price"`
	SellPrice   float64         `json:"sell_price"`
	Category    json.RawMessage `json:"category"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type AdminListResponse struct {
	Data     []Response `json:"data"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	HasNext  bool       `json:"has_next"`
	HasPrev  bool       `json:"has_prev"`
}

type CursorListResponse struct {
	Data       []Response `json:"data"`
	NextCursor *int64     `json:"next_cursor"`
	HasMore    bool       `json:"has_more"`
}
