include .env

# =============================================== #
# HELPERS
# =============================================== #

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

#prerequities target: it will be run before the actual command
.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

# =============================================== #
# DEVELOPMENT
# =============================================== #

.PHONY: run/api
run/api:
	@echo 'Starting the server'
	@go run ./cmd/api -db-dsn=${GREENLIGHT_DATABASE_DSN}

.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DATABASE_DSN}

#@ is used to prevent command being echoed in output
#add confirm target as prerequities target
.PHONY: db/migration/up
db/migration/up: confirm
	@echo 'Running up migrations..'
	@migrate -path ./migrations --database ${GREENLIGHT_DATABASE_DSN} up

#passing arguments
.PHONY: db/migration/new
db/migration/new:
	@echo 'Creating Migration files for ${name}..'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

# =============================================== #
# QUALITY CONTROL
# =============================================== #

.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests'
	go test -race -vet=off ./...