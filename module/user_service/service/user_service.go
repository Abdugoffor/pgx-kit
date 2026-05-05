package user_service

import (
	"context"
	"fmt"

	user_dto "pgx-kit/module/user_service/dto"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Create(ctx context.Context, req user_dto.Create) (*user_dto.Response, error)
	Update(ctx context.Context, id int64, req user_dto.Update) (*user_dto.Response, error)
	Delete(ctx context.Context, id int64) error
	Show(ctx context.Context, id int64) (*user_dto.Response, error)
	AdminList(ctx context.Context, filter user_dto.AdminFilter) ([]user_dto.Response, bool, error)
	CursorList(ctx context.Context, filter user_dto.CursorFilter) ([]user_dto.Response, bool, error)
}

type userService struct {
	db *pgxpool.Pool
}

func NewUserService(db *pgxpool.Pool) UserService {
	return &userService{db: db}
}

func hashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	{
		if err != nil {
			return "", err
		}
	}
	return string(b), nil
}

func (service *userService) Create(ctx context.Context, req user_dto.Create) (*user_dto.Response, error) {
	hashed, err := hashPassword(req.Password)
	{
		if err != nil {
			return nil, err
		}
	}

	role := req.Role
	{
		if role == "" {
			role = "user"
		}
	}

	var res user_dto.Response

	err = service.db.QueryRow(ctx, `
		INSERT INTO users (full_name, phone, password, photo, role, is_active)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, COALESCE($6, true))
		RETURNING id, full_name, phone, photo, role, is_active, created_at, updated_at
	`, req.FullName, req.Phone, hashed, req.Photo, role, req.IsActive).
		Scan(&res.ID, &res.FullName, &res.Phone, &res.Photo, &res.Role, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *userService) Update(ctx context.Context, id int64, req user_dto.Update) (*user_dto.Response, error) {
	var hashedPassword *string

	if req.Password != nil {
		h, err := hashPassword(*req.Password)
		if err != nil {
			return nil, err
		}
		hashedPassword = &h
	}

	var res user_dto.Response

	err := service.db.QueryRow(ctx, `
		UPDATE users
		SET full_name  = COALESCE($1, full_name),
		    phone      = COALESCE($2, phone),
		    password   = COALESCE($3, password),
		    photo      = COALESCE($4, photo),
		    role       = COALESCE($5, role),
		    is_active  = COALESCE($6, is_active),
		    updated_at = now()
		WHERE id = $7 AND deleted_at IS NULL
		RETURNING id, full_name, phone, photo, role, is_active, created_at, updated_at
	`, req.FullName, req.Phone, hashedPassword, req.Photo, req.Role, req.IsActive, id).
		Scan(&res.ID, &res.FullName, &res.Phone, &res.Photo, &res.Role, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *userService) Delete(ctx context.Context, id int64) error {
	_, err := service.db.Exec(ctx, `
		UPDATE users SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return err
}

func (service *userService) Show(ctx context.Context, id int64) (*user_dto.Response, error) {
	var res user_dto.Response

	err := service.db.QueryRow(ctx, `
		SELECT id, full_name, phone, photo, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&res.ID, &res.FullName, &res.Phone, &res.Photo, &res.Role, &res.IsActive, &res.CreatedAt, &res.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (service *userService) AdminList(ctx context.Context, filter user_dto.AdminFilter) ([]user_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT id, full_name, phone, photo, role, is_active, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
			AND full_name ILIKE '%%' || $1 || '%%'
			AND ($2 = '' OR phone ILIKE '%%' || $2 || '%%')
			AND ($3 = '' OR role = $3)
			AND ($4::boolean IS NULL OR is_active = $4)
		ORDER BY %s %s
		LIMIT $5 OFFSET $6
	`, filter.SortBy, filter.SortDir),
		filter.FullName, filter.Phone, filter.Role, filter.IsActive,
		filter.PageSize+1, (filter.Page-1)*filter.PageSize)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []user_dto.Response
	{
		for rows.Next() {
			var res user_dto.Response
			{
				if err := rows.Scan(&res.ID, &res.FullName, &res.Phone, &res.Photo, &res.Role, &res.IsActive, &res.CreatedAt, &res.UpdatedAt); err != nil {
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

func (service *userService) CursorList(ctx context.Context, filter user_dto.CursorFilter) ([]user_dto.Response, bool, error) {
	rows, err := service.db.Query(ctx, fmt.Sprintf(`
		SELECT id, full_name, phone, photo, role, is_active, created_at, updated_at
		FROM users
		WHERE deleted_at IS NULL
			AND full_name ILIKE '%%' || $1 || '%%'
			AND ($2 = '' OR phone ILIKE '%%' || $2 || '%%')
			AND ($3 = '' OR role = $3)
			AND ($4::boolean IS NULL OR is_active = $4)
			AND ($5::bigint IS NULL OR id > $5)
		ORDER BY %s %s, id %s
		LIMIT $6
	`, filter.SortBy, filter.SortDir, filter.SortDir),
		filter.FullName, filter.Phone, filter.Role, filter.IsActive, filter.Cursor, filter.Limit+1)

	if err != nil {
		return nil, false, err
	}

	defer rows.Close()

	var items []user_dto.Response
	{
		for rows.Next() {
			var res user_dto.Response
			{
				if err := rows.Scan(&res.ID, &res.FullName, &res.Phone, &res.Photo, &res.Role, &res.IsActive, &res.CreatedAt, &res.UpdatedAt); err != nil {
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
