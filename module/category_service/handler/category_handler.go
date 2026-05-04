package category_handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	"pgx-kit/helper"
	category_dto "pgx-kit/module/category_service/dto"
	category_service "pgx-kit/module/category_service/service"
)

var sortFields = map[string]bool{
	"id": true, "name": true,
	"is_active": true, "created_at": true, "updated_at": true,
}

type categoryHandler struct {
	service category_service.CategoryService
}

func NewCategoryHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &categoryHandler{service: category_service.NewCategoryService(db)}

	routes := group + "/categories"
	{
		router.POST(routes, handler.Create)
		router.PUT(routes+"/:id", handler.Update)
		router.DELETE(routes+"/:id", handler.Delete)
		router.GET(routes+"/:id", handler.Show)
		router.GET(routes, handler.CursorList)
		router.GET(routes+"/admin", handler.AdminList)
	}
}

func (handler *categoryHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req category_dto.Create
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

	res, err := handler.service.Create(r.Context(), req)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusCreated, res)
}

func (handler *categoryHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
	}

	var req category_dto.Update
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

func (handler *categoryHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
	}

	if err := handler.service.Delete(r.Context(), id); err != nil {
		helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *categoryHandler) Show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
	}

	res, err := handler.service.Show(r.Context(), id)
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

func (handler *categoryHandler) AdminList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	filter := category_dto.AdminFilter{
		Filter: category_dto.Filter{
			Name: q.Get("name"),
		},
		Page:    1,
		PageSize: 20,
		SortBy:  "id",
		SortDir: "asc",
	}

	if v := q.Get("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filter.IsActive = &b
		}
	}

	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			filter.Page = n
		}
	}

	if v := q.Get("page_size"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			filter.PageSize = n
		}
	}

	if v := q.Get("sort_by"); sortFields[v] {
		filter.SortBy = v
	}

	if v := q.Get("sort_dir"); v == "asc" || v == "desc" {
		filter.SortDir = v
	}

	items, hasNext, err := handler.service.AdminList(r.Context(), filter)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusOK, category_dto.AdminListResponse{
		Data:     items,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		HasNext:  hasNext,
		HasPrev:  filter.Page > 1,
	})
}

func (handler *categoryHandler) CursorList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	filter := category_dto.CursorFilter{
		Filter: category_dto.Filter{
			Name: q.Get("name"),
		},
		Limit:   20,
		SortBy:  "id",
		SortDir: "asc",
	}

	if v := q.Get("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			filter.IsActive = &b
		}
	}

	if v := q.Get("cursor"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.Cursor = &n
		}
	}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			filter.Limit = n
		}
	}

	if v := q.Get("sort_by"); sortFields[v] {
		filter.SortBy = v
	}

	if v := q.Get("sort_dir"); v == "asc" || v == "desc" {
		filter.SortDir = v
	}

	items, hasMore, err := handler.service.CursorList(r.Context(), filter)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	resp := category_dto.CursorListResponse{
		Data:    items,
		HasMore: hasMore,
	}

	if hasMore {
		lastID := items[len(items)-1].ID
		resp.NextCursor = &lastID
	}

	helper.JSON(w, http.StatusOK, resp)
}
