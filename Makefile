.PHONY: build up down clean test-router test-node

build:
	@mkdir -p bin
	@go build -o bin/router ./apps/router/cmd/router
	@go build -o bin/node ./apps/node/cmd/node

up:
	@docker compose up -d

down:
	@docker compose down

clean:
	@rm -rf bin
	@go clean -i ./...

test-router: build
	@./bin/router

test-node: build
	@./bin/node
