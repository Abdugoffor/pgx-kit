migrate:
	go run . migrate:create $(name)

.PHONY: migrate

run:
	go run main.go