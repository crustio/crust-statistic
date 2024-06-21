VERSION=-ldflags="-X main.Version=$(shell git describe --tags)"


build:
	@echo "  >  \033[32mBuilding binary...\033[0m "
	go build -o build/statistic $(VERSION)

build-docker: ## Builds a docker image with the binary
	docker build -t crustio/statistic -f ./Dockerfile .

docker-run:
	docker run -d --name statistic -v ./config.ini:/app/config.ini    crustio/statistic
