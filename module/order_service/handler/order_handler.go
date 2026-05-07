package order_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"github.com/xuri/excelize/v2"

	"pgx-kit/helper"
	"pgx-kit/middleware"
	order_dto "pgx-kit/module/order_service/dto"
	order_service "pgx-kit/module/order_service/service"
)

var excelFormats = []string{".xlsx", ".xls"}

const maxExcelSize int64 = 20 // MB

func requireCompany(w http.ResponseWriter, r *http.Request) (int64, bool) {
	companyID := middleware.CompanyID(r)
	{
		if companyID == 0 {
			helper.JSON(w, http.StatusForbidden, map[string]string{"error": order_service.ErrNoCompany.Error()})
			return 0, false
		}
	}
	return companyID, true
}

type orderHandler struct {
	service order_service.OrderService
}

func NewOrderHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &orderHandler{service: order_service.NewOrderService(db)}

	routes := group + "/orders"
	{
		router.POST(routes+"/prihod", middleware.CheckRole(handler.Prihod, "admin", "user"))
		router.POST(routes+"/sotuv", middleware.CheckRole(handler.Sotuv, "admin", "user"))
	}
}

func (h *orderHandler) Prihod(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	companyID, ok := requireCompany(w, r)
	{
		if !ok {
			return
		}
	}

	userID := middleware.UserID(r)

	// Excel faylni o'qish
	if err := r.ParseMultipartForm(maxExcelSize << 20); err != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "file too large or invalid form"})
		return
	}

	file, header, err := r.FormFile("file")
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "file field is required"})
			return
		}
	}

	defer file.Close()

	// Format tekshirish
	ext := strings.ToLower(header.Filename[strings.LastIndex(header.Filename, "."):])

	isExcel := false

	for _, f := range excelFormats {
		if ext == f {
			isExcel = true
			break
		}
	}

	if !isExcel {
		helper.JSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "only .xlsx and .xls files are allowed"})
		return
	}

	f, err := excelize.OpenReader(file)
	{
		if err != nil {
			helper.JSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "cannot read excel file"})
			return
		}
	}

	defer f.Close()

	sheetName := f.GetSheetName(0)

	excelRows, err := f.GetRows(sheetName)
	{
		if err != nil || len(excelRows) < 2 {
			helper.JSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "excel file is empty or has no data rows"})
			return
		}
	}

	// Header qatorini map qilish (index -> column name)
	headerRow := excelRows[0]

	colIndex := make(map[string]int)

	for i, cell := range headerRow {
		colIndex[strings.TrimSpace(strings.ToLower(cell))] = i
	}

	// Kerakli ustunlar bor-yo'qligini tekshirish
	required := []string{"name", "price", "sell_price"}
	{
		for _, col := range required {

			if _, exists := colIndex[col]; !exists {
				helper.JSON(w, http.StatusUnprocessableEntity, map[string]string{
					"error": "missing required column: " + col,
				})

				return
			}
		}
	}

	// Ma'lumotlarni parse qilish
	var rows []order_dto.ExcelRow

	for i := 1; i < len(excelRows); i++ {
		cells := excelRows[i]

		name := cellValue(cells, colIndex, "name")

		if name == "" {
			continue // bo'sh qatorni o'tkazib yuborish
		}

		price := parseFloat(cellValue(cells, colIndex, "price"))

		sellPrice := parseFloat(cellValue(cells, colIndex, "sell_price"))

		if price <= 0 || sellPrice <= 0 {
			helper.JSON(w, http.StatusUnprocessableEntity, map[string]string{
				"error": "invalid price at row " + strconv.Itoa(i+1),
			})
			return
		}

		quantity := parseFloat(cellValue(cells, colIndex, "quantity"))

		if quantity <= 0 {
			quantity = 1
		}

		count := parseInt(cellValue(cells, colIndex, "count"))

		rows = append(rows, order_dto.ExcelRow{
			Name:        name,
			Description: cellValue(cells, colIndex, "description"),
			Price:       price,
			SellPrice:   sellPrice,
			Unit:        cellValue(cells, colIndex, "unit"),
			Quantity:    quantity,
			Count:       count,
		})
	}

	if len(rows) == 0 {
		helper.JSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "no valid data rows found"})
		return
	}

	result, err := h.service.Prihod(r.Context(), companyID, userID, rows)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusCreated, result)
}

func (h *orderHandler) Sotuv(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	companyID, ok := requireCompany(w, r)
	if !ok {
		return
	}

	userID := middleware.UserID(r)

	var req order_dto.SotuvRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	result, err := h.service.Sotuv(r.Context(), companyID, userID, req)
	if err != nil {
		switch {
		case errors.Is(err, order_service.ErrEmptyCart):
			helper.JSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		case errors.Is(err, order_service.ErrProductNotFound):
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		case errors.Is(err, order_service.ErrNotEnoughStock):
			helper.JSON(w, http.StatusConflict, map[string]string{"error": err.Error()})
		default:
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return
	}

	helper.JSON(w, http.StatusCreated, result)
}

func cellValue(cells []string, colIndex map[string]int, col string) string {
	idx, ok := colIndex[col]

	if !ok || idx >= len(cells) {
		return ""
	}

	return strings.TrimSpace(cells[idx])
}

func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", ".")

	v, _ := strconv.ParseFloat(s, 64)

	return v
}

func parseInt(s string) int {
	v, _ := strconv.Atoi(s)

	return v
}
