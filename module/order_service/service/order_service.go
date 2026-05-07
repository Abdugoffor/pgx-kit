package order_service

import (
	"context"
	"errors"
	"fmt"

	"pgx-kit/helper"
	order_dto "pgx-kit/module/order_service/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoCompany       = errors.New("user has no company")
	ErrEmptyFile       = errors.New("excel file is empty or has no data rows")
	ErrInvalidPrice    = errors.New("price must be greater than 0")
	ErrEmptyCart       = errors.New("cart is empty")
	ErrProductNotFound = errors.New("product not found")
	ErrNotEnoughStock  = errors.New("not enough stock")
)

type OrderService interface {
	Prihod(ctx context.Context, companyID, userID int64, rows []order_dto.ExcelRow) (*order_dto.PrihodResponse, error)
	Sotuv(ctx context.Context, companyID, userID int64, req order_dto.SotuvRequest) (*order_dto.SotuvResponse, error)
}

type orderService struct {
	db *pgxpool.Pool
}

func NewOrderService(db *pgxpool.Pool) OrderService {
	return &orderService{db: db}
}

func (s *orderService) Prihod(ctx context.Context, companyID, userID int64, rows []order_dto.ExcelRow) (*order_dto.PrihodResponse, error) {
	if len(rows) == 0 {
		return nil, ErrEmptyFile
	}

	tx, err := s.db.Begin(ctx)
	{
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
	}

	defer tx.Rollback(ctx)

	// 1. Umumiy summani hisoblash
	var totalSum float64
	{
		for _, row := range rows {
			count := row.Count
			if count < 1 {
				count = 1
			}
			totalSum += row.Price * row.Quantity * float64(count)
		}
	}

	// 2. Order yaratish
	var orderID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (company_id, user_id, type, total_sum)
		VALUES ($1, $2, 'prihod', $3)
		RETURNING id
	`, companyID, userID, totalSum).Scan(&orderID)

	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	productsCreated := 0
	lotsCreated := 0

	for _, row := range rows {
		slug := helper.Slug(row.Name)

		// 3. Product ni slug bo'yicha tekshirish
		var productID int64
		err = tx.QueryRow(ctx, `
			SELECT id FROM products
			WHERE slug = $1 AND company_id = $2
		`, slug, companyID).Scan(&productID)

		if errors.Is(err, pgx.ErrNoRows) {
			// Product yo'q — yaratamiz
			err = tx.QueryRow(ctx, `
				INSERT INTO products (company_id, name, slug, description, price, sell_price)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id
			`, companyID, row.Name, slug, row.Description, row.Price, row.SellPrice).Scan(&productID)
			if err != nil {
				return nil, fmt.Errorf("create product %q: %w", row.Name, err)
			}
			productsCreated++
		} else if err != nil {
			return nil, fmt.Errorf("find product %q: %w", row.Name, err)
		} else {
			// Product mavjud — narxini yangilaymiz
			_, err = tx.Exec(ctx, `
				UPDATE products
				SET price = $1, sell_price = $2, updated_at = NOW()
				WHERE id = $3 AND company_id = $4
			`, row.Price, row.SellPrice, productID, companyID)
			if err != nil {
				return nil, fmt.Errorf("update product price %q: %w", row.Name, err)
			}
		}

		// 4. Hozirgi umumiy zaxirani olish (barcha lotlar bo'yicha)
		var currentStock float64
		err = tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(quantity_after), 0)
			FROM product_value
			WHERE product_id = $1 AND company_id = $2
		`, productID, companyID).Scan(&currentStock)
		if err != nil {
			return nil, fmt.Errorf("get current stock for %q: %w", row.Name, err)
		}

		// 5. count bo'yicha product_value va product_history yozish
		count := row.Count
		if count < 1 {
			count = 1
		}

		for i := 0; i < count; i++ {
			// product_value — yangi lot
			var pvID int64
			err = tx.QueryRow(ctx, `
				INSERT INTO product_value (company_id, product_id, price, quantity_before, quantity_after, unit)
				VALUES ($1, $2, $3, $4, $5, $6)
				RETURNING id
			`, companyID, productID, row.Price, row.Quantity, row.Quantity, row.Unit).Scan(&pvID)
			if err != nil {
				return nil, fmt.Errorf("create product_value for %q: %w", row.Name, err)
			}

			// order_items
			_, err = tx.Exec(ctx, `
				INSERT INTO order_items (company_id, order_id, product_id, product_value_id, quantity, sale_price, discount)
				VALUES ($1, $2, $3, $4, $5, $6, 0)
			`, companyID, orderID, productID, pvID, row.Quantity, row.Price)
			if err != nil {
				return nil, fmt.Errorf("create order_item for %q: %w", row.Name, err)
			}

			// product_history — oldin qancha edi, qancha qo'shildi, keyin qancha bo'ldi
			quantityBefore := currentStock
			currentStock += row.Quantity
			quantityAfter := currentStock

			_, err = tx.Exec(ctx, `
				INSERT INTO product_history (company_id, user_id, product_id, product_value_id, order_id, order_type, quantity, quantity_before, quantity_after, price)
				VALUES ($1, $2, $3, $4, $5, 'prihod', $6, $7, $8, $9)
			`, companyID, userID, productID, pvID, orderID, row.Quantity, quantityBefore, quantityAfter, row.Price)
			if err != nil {
				return nil, fmt.Errorf("create product_history for %q: %w", row.Name, err)
			}

			lotsCreated++
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &order_dto.PrihodResponse{
		OrderID:         orderID,
		ProductsCreated: productsCreated,
		LotsCreated:     lotsCreated,
	}, nil
}

func (s *orderService) Sotuv(ctx context.Context, companyID, userID int64, req order_dto.SotuvRequest) (*order_dto.SotuvResponse, error) {
	if len(req.Items) == 0 {
		return nil, ErrEmptyCart
	}

	tx, err := s.db.Begin(ctx)
	{
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
	}

	defer tx.Rollback(ctx)

	// 1. Order yaratish (total_sum keyinroq yangilanadi)
	var orderID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (company_id, user_id, type, total_sum, note)
		VALUES ($1, $2, 'sotuv', 0, $3)
		RETURNING id
	`, companyID, userID, req.Note).Scan(&orderID)

	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	var totalSum, totalCost, totalProfit float64

	results := make([]order_dto.SotuvItemResult, 0, len(req.Items))

	for _, item := range req.Items {
		// 2. Productni olish
		var productName string
		var sellPrice float64
		err = tx.QueryRow(ctx, `
			SELECT name, sell_price FROM products
			WHERE id = $1 AND company_id = $2
		`, item.ProductID, companyID).Scan(&productName, &sellPrice)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, fmt.Errorf("%w: id=%d", ErrProductNotFound, item.ProductID)
			}
			return nil, fmt.Errorf("find product %d: %w", item.ProductID, err)
		}

		// 3. Hozirgi umumiy zaxirani olish
		var stockBefore float64
		err = tx.QueryRow(ctx, `
			SELECT COALESCE(SUM(quantity_after), 0)
			FROM product_value
			WHERE product_id = $1 AND company_id = $2
		`, item.ProductID, companyID).Scan(&stockBefore)

		if err != nil {
			return nil, fmt.Errorf("get stock for product %d: %w", item.ProductID, err)
		}

		if stockBefore < item.Quantity {
			return nil, fmt.Errorf("%w: %s (mavjud: %.3f, so'ralgan: %.3f)", ErrNotEnoughStock, productName, stockBefore, item.Quantity)
		}

		// 4. FIFO bo'yicha lotlardan ayirish (eng eski lot birinchi)
		rows, err := tx.Query(ctx, `
			SELECT id, price, quantity_after
			FROM product_value
			WHERE product_id = $1 AND company_id = $2 AND quantity_after > 0
			ORDER BY id ASC
		`, item.ProductID, companyID)

		if err != nil {
			return nil, fmt.Errorf("get lots for product %d: %w", item.ProductID, err)
		}

		remaining := item.Quantity

		var itemCost float64

		currentStock := stockBefore

		for rows.Next() && remaining > 0 {
			var lotID int64
			var lotPrice, lotQty float64
			if err := rows.Scan(&lotID, &lotPrice, &lotQty); err != nil {
				rows.Close()
				return nil, fmt.Errorf("scan lot: %w", err)
			}

			// Bu lotdan qancha olish mumkin
			take := remaining
			if take > lotQty {
				take = lotQty
			}

			newQty := lotQty - take
			remaining -= take
			itemCost += lotPrice * take

			// product_value ni yangilash
			_, err = tx.Exec(ctx, `
				UPDATE product_value SET quantity_after = $1 WHERE id = $2
			`, newQty, lotID)
			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("update lot %d: %w", lotID, err)
			}

			// order_items — har bir lotdan olingan miqdor uchun
			actualPrice := sellPrice - item.Discount
			if actualPrice < 0 {
				actualPrice = 0
			}
			_, err = tx.Exec(ctx, `
				INSERT INTO order_items (company_id, order_id, product_id, product_value_id, quantity, sale_price, discount)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, companyID, orderID, item.ProductID, lotID, take, sellPrice, item.Discount)
			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("create order_item: %w", err)
			}

			// product_history — har bir lot uchun alohida yozuv
			qtyBefore := currentStock
			currentStock -= take
			_, err = tx.Exec(ctx, `
				INSERT INTO product_history (company_id, user_id, product_id, product_value_id, order_id, order_type, quantity, quantity_before, quantity_after, price)
				VALUES ($1, $2, $3, $4, $5, 'sotuv', $6, $7, $8, $9)
			`, companyID, userID, item.ProductID, lotID, orderID, take, qtyBefore, currentStock, sellPrice)
			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("create product_history: %w", err)
			}
		}
		rows.Close()

		if remaining > 0 {
			return nil, fmt.Errorf("%w: %s", ErrNotEnoughStock, productName)
		}

		// Hisob-kitob
		revenueTotal := (sellPrice - item.Discount) * item.Quantity
		profit := revenueTotal - itemCost

		totalSum += revenueTotal
		totalCost += itemCost
		totalProfit += profit

		results = append(results, order_dto.SotuvItemResult{
			ProductID:    item.ProductID,
			ProductName:  productName,
			Quantity:     item.Quantity,
			SalePrice:    sellPrice,
			Discount:     item.Discount,
			CostTotal:    itemCost,
			RevenueTotal: revenueTotal,
			Profit:       profit,
			StockBefore:  stockBefore,
			StockAfter:   currentStock,
		})
	}

	// 5. Order total_sum ni yangilash
	_, err = tx.Exec(ctx, `UPDATE orders SET total_sum = $1 WHERE id = $2`, totalSum, orderID)
	{
		if err != nil {
			return nil, fmt.Errorf("update order total: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &order_dto.SotuvResponse{
		OrderID:     orderID,
		TotalSum:    totalSum,
		TotalCost:   totalCost,
		TotalProfit: totalProfit,
		ItemsSold:   len(results),
		Items:       results,
	}, nil
}
