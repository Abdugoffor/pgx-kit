package order_cmd

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	order_handler "pgx-kit/module/order_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	order_handler.NewOrderHandler(router, "/api/v1", db)
}
