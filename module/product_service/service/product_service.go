package product_service

import (
	"context"
	"fmt"
	product_dto "pgx-kit/module/product_service/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductService interface {
	Create(ctx context.Context, req product_dto.Create) (int64, error)
	Update(ctx context.Context, id int64, req product_dto.Update) (int64, error)
	Delete(ctx context.Context, id int64) error
	Show(ctx context.Context, id int64) (*product_dto.Response, error)
	AdminList(ctx context.Context, filter product_dto.AdminFilter) ([]product_dto.Response, bool, error)
	CursorList(ctx context.Context, filter product_dto.CursorFilter) ([]product_dto.Response, bool, error)
}

type productService struct {
	db *pgxpool.Pool
}

func NewProductService(db *pgxpool.Pool) ProductService {
	return &productService{db: db}
}

func (service *productService) Create(ctx context.Context, req product_dto.Create) (int64, error) {
	var id int64

	err := service.db.QueryRow(ctx, `
		INSERT INTO products (name, description, price, category_id, is_active)
		VALUES ($1, $2, $3, $4, COALESCE($5, true))
		RETURNING id
	`, req.Name, req.Description, req.Price, req.CategoryID, req.IsActive).Scan(&id)
	{
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (service *productService) Update(ctx context.Context, id int64, req product_dto.Update) (int64, error) {
	var updatedID int64

	err := service.db.QueryRow(ctx, `
		UPDATE products
		SET name        = COALESCE($1, name),
		    description = COALESCE($2, description),
		    price       = COALESCE($3::numeric, price),
		    category_id = COALESCE($4, category_id),
		    is_active   = COALESCE($5, is_active),
		    updated_at  = now()
		WHERE id = $6
		RETURNING id
	`, req.Name, req.Description, req.Price, req.CategoryID, req.IsActive, id).Scan(&updatedID)
	{
		if err != nil {
			return 0, err
		}
	}

	return updatedID, nil
}

func (service *productService) Delete(ctx context.Context, id int64) error {
	_, err := service.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	return err
}

func (service *productService) Show(ctx context.Context, id int64) (*product_dto.Response, error) {
	var res product_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.description, p.price::float8, p.is_active, p.created_at, p.updated_at,
		       json_build_object('id', c.id, 'name', c.name, 'is_active', c.is_active)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1
	`, id).Scan(
		&res.ID, &res.Name, &res.Description, &res.Price, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
		&res.Category,
	)
	{
		if err != nil {
			return nil, err
		}
	}

	return &res, nil
}

func (service *productService) AdminList(ctx context.Context, filter product_dto.AdminFilter) ([]product_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.price::float8, p.is_active, p.created_at, p.updated_at,
		       json_build_object('id', c.id, 'name', c.name, 'is_active', c.is_active)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.name ILIKE '%%' || $1 || '%%'
			AND ($2::boolean IS NULL OR p.is_active = $2)
			AND ($3::bigint  IS NULL OR p.category_id = $3)
		ORDER BY p.%s %s
		LIMIT $4 OFFSET $5
	`, filter.SortBy, filter.SortDir), filter.Name, filter.IsActive, filter.CategoryID, filter.PageSize+1, (filter.Page-1)*filter.PageSize)

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var items []product_dto.Response
	{
		for rows.Next() {
			var res product_dto.Response
			{
				if err := rows.Scan(
					&res.ID, &res.Name, &res.Description, &res.Price, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
					&res.Category,
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

func (service *productService) CursorList(ctx context.Context, filter product_dto.CursorFilter) ([]product_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.price::float8, p.is_active, p.created_at, p.updated_at,
		       json_build_object('id', c.id, 'name', c.name, 'is_active', c.is_active)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.name ILIKE '%%' || $1 || '%%'
			AND ($2::boolean IS NULL OR p.is_active = $2)
			AND ($3::bigint  IS NULL OR p.category_id = $3)
			AND ($4::bigint IS NULL OR p.id > $4)
		ORDER BY p.%s %s, p.id %s
		LIMIT $5
	`, filter.SortBy, filter.SortDir, filter.SortDir), filter.Name, filter.IsActive, filter.CategoryID, filter.Cursor, filter.Limit+1)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []product_dto.Response
	{
		for rows.Next() {
			var res product_dto.Response
			{
				if err := rows.Scan(
					&res.ID, &res.Name, &res.Description, &res.Price, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
					&res.Category,
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
