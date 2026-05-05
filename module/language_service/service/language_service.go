package language_service

import (
	"context"
	"fmt"
	language_dto "pgx-kit/module/language_service/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type LanguageService interface {
	Create(ctx context.Context, req language_dto.Create) (*language_dto.Response, error)
	Update(ctx context.Context, id int64, req language_dto.Update) (*language_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	Show(ctx context.Context, id int64) (*language_dto.Response, error)
	AdminList(ctx context.Context, filter language_dto.AdminFilter) ([]language_dto.Response, bool, error)
	CursorList(ctx context.Context, filter language_dto.CursorFilter) ([]language_dto.Response, bool, error)
}

type languageService struct {
	db *pgxpool.Pool
}

func NewLanguageService(db *pgxpool.Pool) LanguageService {
	return &languageService{db: db}
}

func (service *languageService) Create(ctx context.Context, req language_dto.Create) (*language_dto.Response, error) {
	var res language_dto.Response

	err := service.db.QueryRow(ctx, `
		INSERT INTO languages (name, description, is_active)
		VALUES ($1, $2, COALESCE($3, true))
		RETURNING id, name, description, is_active, created_at, updated_at
	`, req.Name, req.Description, req.IsActive).Scan(&res.ID, &res.Name, &res.Description, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *languageService) Update(ctx context.Context, id int64, req language_dto.Update) (*language_dto.Response, error) {
	var res language_dto.Response

	err := service.db.QueryRow(ctx, `
		UPDATE languages
		SET name = COALESCE($1, name),
			description = COALESCE($2, description),
			is_active = COALESCE($3, is_active),
			updated_at = now()
		WHERE id = $4
		RETURNING id, name, description, is_active, created_at, updated_at
	`, req.Name, req.Description, req.IsActive, id).Scan(&res.ID, &res.Name, &res.Description, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *languageService) Delete(ctx context.Context, id int64) error {
	_, err := service.db.Exec(ctx, `DELETE FROM languages WHERE id = $1`, id)
	return err
}

func (service *languageService) Show(ctx context.Context, id int64) (*language_dto.Response, error) {
	var res language_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM languages
		WHERE id = $1
	`, id).Scan(&res.ID, &res.Name, &res.Description, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *languageService) AdminList(ctx context.Context, filter language_dto.AdminFilter) ([]language_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name, description, is_active, created_at, updated_at
		FROM languages
		WHERE name ILIKE '%%' || $1 || '%%'
			AND description ILIKE '%%' || $2 || '%%'
			AND ($3::boolean IS NULL OR is_active = $3)
		ORDER BY %s %s
		LIMIT $4 OFFSET $5
	`, filter.SortBy, filter.SortDir), filter.Name, filter.Description, filter.IsActive, filter.PageSize+1, (filter.Page-1)*filter.PageSize)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []language_dto.Response
	{
		for rows.Next() {
			var res language_dto.Response
			{
				if err := rows.Scan(&res.ID, &res.Name, &res.Description, &res.IsActive, &res.CreatedAt, &res.UpdatedAt); err != nil {
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

func (service *languageService) CursorList(ctx context.Context, filter language_dto.CursorFilter) ([]language_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT id, name, description, is_active, created_at, updated_at
		FROM languages
		WHERE name ILIKE '%%' || $1 || '%%'
			AND description ILIKE '%%' || $2 || '%%'
			AND ($3::boolean IS NULL OR is_active = $3)
			AND ($4::bigint IS NULL OR id > $4)
		ORDER BY %s %s, id %s
		LIMIT $5
	`, filter.SortBy, filter.SortDir, filter.SortDir), filter.Name, filter.Description, filter.IsActive, filter.Cursor, filter.Limit+1)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []language_dto.Response
	{
		for rows.Next() {
			var res language_dto.Response
			{
				if err := rows.Scan(&res.ID, &res.Name, &res.Description, &res.IsActive, &res.CreatedAt, &res.UpdatedAt); err != nil {
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
