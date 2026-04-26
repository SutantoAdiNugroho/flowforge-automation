.PHONY: test test_api init run shutdown

init:
	cd backend && go mod tidy && go mod download
	cd frontend && npm install

run:
	docker compose down --volumes
	docker compose build
	docker compose up -d

shutdown:
	docker compose down --volumes

logs:
	docker compose logs -f

test:
	cd backend && go test ./pkg/... -v

test_api:
	cd backend && go test ./tests/api_test.go -v