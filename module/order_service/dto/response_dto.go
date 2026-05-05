package order_dto

import (
	"encoding/json"
	"time"
)

type Response struct {
	ID          int64           `json:"id"`
	User        json.RawMessage `json:"user"`
	Status      string          `json:"status"`
	TotalAmount float64         `json:"total_amount"`
	Note        *string         `json:"note"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type ShowResponse struct {
	ID          int64           `json:"id"`
	User        json.RawMessage `json:"user"`
	Status      string          `json:"status"`
	TotalAmount float64         `json:"total_amount"`
	Note        *string         `json:"note"`
	Items       json.RawMessage `json:"items"`
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
