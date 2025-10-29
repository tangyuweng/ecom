swagger:
	swag init -g cmd/api/main.go -o docs
build:
	go build -o bin/api cmd/api/main.go
run: build
	bin/api

.PHONY:[swagger, build, run]