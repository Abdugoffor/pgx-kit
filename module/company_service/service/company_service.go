package company_service

import (
	"context"
	"fmt"

	"pgx-kit/helper"
	company_dto "pgx-kit/module/company_service/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CompanyService interface {
	Create(ctx context.Context, userID int64, role string, req company_dto.Create) (*company_dto.CreateResponse, error)
	Update(ctx context.Context, id int64, req company_dto.Update) (*company_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	Show(ctx context.Context, id int64) (*company_dto.Response, error)
	AdminList(ctx context.Context, filter company_dto.AdminFilter) ([]company_dto.Response, bool, error)
	CursorList(ctx context.Context, filter company_dto.CursorFilter) ([]company_dto.Response, bool, error)
}

type companyService struct {
	db *pgxpool.Pool
}

func NewCompanyService(db *pgxpool.Pool) CompanyService {
	return &companyService{db: db}
}

func (service *companyService) Create(ctx context.Context, userID int64, role string, req company_dto.Create) (*company_dto.CreateResponse, error) {
	tx, err := service.db.BeginTx(ctx, pgx.TxOptions{})
	{
		if err != nil {
			return nil, err
		}
	}
	defer tx.Rollback(ctx)

	var company company_dto.Response

	err = tx.QueryRow(ctx, `
		INSERT INTO companys (name, is_active)
		VALUES ($1, COALESCE($2, true))
		RETURNING id, name, is_active, created_at, updated_at
	`, req.Name, req.IsActive).Scan(&company.ID, &company.Name, &company.IsActive, &company.CreatedAt, &company.UpdatedAt)

	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO company_users (company_id, user_id)
		VALUES ($1, $2)
	`, company.ID, userID)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	token, err := helper.GenerateToken(userID, role, company.ID)
	{
		if err != nil {
			return nil, err
		}
	}

	return &company_dto.CreateResponse{Company: company, AccessToken: token}, nil
}

func (service *companyService) Update(ctx context.Context, id int64, req company_dto.Update) (*company_dto.Response, error) {
	var res company_dto.Response

	err := service.db.QueryRow(ctx, `
		UPDATE companys
		SET name       = COALESCE($1, name),
		    is_active  = COALESCE($2, is_active),
		    updated_at = now()
		WHERE id = $3
		RETURNING id, name, is_active, created_at, updated_at
	`, req.Name, req.IsActive, id).Scan(&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *companyService) Delete(ctx context.Context, id int64) error {
	_, err := service.db.Exec(ctx, `DELETE FROM companys WHERE id = $1`, id)
	return err
}

func (service *companyService) Show(ctx context.Context, id int64) (*company_dto.Response, error) {
	var res company_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT id, name, is_active, created_at, updated_at
		FROM companys
		WHERE id = $1
	`, id).Scan(&res.ID, &res.Name, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *companyService) AdminList(ctx context.Context, filter company_dto.AdminFilter) ([]company_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name, is_active, created_at, updated_at
		FROM companys
		WHERE name ILIKE '%%' || $1 || '%%'
			AND ($2::boolean IS NULL OR is_active = $2)
		ORDER BY %s %s
		LIMIT $3 OFFSET $4
	`, filter.SortBy, filter.SortDir), filter.Name, filter.IsActive, filter.PageSize+1, (filter.Page-1)*filter.PageSize)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []company_dto.Response
	{
		for rows.Next() {
			var res company_dto.Response
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

func (service *companyService) CursorList(ctx context.Context, filter company_dto.CursorFilter) ([]company_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name, is_active, created_at, updated_at
		FROM companys
		WHERE name ILIKE '%%' || $1 || '%%'
			AND ($2::boolean IS NULL OR is_active = $2)
			AND ($3::bigint IS NULL OR id > $3)
		ORDER BY %s %s, id %s
		LIMIT $4
	`, filter.SortBy, filter.SortDir, filter.SortDir), filter.Name, filter.IsActive, filter.Cursor, filter.Limit+1)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []company_dto.Response
	{
		for rows.Next() {
			var res company_dto.Response
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
