package language_cmd

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	language_handler "pgx-kit/module/language_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	language_handler.NewLanguageHandler(router, "/api/v1", db)
}
