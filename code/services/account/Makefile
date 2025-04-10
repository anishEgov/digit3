DB_URL ?= postgres://postgres:password@db:5432/digit?sslmode=disable
MIGRATE=docker run --rm -v $(shell pwd)/migrations:/migrations --network host migrate/migrate -path=/migrations -database "$(DB_URL)"

.PHONY: migrate-up migrate-down create-migration

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

create-migration:
	docker run --rm -v $(shell pwd)/migrations:/migrations migrate/migrate create -ext sql -dir /migrations -seq $(name)