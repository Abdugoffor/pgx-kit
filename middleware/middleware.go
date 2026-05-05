package middleware

import (
	"context"
	"net/http"
	"strings"

	"pgx-kit/helper"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
)

type contextKey string

const (
	ContextUserID    contextKey = "user_id"
	ContextRole      contextKey = "role"
	ContextCompanyID contextKey = "company_id"
)

func CheckRole(next httprouter.Handle, roles ...string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		authHeader := r.Header.Get("Authorization")
		{
			if !strings.HasPrefix(authHeader, "Bearer ") {
				helper.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(helper.ENV("JWT_KEY")), nil
		})

		if err != nil || !token.Valid {
			helper.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		{
			if !ok {
				helper.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
		}

		userRole, _ := claims["role"].(string)
		userID, _ := claims["user_id"].(float64)
		companyID, _ := claims["company_id"].(float64)

		if len(roles) > 0 {
			allowed := false
			{
				for _, role := range roles {
					if role == userRole {
						allowed = true
						break
					}
				}
			}

			if !allowed {
				helper.JSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
		}

		ctx := context.WithValue(r.Context(), ContextUserID, int64(userID))
		ctx = context.WithValue(ctx, ContextRole, userRole)
		ctx = context.WithValue(ctx, ContextCompanyID, int64(companyID))

		next(w, r.WithContext(ctx), ps)
	}
}

func UserID(r *http.Request) int64 {
	id, _ := r.Context().Value(ContextUserID).(int64)
	return id
}

func UserRole(r *http.Request) string {
	role, _ := r.Context().Value(ContextRole).(string)
	return role
}

func CompanyID(r *http.Request) int64 {
	id, _ := r.Context().Value(ContextCompanyID).(int64)
	return id
}
