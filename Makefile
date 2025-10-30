swagger:
	swag init -g cmd/api/main.go -o docs
build:
	go build -o bin/api cmd/api/main.go
run: build
	bin/api
docker_run:
	docker-compose up -d
docker_stop:
	docker-cpmpose down

.PHONY:[swagger, build, run, docker_run, docker_stop]