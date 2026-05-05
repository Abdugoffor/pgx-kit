package language_handler

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
	language_dto "pgx-kit/module/language_service/dto"
	language_service "pgx-kit/module/language_service/service"
)

var sortCols = map[string]string{
	"id": "id", "name": "name", "description": "description",
	"is_active": "is_active", "created_at": "created_at", "updated_at": "updated_at",
}

func parseSortBy(v string) string {
	if col, ok := sortCols[v]; ok {
		return col
	}
	return "id"
}

func parseSortDir(v string) string {
	if v == "desc" {
		return "desc"
	}
	return "asc"
}

type languageHandler struct {
	service language_service.LanguageService
}

func NewLanguageHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &languageHandler{service: language_service.NewLanguageService(db)}

	routes := group + "/languages"
	{
		router.POST(routes, middleware.CheckRole(handler.Create, "admin", "user"))
		router.PUT(routes+"/:id", middleware.CheckRole(handler.Update, "admin"))
		router.DELETE(routes+"/:id", middleware.CheckRole(handler.Delete, "admin"))
		router.GET(routes+"/:id", middleware.CheckRole(handler.Show, "admin", "user"))
		router.GET(routes, middleware.CheckRole(handler.CursorList, "admin", "user"))
		router.GET(group+"/admin/languages", middleware.CheckRole(handler.AdminList, "admin", "user"))
	}
}

func (handler *languageHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req language_dto.Create
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

func (handler *languageHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	{
		if err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}
	}

	var req language_dto.Update
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

func (handler *languageHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func (handler *languageHandler) Show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

func (handler *languageHandler) AdminList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	filter := language_dto.AdminFilter{
		Filter: language_dto.Filter{
			Name:        q.Get("name"),
			Description: q.Get("description"),
		},
		Page:     1,
		PageSize: 20,
		SortBy:   "id",
		SortDir:  "asc",
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

	filter.SortBy = parseSortBy(q.Get("sort_by"))
	filter.SortDir = parseSortDir(q.Get("sort_dir"))

	items, hasNext, err := handler.service.AdminList(r.Context(), filter)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	helper.JSON(w, http.StatusOK, language_dto.AdminListResponse{
		Data:     items,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		HasNext:  hasNext,
		HasPrev:  filter.Page > 1,
	})
}

func (handler *languageHandler) CursorList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	filter := language_dto.CursorFilter{
		Filter: language_dto.Filter{
			Name:        q.Get("name"),
			Description: q.Get("description"),
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
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			filter.Cursor = &n
		}
	}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			filter.Limit = n
		}
	}

	filter.SortBy = parseSortBy(q.Get("sort_by"))
	filter.SortDir = parseSortDir(q.Get("sort_dir"))

	items, hasMore, err := handler.service.CursorList(r.Context(), filter)
	{
		if err != nil {
			helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
	}

	resp := language_dto.CursorListResponse{
		Data:    items,
		HasMore: hasMore,
	}

	if hasMore {
		lastID := items[len(items)-1].ID
		resp.NextCursor = &lastID
	}

	helper.JSON(w, http.StatusOK, resp)
}
