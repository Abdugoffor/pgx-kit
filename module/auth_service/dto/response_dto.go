package auth_dto

import "time"

type UserInfo struct {
	ID        int64     `json:"id"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	CompanyID int64     `json:"company_id"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenResponse struct {
	AccessToken string   `json:"access_token"`
	User        UserInfo `json:"user"`
}
