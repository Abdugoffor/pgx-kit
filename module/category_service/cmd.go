package category_cmd

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	category_handler "pgx-kit/module/category_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	category_handler.NewCategoryHandler(router, "/api/v1", db)
}
