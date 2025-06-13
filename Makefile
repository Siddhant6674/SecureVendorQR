build:
	@go build -o bin/VENDORQR cmd/main.go

test:build
	@go test -v ./...

run:build
	@./bin/VENDORQR
