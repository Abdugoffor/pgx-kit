package user_cmd

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	user_handler "pgx-kit/module/user_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	user_handler.NewUserHandler(router, "/api/v1", db)
}
