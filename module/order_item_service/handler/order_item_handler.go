package order_item_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	"pgx-kit/helper"
	"pgx-kit/middleware"
	order_item_dto "pgx-kit/module/order_item_service/dto"
	order_item_service "pgx-kit/module/order_item_service/service"
)

type orderItemHandler struct {
	service order_item_service.OrderItemService
}

func NewOrderItemHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &orderItemHandler{service: order_item_service.NewOrderItemService(db)}

	router.POST(group+"/orders/:order_id/items", middleware.CheckRole(handler.Create, "admin", "user"))
	router.GET(group+"/orders/:order_id/items", middleware.CheckRole(handler.List, "admin", "user"))
	router.PUT(group+"/order-items/:id", middleware.CheckRole(handler.Update, "admin"))
	router.DELETE(group+"/order-items/:id", middleware.CheckRole(handler.Delete, "admin"))
}

func (handler *orderItemHandler) Create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	orderID, err := strconv.ParseInt(ps.ByName("order_id"), 10, 64)
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order_id"})
			return
		}
	}

	var req order_item_dto.Create
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

	res, err := handler.service.Create(r.Context(), orderID, req)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusCreated, res)
}

func (handler *orderItemHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
	}

	var req order_item_dto.Update
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

	res, err := handler.service.Update(r.Context(), id, req)
	{
		if err != nil {

			if errors.Is(err, pgx.ErrNoRows) {
				helper.JSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
				return
			}

			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *orderItemHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
	}

	if err := handler.service.Delete(r.Context(), id); err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}

		helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *orderItemHandler) List(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	orderID, err := strconv.ParseInt(ps.ByName("order_id"), 10, 64)
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid order_id"})
			return
		}
	}

	items, err := handler.service.List(r.Context(), orderID)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusOK, map[string]any{"data": items})
}
