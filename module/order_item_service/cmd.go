package order_item_cmd

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	order_item_handler "pgx-kit/module/order_item_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	order_item_handler.NewOrderItemHandler(router, "/api/v1", db)
}
