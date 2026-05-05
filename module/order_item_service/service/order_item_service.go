package order_item_service

import (
	"context"

	order_item_dto "pgx-kit/module/order_item_service/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderItemService interface {
	Create(ctx context.Context, orderID int64, req order_item_dto.Create) (*order_item_dto.Response, error)
	Update(ctx context.Context, id int64, req order_item_dto.Update) (*order_item_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, orderID int64) ([]order_item_dto.Response, error)
}

type orderItemService struct {
	db *pgxpool.Pool
}

func NewOrderItemService(db *pgxpool.Pool) OrderItemService {
	return &orderItemService{db: db}
}

const recalcSQL = `
	UPDATE orders
	SET total_amount = (
		SELECT COALESCE(SUM(price * quantity), 0) FROM order_items WHERE order_id = $1
	),
	updated_at = now()
	WHERE id = $1
`

func (service *orderItemService) Create(ctx context.Context, orderID int64, req order_item_dto.Create) (*order_item_dto.Response, error) {
	tx, err := service.db.Begin(ctx)
	{
		if err != nil {
			return nil, err
		}
	}
	defer tx.Rollback(ctx)

	var (
		res       order_item_dto.Response
		productID int64
	)

	err = tx.QueryRow(ctx, `
		INSERT INTO order_items (order_id, product_id, quantity, price)
		SELECT $1, $2, $3, price
		FROM products
		WHERE id = $2
		RETURNING id, order_id, product_id, quantity, price::float8, created_at, updated_at
	`, orderID, req.ProductID, req.Quantity).
		Scan(&res.ID, &res.OrderID, &productID, &res.Quantity, &res.Price, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, recalcSQL, orderID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	err = service.db.QueryRow(ctx, `
		SELECT json_build_object('id', p.id, 'name', p.name, 'price', p.price::float8)
		FROM products p WHERE p.id = $1
	`, productID).Scan(&res.Product)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *orderItemService) Update(ctx context.Context, id int64, req order_item_dto.Update) (*order_item_dto.Response, error) {
	tx, err := service.db.Begin(ctx)
	{
		if err != nil {
			return nil, err
		}
	}
	defer tx.Rollback(ctx)

	var (
		res       order_item_dto.Response
		productID int64
	)

	err = tx.QueryRow(ctx, `
		UPDATE order_items
		SET quantity   = $1,
		    updated_at = now()
		WHERE id = $2
		RETURNING id, order_id, product_id, quantity, price::float8, created_at, updated_at
	`, req.Quantity, id).
		Scan(&res.ID, &res.OrderID, &productID, &res.Quantity, &res.Price, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, recalcSQL, res.OrderID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	err = service.db.QueryRow(ctx, `
		SELECT json_build_object('id', p.id, 'name', p.name, 'price', p.price::float8)
		FROM products p WHERE p.id = $1
	`, productID).Scan(&res.Product)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *orderItemService) Delete(ctx context.Context, id int64) error {
	tx, err := service.db.Begin(ctx)
	{
		if err != nil {
			return err
		}
	}
	defer tx.Rollback(ctx)

	var orderID int64

	err = tx.QueryRow(ctx, `
		DELETE FROM order_items WHERE id = $1 RETURNING order_id
	`, id).Scan(&orderID)

	if err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, recalcSQL, orderID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (service *orderItemService) List(ctx context.Context, orderID int64) ([]order_item_dto.Response, error) {
	rows, err := service.db.Query(ctx, `
		SELECT
			oi.id,
			oi.order_id,
			json_build_object('id', p.id, 'name', p.name, 'price', p.price::float8),
			oi.quantity,
			oi.price::float8,
			oi.created_at,
			oi.updated_at
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
		ORDER BY oi.id
	`, orderID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var items []order_item_dto.Response
	{
		for rows.Next() {
			var res order_item_dto.Response
			{
				if err := rows.Scan(
					&res.ID, &res.OrderID, &res.Product, &res.Quantity, &res.Price, &res.CreatedAt, &res.UpdatedAt,
				); err != nil {
					return nil, err
				}
			}
			items = append(items, res)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
