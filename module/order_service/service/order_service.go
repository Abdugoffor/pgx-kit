package order_service

import (
	"context"
	"fmt"

	order_dto "pgx-kit/module/order_service/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderService interface {
	Create(ctx context.Context, userID int64, req order_dto.Create) (*order_dto.Response, error)
	Update(ctx context.Context, id int64, req order_dto.Update) (*order_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	Show(ctx context.Context, id int64) (*order_dto.ShowResponse, error)
	AdminList(ctx context.Context, filter order_dto.AdminFilter) ([]order_dto.Response, bool, error)
	CursorList(ctx context.Context, filter order_dto.CursorFilter) ([]order_dto.Response, bool, error)
}

type orderService struct {
	db *pgxpool.Pool
}

func NewOrderService(db *pgxpool.Pool) OrderService {
	return &orderService{db: db}
}

const userJSON = `json_build_object('id', u.id, 'full_name', u.full_name, 'phone', u.phone)`

func (service *orderService) Create(ctx context.Context, userID int64, req order_dto.Create) (*order_dto.Response, error) {
	var res order_dto.Response

	err := service.db.QueryRow(ctx, `
		INSERT INTO orders (user_id, note)
		VALUES ($1, $2)
		RETURNING id, status, total_amount::float8, note, created_at, updated_at
	`, userID, req.Note).
		Scan(&res.ID, &res.Status, &res.TotalAmount, &res.Note, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	err = service.db.QueryRow(ctx,
		`SELECT `+userJSON+` FROM users WHERE id = $1`, userID,
	).Scan(&res.User)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *orderService) Update(ctx context.Context, id int64, req order_dto.Update) (*order_dto.Response, error) {
	var res order_dto.Response

	err := service.db.QueryRow(ctx, `
		UPDATE orders
		SET status     = COALESCE($1, status),
		    note       = COALESCE($2, note),
		    updated_at = now()
		WHERE id = $3 AND deleted_at IS NULL
		RETURNING id, user_id, status, total_amount::float8, note, created_at, updated_at
	`, req.Status, req.Note, id).
		Scan(&res.ID, new(int64), &res.Status, &res.TotalAmount, &res.Note, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	err = service.db.QueryRow(ctx,
		`SELECT `+userJSON+` FROM users u JOIN orders o ON o.user_id = u.id WHERE o.id = $1`, id,
	).Scan(&res.User)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *orderService) Delete(ctx context.Context, id int64) error {
	_, err := service.db.Exec(ctx, `
		UPDATE orders SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return err
}

func (service *orderService) Show(ctx context.Context, id int64) (*order_dto.ShowResponse, error) {
	var res order_dto.ShowResponse

	err := service.db.QueryRow(ctx, `
		SELECT
			o.id,
			`+userJSON+`,
			o.status,
			o.total_amount::float8,
			o.note,
			COALESCE(
				json_agg(
					json_build_object(
						'id',       oi.id,
						'product',  json_build_object('id', p.id, 'name', p.name, 'price', p.price::float8),
						'quantity', oi.quantity,
						'price',    oi.price::float8
					) ORDER BY oi.id
				) FILTER (WHERE oi.id IS NOT NULL),
				'[]'::json
			),
			o.created_at,
			o.updated_at
		FROM orders o
		JOIN users u ON u.id = o.user_id
		LEFT JOIN order_items oi ON oi.order_id = o.id
		LEFT JOIN products p ON p.id = oi.product_id
		WHERE o.id = $1 AND o.deleted_at IS NULL
		GROUP BY o.id, u.id
	`, id).Scan(
		&res.ID, &res.User, &res.Status, &res.TotalAmount, &res.Note,
		&res.Items, &res.CreatedAt, &res.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *orderService) AdminList(ctx context.Context, filter order_dto.AdminFilter) ([]order_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT o.id, `+userJSON+`, o.status, o.total_amount::float8, o.note, o.created_at, o.updated_at
		FROM orders o
		JOIN users u ON u.id = o.user_id
		WHERE o.deleted_at IS NULL
			AND ($1 = '' OR o.status = $1)
			AND ($2::bigint IS NULL OR o.user_id = $2)
		ORDER BY o.%s %s
		LIMIT $3 OFFSET $4
	`, filter.SortBy, filter.SortDir),
		filter.Status, filter.UserID, filter.PageSize+1, (filter.Page-1)*filter.PageSize)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []order_dto.Response
	{
		for rows.Next() {
			var res order_dto.Response
			{
				if err := rows.Scan(
					&res.ID, &res.User, &res.Status, &res.TotalAmount, &res.Note, &res.CreatedAt, &res.UpdatedAt,
				); err != nil {
					return nil, false, err
				}
			}
			items = append(items, res)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasNext := len(items) > filter.PageSize
	{
		if hasNext {
			items = items[:filter.PageSize]
		}
	}

	return items, hasNext, nil
}

func (service *orderService) CursorList(ctx context.Context, filter order_dto.CursorFilter) ([]order_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT o.id, `+userJSON+`, o.status, o.total_amount::float8, o.note, o.created_at, o.updated_at
		FROM orders o
		JOIN users u ON u.id = o.user_id
		WHERE o.deleted_at IS NULL
			AND ($1 = '' OR o.status = $1)
			AND ($2::bigint IS NULL OR o.user_id = $2)
			AND ($3::bigint IS NULL OR o.id > $3)
		ORDER BY o.%s %s, o.id %s
		LIMIT $4
	`, filter.SortBy, filter.SortDir, filter.SortDir),
		filter.Status, filter.UserID, filter.Cursor, filter.Limit+1)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []order_dto.Response
	{
		for rows.Next() {
			var res order_dto.Response
			{
				if err := rows.Scan(
					&res.ID, &res.User, &res.Status, &res.TotalAmount, &res.Note, &res.CreatedAt, &res.UpdatedAt,
				); err != nil {
					return nil, false, err
				}
			}
			items = append(items, res)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(items) > filter.Limit
	{
		if hasMore {
			items = items[:filter.Limit]
		}
	}

	return items, hasMore, nil
}
