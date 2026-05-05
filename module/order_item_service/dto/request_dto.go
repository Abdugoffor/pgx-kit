package order_item_dto

type Create struct {
	ProductID int64 `json:"product_id" validate:"required"`
	Quantity  int   `json:"quantity"   validate:"required,min=1"`
}

type Update struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}
