include .envrc

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@echo  'Running application...'
	@go run ./cmd/api -port=${PORT} -env=${ENVIRONMENT} -db-dsn=${DB_DSN} -limiter-rps=${LIMITER_RPS} -limiter-burst=${LIMITER_BURST} -limiter-enabled=${LIMITER_ENABLED} -cors-trusted-origins=${CORS_TRUSTED_ORIGINS} -smtp-host=${MAILTRAP_HOST} -smtp-port=${MAILTRAP_PORT} -smtp-username=${MAILTRAP_USERNAME} -smtp-password=${MAILTRAP_PASSWORD} -smtp-sender=${MAILTRAP_SENDER}

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
	@echo 'Running up seeders...'
	psql ${DB_DSN} -f ./seeders/users_seeder_up.sql
	psql ${DB_DSN} -f ./seeders/classes_seeder_up.sql

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

## test/tokens: generate tokens for different roles
.PHONY: test/tokens
test/tokens:
	@echo 'Generating session tokens for different roles...'
	curl -d '{"email": "${ADMIN_EMAIL}", "password": "${ACCOUNT_PASSWORD}"}' localhost:4000/tokens/authentication
	curl -d '{"email": "${TRAINER_EMAIL}", "password": "${ACCOUNT_PASSWORD}"}' localhost:4000/tokens/authentication
	curl -d '{"email": "${MEMBER_EMAIL}", "password": "${ACCOUNT_PASSWORD}"}' localhost:4000/tokens/authentication
	curl -d '{"email": "${UNACTIVATED_EMAIL}", "password": "${ACCOUNT_PASSWORD}"}' localhost:4000/tokens/authentication

## test/authentication: test different authentication scenarios
.PHONY: test/authentication
test/authentication:
	@echo 'Testing guest account for open route [should succeed]...'
	curl -i localhost:4000/gyms/2
	@echo 'Testing guest account for closed route [should fail]...'
	curl -i localhost:4000/sessions
	@echo 'Testing activated account for closed route [should succeed]...'
	curl -i -H "Authorization: Bearer ${ADMIN_TOKEN}" localhost:4000/members/3
	@echo 'Testing unactivated account for closed route [should fail]...'
	curl -i -H "Authorization: Bearer ${UNACTIVATED_TOKEN}" localhost:4000/members

## test/permissions: test different permissions scenarios
.PHONY: test/permissions
test/permissions:
	@echo 'Testing member adding a studio [should fail]...'
	curl -i -X POST localhost:4000/studios/add \
	-H "Authorization: Bearer ${MEMBER_TOKEN}" \
  	-H "Content-Type: application/json" \
  	-d '{"gym_id": 1, "name": "Main Studio", "access": "classes"}'
	@echo 'Testing trainer adding a studio [should succeed]...'
	curl -i -X POST localhost:4000/studios/add \
	-H "Authorization: Bearer ${TRAINER_TOKEN}" \
  	-H "Content-Type: application/json" \
  	-d '{"gym_id": 1, "name": "Main Studio", "access": "classes"}'
	@echo 'Testing trainer editing a gym [should fail]...'
	curl -X PATCH \
	-H "Authorization: Bearer ${TRAINER_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"location": "Spanish Lookout"}' localhost:4000/gyms/2/update
	@echo 'Testing admin editing a gym [should succeed]...'
	curl -X PATCH \
	-H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{"location": "Spanish Lookout"}' localhost:4000/gyms/2/update

## test/authorization: test different authorization scenarios
.PHONY: test/authorization
test/authorization:
	@echo 'Testing member viewing sessions [should only view sessions they have attended (1, 2)]...'
	curl -i -H "Authorization: Bearer ${MEMBER_TOKEN}" localhost:4000/sessions
	@echo 'Testing trainer viewing sessions [should only view sessions linked to their classes (1, 5, 9, 13)]...'
	curl -i -H "Authorization: Bearer ${TRAINER_TOKEN}" localhost:4000/sessions
	@echo 'Testing admin viewing sessions [should view all sessions]...'
	curl -i -H "Authorization: Bearer ${ADMIN_TOKEN}" localhost:4000/sessions
	@echo "Testing trainer deleting another trainer's class [should fail]..."
	curl -i -X DELETE -H "Authorization: Bearer ${TRAINER_TOKEN}" localhost:4000/classes/3/delete
	@echo "Testing trainer deleting their own class [should succeed]..."
	curl -i -X DELETE -H "Authorization: Bearer ${TRAINER_TOKEN}" localhost:4000/classes/1/delete

## test/newuser: test adding a new user
.PHONY: test/newuser
test/newuser:
	@echo 'Testing adding a new user...'
	curl -i localhost:4000/users/add \
	-H "Authorization: Bearer ${ADMIN_TOKEN}" \
	-d '{"role_id": 3, "username":"Kathy Rivers", "email":"kr@example.com", "password": "password"}'

## test/activateuser: test activating a new user
.PHONY: test/activateuser
test/activateuser:
	@echo 'Testing activating a new user...'
	curl -X PUT -d '{"token": "${ACTIVATE_TOKEN}"}' localhost:4000/users/activated


