.PHONY: build run test vet fmt theme-doc langs-json docker

build:
	go build -o grs-server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -l cmd internal scripts

theme-doc:
	go run ./scripts/go/generate-theme-doc

langs-json:
	go run ./scripts/go/generate-langs-json

docker:
	docker build -t github-readme-stats .
