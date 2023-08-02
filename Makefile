GOMOD = go mod
GOBUILD = go build -trimpath -race -v
GOTEST = go test -v -cover -race

ROOT = $(shell git rev-parse --show-toplevel)
BIN = dist/install-via-git

.PHONY: $(BIN)
$(BIN):
	$(GOBUILD) -o $@

.PHONY: test
test:
	$(GOTEST) ./...

.PHONY: init
init:
	$(GOMOD) tidy

.PHONY: generate
generate: clean-generated
	go generate ./...

.PHONY: clean-generated
clean-generated:
	find . -name "*_generated.go" -type f -delete

.PHONY: vuln
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

DOCKER_RUN = docker run --rm -v "$(ROOT)":/usr/src/myapp -w /usr/src/myapp
DOCKER_GO_IMAGE = golang:1.20
DOCKER_LINT_IMAGE = golangci/golangci-lint:v1.53.3

.PHONY: docker-test
docker-test:
	$(DOCKER_RUN) $(DOCKER_GO_IMAGE) $(GOTEST) ./...

.PHONY: docker-dist
docker-dist:
	$(DOCKER_RUN) $(DOCKER_GO_IMAGE) $(GOBUILD) -o $(BIN) $(CMD)

.PHONY: docker-lint
docker-lint:
	$(DOCKER_RUN) $(DOCKER_LINT_IMAGE) golangci-lint run -v
