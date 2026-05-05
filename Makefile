migrate:
	go run . migrate:create $(name)

gen:
	go run ./tools/gen -name $(name) -table $(table) -fields "$(fields)"

.PHONY: migrate gen

run:
	go run main.go