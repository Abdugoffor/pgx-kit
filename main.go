package main

import (
	"fmt"
	"net/http"
	"os"
	"pgx-kit/config"
	"pgx-kit/helper"
	category_cmd "pgx-kit/module/category_service"
	language_cmd "pgx-kit/module/language_service"
	product_cmd  "pgx-kit/module/product_service"

	"github.com/julienschmidt/httprouter"
)

func main() {

	helper.LoadEnv()

	migrate()

	db := config.DBConnect()

	router := httprouter.New()
	{
		language_cmd.Cmd(router, db)
		category_cmd.Cmd(router, db)
		product_cmd.Cmd(router, db)
	}

	fmt.Println("🚀 Server started on http://localhost:8080")

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}

func migrate() {
	if len(os.Args) > 1 {

		switch os.Args[1] {

		case "migrate:create":

			if len(os.Args) < 3 {
				fmt.Println("Usage:   go run . migrate:create <name>")

				fmt.Println("Example: go run . migrate:create add_code_to_regions")

				os.Exit(1)
			}

			config.MigrateCreate(os.Args[2])

			return
		}
	}
}
