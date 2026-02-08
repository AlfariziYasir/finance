up:
	docker compose up --force-recreate
build:
	docker compose build --no-cache
swagger:
	swag init -g cmd/main.go --parseDependency --parseInternal
test:
	go test -v ./...