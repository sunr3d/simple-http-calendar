up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f app

test:
	go test -v ./...

fmt:
	go fmt ./...

build:
	go build -o simple-http-calendar cmd/main.go