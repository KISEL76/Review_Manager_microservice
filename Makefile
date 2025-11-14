.PHONY: fmt lint

fmt:
	goimports -w ./microservice/

lint:
	cd microservice && golangci-lint run ./...