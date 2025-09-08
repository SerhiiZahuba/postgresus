run:
	go run cmd/main.go

test:
	go test -count=1 ./internal/...

lint:
	golangci-lint fmt && golangci-lint run

migration-create:
	goose create $(name) sql

migration-up:
	goose up

migration-down:
	goose down

swagger:
	swag init -g ./cmd/main.go -o swagger
