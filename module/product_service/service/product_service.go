package product_service

import (
	"context"
	"errors"
	"fmt"

	product_dto "pgx-kit/module/product_service/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNoCompany       = errors.New("user has no company")
	ErrCategoryInvalid = errors.New("category not found in your company")
)

type ProductService interface {
	Create(ctx context.Context, companyID int64, req product_dto.Create) (int64, error)
	Update(ctx context.Context, companyID, id int64, req product_dto.Update) (int64, error)
	Delete(ctx context.Context, companyID, id int64) error
	Show(ctx context.Context, companyID, id int64) (*product_dto.Response, error)
	AdminList(ctx context.Context, companyID int64, filter product_dto.AdminFilter) ([]product_dto.Response, bool, error)
	CursorList(ctx context.Context, companyID int64, filter product_dto.CursorFilter) ([]product_dto.Response, bool, error)
}

type productService struct {
	db *pgxpool.Pool
}

func NewProductService(db *pgxpool.Pool) ProductService {
	return &productService{db: db}
}

func (service *productService) ensureCategoryInCompany(ctx context.Context, categoryID, companyID int64) error {
	var ok int

	err := service.db.QueryRow(ctx, `
		SELECT 1 FROM categories WHERE id = $1 AND company_id = $2
	`, categoryID, companyID).Scan(&ok)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCategoryInvalid
		}
		return err
	}

	return nil
}

func (service *productService) Create(ctx context.Context, companyID int64, req product_dto.Create) (int64, error) {
	if err := service.ensureCategoryInCompany(ctx, req.CategoryID, companyID); err != nil {
		return 0, err
	}

	var id int64

	err := service.db.QueryRow(ctx, `
		INSERT INTO products (name, description, cost_price, sell_price, category_id, is_active, company_id)
		VALUES ($1, $2, $3, $4, $5, COALESCE($6, true), $7)
		RETURNING id
	`, req.Name, req.Description, req.CostPrice, req.SellPrice, req.CategoryID, req.IsActive, companyID).Scan(&id)
	{
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (service *productService) Update(ctx context.Context, companyID, id int64, req product_dto.Update) (int64, error) {
	if req.CategoryID != nil {
		if err := service.ensureCategoryInCompany(ctx, *req.CategoryID, companyID); err != nil {
			return 0, err
		}
	}

	var updatedID int64

	err := service.db.QueryRow(ctx, `
		UPDATE products
		SET name        = COALESCE($1, name),
		    description = COALESCE($2, description),
		    cost_price  = COALESCE($3::numeric, cost_price),
		    sell_price  = COALESCE($4::numeric, sell_price),
		    category_id = COALESCE($5, category_id),
		    is_active   = COALESCE($6, is_active),
		    updated_at  = now()
		WHERE id = $7 AND company_id = $8
		RETURNING id
	`, req.Name, req.Description, req.CostPrice, req.SellPrice, req.CategoryID, req.IsActive, id, companyID).Scan(&updatedID)
	{
		if err != nil {
			return 0, err
		}
	}

	return updatedID, nil
}

func (service *productService) Delete(ctx context.Context, companyID, id int64) error {
	tag, err := service.db.Exec(ctx, `
		DELETE FROM products WHERE id = $1 AND company_id = $2
	`, id, companyID)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (service *productService) Show(ctx context.Context, companyID, id int64) (*product_dto.Response, error) {
	var res product_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT p.id, p.name, p.description, p.cost_price::float8, p.sell_price::float8, p.is_active, p.created_at, p.updated_at,
		       json_build_object('id', c.id, 'name', c.name, 'is_active', c.is_active)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.id = $1 AND p.company_id = $2
	`, id, companyID).Scan(
		&res.ID, &res.Name, &res.Description, &res.CostPrice, &res.SellPrice, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
		&res.Category,
	)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *productService) AdminList(ctx context.Context, companyID int64, filter product_dto.AdminFilter) ([]product_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.cost_price::float8, p.sell_price::float8, p.is_active, p.created_at, p.updated_at,
		       json_build_object('id', c.id, 'name', c.name, 'is_active', c.is_active)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.company_id = $1
			AND p.name ILIKE '%%' || $2 || '%%'
			AND ($3::boolean IS NULL OR p.is_active = $3)
			AND ($4::bigint  IS NULL OR p.category_id = $4)
		ORDER BY p.%s %s
		LIMIT $5 OFFSET $6
	`, filter.SortBy, filter.SortDir), companyID, filter.Name, filter.IsActive, filter.CategoryID, filter.PageSize+1, (filter.Page-1)*filter.PageSize)

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
					&res.ID, &res.Name, &res.Description, &res.CostPrice, &res.SellPrice, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
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

func (service *productService) CursorList(ctx context.Context, companyID int64, filter product_dto.CursorFilter) ([]product_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.cost_price::float8, p.sell_price::float8, p.is_active, p.created_at, p.updated_at,
		       json_build_object('id', c.id, 'name', c.name, 'is_active', c.is_active)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		WHERE p.company_id = $1
			AND p.name ILIKE '%%' || $2 || '%%'
			AND ($3::boolean IS NULL OR p.is_active = $3)
			AND ($4::bigint  IS NULL OR p.category_id = $4)
			AND ($5::bigint IS NULL OR p.id > $5)
		ORDER BY p.%s %s, p.id %s
		LIMIT $6
	`, filter.SortBy, filter.SortDir, filter.SortDir), companyID, filter.Name, filter.IsActive, filter.CategoryID, filter.Cursor, filter.Limit+1)

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
					&res.ID, &res.Name, &res.Description, &res.CostPrice, &res.SellPrice, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
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
