package auth_cmd

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	auth_handler "pgx-kit/module/auth_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	auth_handler.NewAuthHandler(router, "/api/v1", db)
}
