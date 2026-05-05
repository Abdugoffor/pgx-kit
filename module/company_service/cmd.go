package company_cmd

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"

	company_handler "pgx-kit/module/company_service/handler"
)

func Cmd(router *httprouter.Router, db *pgxpool.Pool) {
	company_handler.NewCompanyHandler(router, "/api/v1", db)
}
