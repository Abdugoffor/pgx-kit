package product_cmd

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	product_handler "pgx-kit/module/product_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	product_handler.NewProductHandler(router, "/api/v1", db)
}
