.PHONY: run certs migrate-up migrate-down postgres-up postgres-down postgres-delete containers-down migrate-version redis-up redis-down redis-delete


#first you have to export all the environment variables
#run 
#set -a && source .env && set +a
#set -a && source db/postgres/.env && set +a
#set -a && source db/redis/.env && set +a
#then everything works

certs:
	./scripts/makecerts.sh


postgres-up:
	docker compose -f db/postgres/postgres.yaml up -d
postgres-down:
	docker compose -f db/postgres/postgres.yaml down 
postgres-delete:
	docker compose -f db/postgres/postgres.yaml down -v

migrate-up:
	migrate -path db/postgres/migrations/ -database $(PG_URL) up
migrate-down:
	migrate -path db/postgres/migrations/ -database $(PG_URL) down
migrate-version:
	migrate -path db/postgres/migrations/ -database $(PG_URL) version



redis-up:
	docker compose -f db/redis/redis.yaml up -d
redis-down:
	docker compose -f db/redis/redis.yaml down 
redis-delete:
	docker compose -f db/redis/redis.yaml down -v


run:
	go run cmd/server/main.go

run-all: 
	./scripts/makecerts.sh 
	docker compose -f db/postgres/postgres.yaml up -d 
	docker compose -f db/redis/redis.yaml up -d
	@echo "Waiting for Postgres..."
	@until pg_isready -h $(PGHOST) -p $(PGPORT); do \
		sleep 1; \
	done
	migrate -path db/postgres/migrations/ -database $(PG_URL) up
	@echo "Waiting for Redis..."
	@until redis-cli ping | grep -q PONG; do \
		sleep 1; \
	done
	@echo "Redis ready!"
	go run cmd/server/main.go

containers-delete:
	make redis-delete
	make postgres-delete

containers-down:
	make redis-down
	make postgres-down

help:
	@grep -E '^[a-zA-Z0-9_-]+:' Makefile | awk -F: '{print $$1}'
t:
	@bash -c 'target=$$(make -qp | awk -F: "/^[a-zA-Z0-9][^$#\/\t=]*:([^=]|$$)/ {print $$1}" | sed "s/:$$//" | fzf) && make $$target'
