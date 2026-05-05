package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// make gen name=brand table=brands fields="name:string:required max=255,slug:string:required max=255"

// Field represents a single model field parsed from -fields flag.
// Format: "fieldName:goType:validateTag"
// Example: "name:string:required max=255,price:float64:required gt=0,note:*string:omitempty"
type Field struct {
	Name        string // Go struct field name, e.g. "Name"
	GoType      string // e.g. "string", "*string", "float64"
	ValidateTag string // e.g. "required,max=255"
	JSONName    string // snake_case json key
	SQLName     string // snake_case column name
}

func main() {
	name := flag.String("name", "", "module name in snake_case, e.g. product")
	table := flag.String("table", "", "DB table name, e.g. products (default: name+'s')")
	fields := flag.String("fields", "", `comma-separated fields: "name:string:required,price:float64:required gt=0,note:*string:omitempty"`)
	out := flag.String("out", "module", "output base dir (relative to project root)")
	module := flag.String("module", "pgx-kit", "Go module path from go.mod")
	flag.Parse()

	if *name == "" || *fields == "" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./tools/gen -name <snake_name> -fields \"field:type:validate,...\"")
		fmt.Fprintln(os.Stderr, "Example: go run ./tools/gen -name product -table products -fields \"name:string:required,price:float64:required gt=0,description:*string:omitempty\"")
		os.Exit(1)
	}

	if *table == "" {
		*table = *name + "s"
	}

	parsed := parseFields(*fields)
	if len(parsed) == 0 {
		fmt.Fprintln(os.Stderr, "No fields parsed. Check -fields format.")
		os.Exit(1)
	}

	g := &Gen{
		Name:       *name,
		Table:      *table,
		Fields:     parsed,
		OutBase:    *out,
		ModulePath: *module,
		Pascal:     toPascal(*name),
		Camel:      toCamel(*name),
	}

	g.generate()
}

// Gen holds all naming conventions derived from the base name.
type Gen struct {
	Name       string
	Table      string
	Fields     []Field
	OutBase    string
	ModulePath string
	Pascal     string // ProductItem
	Camel      string // productItem
}

func (g *Gen) generate() {
	moduleDir := filepath.Join(g.OutBase, g.Name+"_service")
	g.write(filepath.Join(moduleDir, "cmd.go"), g.genCmd())
	g.write(filepath.Join(moduleDir, "handler", g.Name+"_handler.go"), g.genHandler())
	g.write(filepath.Join(moduleDir, "service", g.Name+"_service.go"), g.genService())
	g.write(filepath.Join(moduleDir, "dto", "request_dto.go"), g.genRequestDTO())
	g.write(filepath.Join(moduleDir, "dto", "response_dto.go"), g.genResponseDTO())

	fmt.Println("\n✅ Generated files:")
	fmt.Printf("   %s/cmd.go\n", moduleDir)
	fmt.Printf("   %s/handler/%s_handler.go\n", moduleDir, g.Name)
	fmt.Printf("   %s/service/%s_service.go\n", moduleDir, g.Name)
	fmt.Printf("   %s/dto/request_dto.go\n", moduleDir)
	fmt.Printf("   %s/dto/response_dto.go\n", moduleDir)
	fmt.Printf("\n📌 Add to main.go:\n")
	fmt.Printf("   import %s_cmd \"%s/module/%s_service\"\n", g.Name, g.ModulePath, g.Name)
	fmt.Printf("   %s_cmd.Cmd(router, db)\n", g.Name)
}

func (g *Gen) write(path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", filepath.Dir(path), err)
		os.Exit(1)
	}
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(os.Stderr, "⚠️  skip (exists): %s\n", path)
		return
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
		os.Exit(1)
	}
	fmt.Printf("   created: %s\n", path)
}

// ---------- cmd.go ----------

func (g *Gen) genCmd() string {
	pkg := g.Name + "_cmd"
	handlerPkg := g.Name + "_handler"
	return fmt.Sprintf(`package %s

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	%s "%s/module/%s_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	%s.New%sHandler(router, "/api/v1", db)
}
`, pkg, handlerPkg, g.ModulePath, g.Name, handlerPkg, g.Pascal)
}

// ---------- handler ----------

func (g *Gen) genHandler() string {
	handlerPkg := g.Name + "_handler"
	dtoPkg := g.Name + "_dto"
	svcPkg := g.Name + "_service"
	plural := g.Table

	// sort columns from fields + always include id, created_at, updated_at
	sortEntries := []string{`"id": "id"`, `"created_at": "created_at"`, `"updated_at": "updated_at"`}
	for _, f := range g.Fields {
		if f.GoType == "string" || f.GoType == "*string" {
			sortEntries = append(sortEntries, fmt.Sprintf(`"%s": "%s"`, f.SQLName, f.SQLName))
		}
	}

	return fmt.Sprintf(`package %s

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
	%s "%s/module/%s_service/dto"
	%s "%s/module/%s_service/service"
)

var sortCols = map[string]string{
	%s,
}

func parseSortBy(v string) string {
	if col, ok := sortCols[v]; ok {
		return col
	}
	return "id"
}

func parseSortDir(v string) string {
	if v == "asc" {
		return "asc"
	}
	return "desc"
}

type %sHandler struct {
	service %s.%sService
}

func New%sHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &%sHandler{service: %s.New%sService(db)}

	routes := group + "/%s"
	{
		router.POST(routes, middleware.CheckRole(handler.Create, "admin", "user"))
		router.PUT(routes+"/:id", middleware.CheckRole(handler.Update, "admin"))
		router.DELETE(routes+"/:id", middleware.CheckRole(handler.Delete, "admin"))
		router.GET(routes+"/:id", middleware.CheckRole(handler.Show, "admin", "user"))
		router.GET(routes, middleware.CheckRole(handler.CursorList, "admin", "user"))
		router.GET(group+"/admin/%s", middleware.CheckRole(handler.AdminList, "admin"))
	}
}

func (handler *%sHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req %s.Create
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	res, err := handler.service.Create(r.Context(), req)
	if err != nil {
		helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helper.JSON(w, http.StatusCreated, res)
}

func (handler *%sHandler) Update(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	var req %s.Update
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errs})
		return
	}

	res, err := handler.service.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *%sHandler) Delete(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	if err := handler.service.Delete(r.Context(), id); err != nil {
		helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *%sHandler) Show(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	res, err := handler.service.Show(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
			return
		}
		helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *%sHandler) AdminList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	filter := %s.AdminFilter{
		Page:     1,
		PageSize: 20,
		SortBy:   "id",
		SortDir:  "desc",
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
	if err != nil {
		helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	helper.JSON(w, http.StatusOK, %s.AdminListResponse{
		Data:     items,
		Page:     filter.Page,
		PageSize: filter.PageSize,
		HasNext:  hasNext,
		HasPrev:  filter.Page > 1,
	})
}

func (handler *%sHandler) CursorList(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	q := r.URL.Query()

	filter := %s.CursorFilter{
		Limit:   20,
		SortBy:  "id",
		SortDir: "desc",
	}

	if v := q.Get("cursor"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			filter.Cursor = &n
		}
	}

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			filter.Limit = n
		}
	}

	filter.SortBy = parseSortBy(q.Get("sort_by"))
	filter.SortDir = parseSortDir(q.Get("sort_dir"))

	items, hasMore, err := handler.service.CursorList(r.Context(), filter)
	if err != nil {
		helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	resp := %s.CursorListResponse{
		Data:    items,
		HasMore: hasMore,
	}

	if hasMore {
		lastID := items[len(items)-1].ID
		resp.NextCursor = &lastID
	}

	helper.JSON(w, http.StatusOK, resp)
}
`,
		handlerPkg,
		dtoPkg, g.ModulePath, g.Name,
		svcPkg, g.ModulePath, g.Name,
		strings.Join(sortEntries, ", "),
		g.Camel, svcPkg, g.Pascal,
		g.Pascal,
		g.Camel, svcPkg, g.Pascal,
		plural,
		plural,
		// Create
		g.Camel, dtoPkg,
		g.Camel,
		// Update
		g.Camel, dtoPkg,
		g.Camel,
		// Delete
		g.Camel,
		// Show
		g.Camel,
		// AdminList
		g.Camel, dtoPkg,
		g.Camel,
		dtoPkg,
		// CursorList
		g.Camel, dtoPkg,
		g.Camel,
		dtoPkg,
	)
}

// ---------- service ----------

func (g *Gen) genService() string {
	svcPkg := g.Name + "_service"
	dtoPkg := g.Name + "_dto"

	// Build INSERT columns and placeholders
	var insertCols, insertParams, insertScans []string
	insertCols = append(insertCols, "-- fill in required columns")
	for i, f := range g.Fields {
		insertCols = append(insertCols, f.SQLName)
		insertParams = append(insertParams, fmt.Sprintf("$%d", i+1))
		insertScans = append(insertScans, fmt.Sprintf("req.%s", f.Name))
	}

	// RETURNING clause
	returnCols := g.returningCols()
	scanArgs := g.scanArgs("res")

	// UPDATE SET
	var setClauses []string
	for i, f := range g.Fields {
		setClauses = append(setClauses, fmt.Sprintf("\t\t    %s = COALESCE($%d, %s),", f.SQLName, i+1, f.SQLName))
	}

	var updateArgs []string
	for _, f := range g.Fields {
		updateArgs = append(updateArgs, fmt.Sprintf("req.%s", f.Name))
	}
	updateArgs = append(updateArgs, "id")

	return fmt.Sprintf(`package %s

import (
	"context"
	"fmt"

	%s "%s/module/%s_service/dto"

	"github.com/jackc/pgx/v5/pgxpool"
)

type %sService interface {
	Create(ctx context.Context, req %s.Create) (*%s.Response, error)
	Update(ctx context.Context, id int64, req %s.Update) (*%s.Response, error)
	Delete(ctx context.Context, id int64) error
	Show(ctx context.Context, id int64) (*%s.Response, error)
	AdminList(ctx context.Context, filter %s.AdminFilter) ([]%s.Response, bool, error)
	CursorList(ctx context.Context, filter %s.CursorFilter) ([]%s.Response, bool, error)
}

type %sService struct {
	db *pgxpool.Pool
}

func New%sService(db *pgxpool.Pool) %sService {
	return &%sService{db: db}
}

func (s *%sService) Create(ctx context.Context, req %s.Create) (*%s.Response, error) {
	var res %s.Response

	err := s.db.QueryRow(ctx, `+"`"+`
		INSERT INTO %s (%s)
		VALUES (%s)
		RETURNING %s
	`+"`"+`, %s).
		Scan(%s)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *%sService) Update(ctx context.Context, id int64, req %s.Update) (*%s.Response, error) {
	var res %s.Response

	err := s.db.QueryRow(ctx, `+"`"+`
		UPDATE %s
		SET
%s
		    updated_at = now()
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING %s
	`+"`"+`, %s).
		Scan(%s)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *%sService) Delete(ctx context.Context, id int64) error {
	_, err := s.db.Exec(ctx, `+"`"+`
		UPDATE %s SET deleted_at = now() WHERE id = $1 AND deleted_at IS NULL
	`+"`"+`, id)
	return err
}

func (s *%sService) Show(ctx context.Context, id int64) (*%s.Response, error) {
	var res %s.Response

	err := s.db.QueryRow(ctx, `+"`"+`
		SELECT %s
		FROM %s
		WHERE id = $1 AND deleted_at IS NULL
	`+"`"+`, id).Scan(%s)

	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *%sService) AdminList(ctx context.Context, filter %s.AdminFilter) ([]%s.Response, bool, error) {
	rows, err := s.db.Query(ctx, fmt.Sprintf(`+"`"+`
		SELECT %s
		FROM %s
		WHERE deleted_at IS NULL
		ORDER BY %%s %%s
		LIMIT $1 OFFSET $2
	`+"`"+`, filter.SortBy, filter.SortDir),
		filter.PageSize+1, (filter.Page-1)*filter.PageSize)

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var items []%s.Response
	for rows.Next() {
		var res %s.Response
		if err := rows.Scan(%s); err != nil {
			return nil, false, err
		}
		items = append(items, res)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasNext := len(items) > filter.PageSize
	if hasNext {
		items = items[:filter.PageSize]
	}

	return items, hasNext, nil
}

func (s *%sService) CursorList(ctx context.Context, filter %s.CursorFilter) ([]%s.Response, bool, error) {
	rows, err := s.db.Query(ctx, fmt.Sprintf(`+"`"+`
		SELECT %s
		FROM %s
		WHERE deleted_at IS NULL
			AND ($1::bigint IS NULL OR id > $1)
		ORDER BY %%s %%s, id %%s
		LIMIT $2
	`+"`"+`, filter.SortBy, filter.SortDir, filter.SortDir),
		filter.Cursor, filter.Limit+1)

	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var items []%s.Response
	for rows.Next() {
		var res %s.Response
		if err := rows.Scan(%s); err != nil {
			return nil, false, err
		}
		items = append(items, res)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(items) > filter.Limit
	if hasMore {
		items = items[:filter.Limit]
	}

	return items, hasMore, nil
}
`,
		svcPkg,
		dtoPkg, g.ModulePath, g.Name,
		// interface
		g.Pascal,
		dtoPkg, dtoPkg,
		dtoPkg, dtoPkg,
		dtoPkg,
		dtoPkg, dtoPkg,
		dtoPkg, dtoPkg,
		// struct
		g.camelLower(),
		g.Pascal, g.Pascal,
		g.camelLower(),
		// Create
		g.camelLower(), dtoPkg, dtoPkg,
		dtoPkg,
		g.Table, strings.Join(g.fieldCols(), ", "), strings.Join(g.fieldPlaceholders(), ", "),
		returnCols,
		strings.Join(g.fieldReqArgs("req"), ", "),
		scanArgs,
		// Update
		g.camelLower(), dtoPkg, dtoPkg,
		dtoPkg,
		g.Table,
		strings.Join(setClauses, "\n"),
		len(g.Fields)+1,
		returnCols,
		strings.Join(append(g.fieldReqArgs("req"), "id"), ", "),
		scanArgs,
		// Delete
		g.camelLower(), g.Table,
		// Show
		g.camelLower(), dtoPkg, dtoPkg,
		returnCols, g.Table, scanArgs,
		// AdminList
		g.camelLower(), dtoPkg, dtoPkg,
		returnCols, g.Table,
		dtoPkg, dtoPkg, scanArgs,
		// CursorList
		g.camelLower(), dtoPkg, dtoPkg,
		returnCols, g.Table,
		dtoPkg, dtoPkg, scanArgs,
	)
}

// ---------- dto ----------

func (g *Gen) genRequestDTO() string {
	dtoPkg := g.Name + "_dto"

	// Create struct fields
	var createFields, updateFields []string
	for _, f := range g.Fields {
		validate := f.ValidateTag
		goType := f.GoType
		// For Update, always make pointers optional
		updateType := goType
		updateValidate := validate
		if !strings.HasPrefix(goType, "*") {
			updateType = "*" + goType
			// replace "required" with "omitempty" for update
			updateValidate = strings.ReplaceAll(validate, "required", "omitempty")
		}
		createFields = append(createFields,
			fmt.Sprintf("\t%s %s `json:\"%s\" validate:\"%s\"`", f.Name, goType, f.JSONName, validate))
		updateFields = append(updateFields,
			fmt.Sprintf("\t%s %s `json:\"%s\" validate:\"%s\"`", f.Name, updateType, f.JSONName, updateValidate))
	}

	return fmt.Sprintf(`package %s

type Create struct {
%s
}

type Update struct {
%s
}

type AdminFilter struct {
	Page     int
	PageSize int
	SortBy   string
	SortDir  string
}

type CursorFilter struct {
	Cursor  *int64
	Limit   int
	SortBy  string
	SortDir string
}
`, dtoPkg,
		strings.Join(createFields, "\n"),
		strings.Join(updateFields, "\n"),
	)
}

func (g *Gen) genResponseDTO() string {
	dtoPkg := g.Name + "_dto"

	var responseFields []string
	responseFields = append(responseFields, "\tID        int64     `json:\"id\"`")
	for _, f := range g.Fields {
		responseFields = append(responseFields,
			fmt.Sprintf("\t%-10s %-10s `json:\"%s\"`", f.Name, f.GoType, f.JSONName))
	}
	responseFields = append(responseFields,
		"\tCreatedAt time.Time `json:\"created_at\"`",
		"\tUpdatedAt time.Time `json:\"updated_at\"`",
	)

	return fmt.Sprintf(`package %s

import "time"

type Response struct {
%s
}

type AdminListResponse struct {
	Data     []Response `+"`"+`json:"data"`+"`"+`
	Page     int        `+"`"+`json:"page"`+"`"+`
	PageSize int        `+"`"+`json:"page_size"`+"`"+`
	HasNext  bool       `+"`"+`json:"has_next"`+"`"+`
	HasPrev  bool       `+"`"+`json:"has_prev"`+"`"+`
}

type CursorListResponse struct {
	Data       []Response `+"`"+`json:"data"`+"`"+`
	NextCursor *int64     `+"`"+`json:"next_cursor"`+"`"+`
	HasMore    bool       `+"`"+`json:"has_more"`+"`"+`
}
`, dtoPkg, strings.Join(responseFields, "\n"))
}

// ---------- helpers ----------

func (g *Gen) returningCols() string {
	cols := []string{"id"}
	for _, f := range g.Fields {
		col := f.SQLName
		if f.GoType == "float64" || f.GoType == "*float64" {
			col += "::float8"
		}
		cols = append(cols, col)
	}
	cols = append(cols, "created_at", "updated_at")
	return strings.Join(cols, ", ")
}

func (g *Gen) scanArgs(varName string) string {
	args := []string{"&" + varName + ".ID"}
	for _, f := range g.Fields {
		args = append(args, fmt.Sprintf("&%s.%s", varName, f.Name))
	}
	args = append(args, "&"+varName+".CreatedAt", "&"+varName+".UpdatedAt")
	return strings.Join(args, ", ")
}

func (g *Gen) fieldCols() []string {
	var cols []string
	for _, f := range g.Fields {
		cols = append(cols, f.SQLName)
	}
	return cols
}

func (g *Gen) fieldPlaceholders() []string {
	var ph []string
	for i := range g.Fields {
		ph = append(ph, fmt.Sprintf("$%d", i+1))
	}
	return ph
}

func (g *Gen) fieldReqArgs(varName string) []string {
	var args []string
	for _, f := range g.Fields {
		args = append(args, fmt.Sprintf("%s.%s", varName, f.Name))
	}
	return args
}

func (g *Gen) camelLower() string {
	return strings.ToLower(string(g.Pascal[0])) + g.Pascal[1:]
}

// parseFields parses "name:string:required max=255,price:float64:required" into []Field.
func parseFields(raw string) []Field {
	var fields []Field
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		segments := strings.SplitN(part, ":", 3)
		if len(segments) < 2 {
			continue
		}
		name := strings.TrimSpace(segments[0])
		goType := strings.TrimSpace(segments[1])
		validate := ""
		if len(segments) == 3 {
			validate = strings.TrimSpace(segments[2])
		}
		// convert spaces within validate tag to commas (user can write "required max=255" or "required,max=255")
		validate = strings.ReplaceAll(validate, " ", ",")

		f := Field{
			Name:        toPascal(name),
			GoType:      goType,
			ValidateTag: validate,
			JSONName:    toSnake(name),
			SQLName:     toSnake(name),
		}
		fields = append(fields, f)
	}
	return fields
}

// toPascal converts snake_case or camelCase to PascalCase.
func toPascal(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == '_' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// toCamel converts snake_case to camelCase.
func toCamel(s string) string {
	p := toPascal(s)
	if p == "" {
		return p
	}
	return strings.ToLower(string(p[0])) + p[1:]
}

// toSnake converts camelCase/PascalCase to snake_case.
func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteRune('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
