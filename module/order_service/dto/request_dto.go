package order_dto

// ExcelRow — bir qator Excel fayldan o'qilgan ma'lumot
type ExcelRow struct {
	Name        string
	Description string
	Price       float64
	SellPrice   float64
	Unit        string
	Quantity    float64
	Count       int
}

// PrihodResponse — prihod natijasi
type PrihodResponse struct {
	OrderID         int64 `json:"order_id"`
	ProductsCreated int   `json:"products_created"`
	LotsCreated     int   `json:"lots_created"`
}

// ─── Sotuv (sale) ───

// SotuvItem — savatchadagi bitta tovar
type SotuvItem struct {
	ProductID int64   `json:"product_id" validate:"required,gt=0"`
	Quantity  float64 `json:"quantity"   validate:"required,gt=0"`
	Discount  float64 `json:"discount"   validate:"gte=0"` // chegirma summasi (0 = chegirmasiz)
}

// SotuvRequest — kassadan keladigan savatcha
type SotuvRequest struct {
	Items []SotuvItem `json:"items" validate:"required,min=1,dive"`
	Note  string      `json:"note"`
}

// SotuvItemResult — bitta tovar bo'yicha natija
type SotuvItemResult struct {
	ProductID     int64   `json:"product_id"`
	ProductName   string  `json:"product_name"`
	Quantity      float64 `json:"quantity"`
	SalePrice     float64 `json:"sale_price"`
	Discount      float64 `json:"discount"`
	CostTotal     float64 `json:"cost_total"`
	RevenueTotal  float64 `json:"revenue_total"`
	Profit        float64 `json:"profit"`
	StockBefore   float64 `json:"stock_before"`
	StockAfter    float64 `json:"stock_after"`
}

// SotuvResponse — sotuv natijasi
type SotuvResponse struct {
	OrderID      int64             `json:"order_id"`
	TotalSum     float64           `json:"total_sum"`
	TotalCost    float64           `json:"total_cost"`
	TotalProfit  float64           `json:"total_profit"`
	ItemsSold    int               `json:"items_sold"`
	Items        []SotuvItemResult `json:"items"`
}
