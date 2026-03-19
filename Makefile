include .envrc

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo  'Running application…'
	@go run ./cmd/api -port=${PORT} -env=${ENVIRONMENT} -db-dsn=${DB_DSN} -limiter-rps=${LIMITER_RPS} -limiter-burst=${LIMITER_BURST} -limiter-enabled=${LIMITER_ENABLED} -cors-trusted-origins=${CORS_TRUSTED_ORIGINS}

## db/psql: connect to the database using psql (terminal)
.PHONY: db/psql
db/psql:
	psql ${DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}


## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${DB_DSN} up
	@echo 'Running seeder...'
	psql ${DB_DSN} -f ./seeders/seeder_up.sql

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running down migrations...'
	migrate -path ./migrations -database ${DB_DSN} down

## db/migrations/force: apply all force database migrations
.PHONY: db/migrations/force
db/migrations/force:
	@echo 'Running force migrations...'
	migrate -path ./migrations -database ${DB_DSN} force 1