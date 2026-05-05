package auth_handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	"pgx-kit/helper"
	auth_dto "pgx-kit/module/auth_service/dto"
	auth_service "pgx-kit/module/auth_service/service"
)

type authHandler struct {
	service auth_service.AuthService
}

func NewAuthHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &authHandler{service: auth_service.NewAuthService(db)}

	router.POST(group+"/auth/register", handler.Register)
	router.POST(group+"/auth/login", handler.Login)
}

func (handler *authHandler) Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req auth_dto.Register
	{
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	res, err := handler.service.Register(r.Context(), req)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusCreated, res)
}

func (handler *authHandler) Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req auth_dto.Login
	{
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	res, err := handler.service.Login(r.Context(), req)
	{
		if err != nil {

			if errors.Is(err, auth_service.ErrInvalidCredentials) {
				helper.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
				return
			}

			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}
