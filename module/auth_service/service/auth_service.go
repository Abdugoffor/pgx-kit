package auth_service

import (
	"context"
	"errors"

	"pgx-kit/helper"
	auth_dto "pgx-kit/module/auth_service/dto"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid phone or password")

type AuthService interface {
	Register(ctx context.Context, req auth_dto.Register) (*auth_dto.TokenResponse, error)
	Login(ctx context.Context, req auth_dto.Login) (*auth_dto.TokenResponse, error)
}

type authService struct {
	db *pgxpool.Pool
}

func NewAuthService(db *pgxpool.Pool) AuthService {
	return &authService{db: db}
}

func (service *authService) lookupCompanyID(ctx context.Context, userID int64) (int64, error) {
	var id int64

	err := service.db.QueryRow(ctx, `
		SELECT company_id FROM company_users WHERE user_id = $1 LIMIT 1
	`, userID).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return id, nil
}

func (service *authService) Register(ctx context.Context, req auth_dto.Register) (*auth_dto.TokenResponse, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	{
		if err != nil {
			return nil, err
		}
	}

	var res auth_dto.TokenResponse

	err = service.db.QueryRow(ctx, `
		INSERT INTO users (full_name, phone, password)
		VALUES ($1, $2, $3)
		RETURNING id, full_name, phone, role, created_at
	`, req.FullName, req.Phone, string(hashed)).
		Scan(&res.User.ID, &res.User.FullName, &res.User.Phone, &res.User.Role, &res.User.CreatedAt)

	if err != nil {
		return nil, err
	}

	token, err := helper.GenerateToken(res.User.ID, res.User.Role, 0)
	{
		if err != nil {
			return nil, err
		}
	}

	res.AccessToken = token

	return &res, nil
}

func (service *authService) Login(ctx context.Context, req auth_dto.Login) (*auth_dto.TokenResponse, error) {
	var (
		hashedPassword string
		res            auth_dto.TokenResponse
	)

	err := service.db.QueryRow(ctx, `
		SELECT id, full_name, phone, role, password, created_at
		FROM users
		WHERE phone = $1 AND is_active = true AND deleted_at IS NULL
	`, req.Phone).
		Scan(&res.User.ID, &res.User.FullName, &res.User.Phone, &res.User.Role, &hashedPassword, &res.User.CreatedAt)

	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	companyID, err := service.lookupCompanyID(ctx, res.User.ID)
	{
		if err != nil {
			return nil, err
		}
	}

	token, err := helper.GenerateToken(res.User.ID, res.User.Role, companyID)
	{
		if err != nil {
			return nil, err
		}
	}

	res.AccessToken = token

	return &res, nil
}
