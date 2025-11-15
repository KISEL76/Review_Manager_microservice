all: up

up:
	docker-compose up --build -d 

down:
	docker-compose down

down-v:
	docker-compose down -v 

logs:
	docker-compose logs -f app

test:
	cd microservice && go test ./tests/...
	
test-unit:
	cd microservice && go test -v ./tests/unittests/... 

fmt:
	goimports -w ./microservice/

lint:
	cd microservice && golangci-lint run ./...

.PHONY: fmt lint