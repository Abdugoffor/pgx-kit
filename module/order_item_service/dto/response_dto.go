package order_item_dto

import (
	"encoding/json"
	"time"
)

type Response struct {
	ID        int64           `json:"id"`
	OrderID   int64           `json:"order_id"`
	Product   json.RawMessage `json:"product"`
	Quantity  int             `json:"quantity"`
	Price     float64         `json:"price"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
