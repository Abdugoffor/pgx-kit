package category_service

import (
	"context"
	"errors"
	"fmt"

	category_dto "pgx-kit/module/category_service/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoCompany = errors.New("user has no company")

type CategoryService interface {
	Create(ctx context.Context, companyID int64, req category_dto.Create) (*category_dto.Response, error)
	Update(ctx context.Context, companyID, id int64, req category_dto.Update) (*category_dto.Response, error)
	Delete(ctx context.Context, companyID, id int64) error
	Show(ctx context.Context, companyID, id int64) (*category_dto.Response, error)
	AdminList(ctx context.Context, companyID int64, filter category_dto.AdminFilter) ([]category_dto.Response, bool, error)
	CursorList(ctx context.Context, companyID int64, filter category_dto.CursorFilter) ([]category_dto.Response, bool, error)
}

type categoryService struct {
	db *pgxpool.Pool
}

func NewCategoryService(db *pgxpool.Pool) CategoryService {
	return &categoryService{db: db}
}

func (service *categoryService) Create(ctx context.Context, companyID int64, req category_dto.Create) (*category_dto.Response, error) {
	var res category_dto.Response

	err := service.db.QueryRow(ctx, `
		INSERT INTO categories (name, is_active, company_id)
		VALUES ($1, COALESCE($2, true), $3)
		RETURNING id, name, is_active, created_at, updated_at
	`, req.Name, req.IsActive, companyID).Scan(&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *categoryService) Update(ctx context.Context, companyID, id int64, req category_dto.Update) (*category_dto.Response, error) {
	var res category_dto.Response

	err := service.db.QueryRow(ctx, `
		UPDATE categories
		SET name       = COALESCE($1, name),
		    is_active  = COALESCE($2, is_active),
		    updated_at = now()
		WHERE id = $3 AND company_id = $4
		RETURNING id, name, is_active, created_at, updated_at
	`, req.Name, req.IsActive, id, companyID).Scan(&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *categoryService) Delete(ctx context.Context, companyID, id int64) error {
	tag, err := service.db.Exec(ctx, `
		DELETE FROM categories WHERE id = $1 AND company_id = $2
	`, id, companyID)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (service *categoryService) Show(ctx context.Context, companyID, id int64) (*category_dto.Response, error) {
	var res category_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT id, name, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1 AND company_id = $2
	`, id, companyID).Scan(&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *categoryService) AdminList(ctx context.Context, companyID int64, filter category_dto.AdminFilter) ([]category_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name, is_active, created_at, updated_at
		FROM categories
		WHERE company_id = $1
			AND name ILIKE '%%' || $2 || '%%'
			AND ($3::boolean IS NULL OR is_active = $3)
		ORDER BY %s %s
		LIMIT $4 OFFSET $5
	`, filter.SortBy, filter.SortDir), companyID, filter.Name, filter.IsActive, filter.PageSize+1, (filter.Page-1)*filter.PageSize)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []category_dto.Response
	{
		for rows.Next() {
			var res category_dto.Response
			{
				if err := rows.Scan(&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt); err != nil {
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

func (service *categoryService) CursorList(ctx context.Context, companyID int64, filter category_dto.CursorFilter) ([]category_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name, is_active, created_at, updated_at
		FROM categories
		WHERE company_id = $1
			AND name ILIKE '%%' || $2 || '%%'
			AND ($3::boolean IS NULL OR is_active = $3)
			AND ($4::bigint IS NULL OR id > $4)
		ORDER BY %s %s, id %s
		LIMIT $5
	`, filter.SortBy, filter.SortDir, filter.SortDir), companyID, filter.Name, filter.IsActive, filter.Cursor, filter.Limit+1)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []category_dto.Response
	{
		for rows.Next() {
			var res category_dto.Response
			{
				if err := rows.Scan(&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt); err != nil {
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
