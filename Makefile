include .envrc

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo  'Running application...'
	@go run ./cmd/api -port=${PORT} -env=${ENVIRONMENT} -db-dsn=${DB_DSN} -limiter-rps=${LIMITER_RPS} -limiter-burst=${LIMITER_BURST} -limiter-enabled=${LIMITER_ENABLED} -cors-trusted-origins=${CORS_TRUSTED_ORIGINS}

# test the rate limiter
.PHONY: test/ratelimiter
test/ratelimiter:
	@echo 'Testing rate limiter...'
	@for i in $$(seq 1 10); do curl -i localhost:4000/classes/1; done

.PHONY: test/responsecompression
test/responsecompression:
	@echo 'Testing response size without gzip encoding...'
	@curl -i localhost:4000/classes
	@echo 'Testing response size with gzip encoding...'
	@curl -i --compressed localhost:4000/classes

.PHONY: test/classregistration
test/classregistration:
	@echo 'Testing member registering with lower membership tier...'
	@curl -d '{"class_id": 2, "member_id": 1, "status": "active"}' localhost:4000/registrations/add
	@echo 'Testing member registering into full class...'
	@curl -d '{"class_id": 5, "member_id": 2, "status": "active"}' localhost:4000/registrations/add
	@echo 'Testing member registering into a terminated class...'
	@curl -d '{"class_id": 12, "member_id": 1, "status": "active"}' localhost:4000/registrations/add
	@echo 'Testing expired member registering...'
	@curl -d '{"class_id": 1, "member_id": 5, "status": "active"}' localhost:4000/registrations/add
	@echo 'Testing member registering with schedule conflicts...'
	@curl -d '{"class_id": 5, "member_id": 1, "status": "active"}' localhost:4000/registrations/add
	@echo 'Testing normal class registration...'
	@curl -d '{"class_id": 1, "member_id": 2, "status": "active"}' localhost:4000/registrations/add

.PHONY: test/sessiontimes
test/sessiontimes:
	@echo 'Testing session overlap of same class...'
	@curl -d '{"class_id": 1, "day": "wed", "time": "8:00", "duration": 61}' localhost:4000/sessiontimes/add
	@echo 'Testing trainer overlap...'
	@curl -d '{"class_id": 6, "day": "wed", "time": "9:30", "duration": 40}' localhost:4000/sessiontimes/add
	@echo 'Testing studio booking overlap...'
	@curl -d '{"class_id": 7, "day": "mon", "time": "7:30", "duration": 40}' localhost:4000/sessiontimes/add
	@echo 'Testing session posting on terminated class...'
	@curl -d '{"class_id": 12, "day": "mon", "time": "7:30", "duration": 40}' localhost:4000/sessiontimes/add
	@echo 'Testing normal session time posting...'
	@curl -d '{"class_id": 6, "day": "fri", "time": "10:30", "duration": 80}' localhost:4000/sessiontimes/add

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