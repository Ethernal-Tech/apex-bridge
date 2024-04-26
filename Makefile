.PHONY: lint
lint: check-lint
	golangci-lint run --config .golangci.yml

.PHONY: build
build: check-go check-git
#	$(eval COMMIT_HASH = $(shell git rev-parse HEAD))
#		-X 'github.com/Ethernal-Tech/apex-bridge/versioning.Commit=$(COMMIT_HASH)'\
#   $(eval VERSION = $(shell git describe --tags --abbrev=0 ${COMMIT_HASH}))
#       -X 'github.com/Ethernal-Tech/apex-bridge/versioning.Version=$(VERSION)'
	$(eval BRANCH = $(shell git rev-parse --abbrev-ref HEAD | tr -d '\040\011\012\015\n'))
	$(eval TIME = $(shell date))
	go build -o apex-bridge -ldflags="\
		-X 'github.com/Ethernal-Tech/apex-bridge/versioning.Branch=$(BRANCH)'\
		-X 'github.com/Ethernal-Tech/apex-bridge/versioning.BuildTime=$(TIME)'" \
	main.go

.PHONY: unit-test
unit-test: check-go
	go test -race -shuffle=on -coverprofile coverage.out -timeout 20m `go list ./... | grep -v e2e`	

.PHONY: check-lint
check-lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint is not installed. Please install and try again."; exit 1)

.PHONY: check-go
check-go:
	@which go > /dev/null || (echo "Go is not installed.. Please install and try again."; exit 1)

.PHONY: check-git
check-git:
	@which git > /dev/null || (echo "git is not installed. Please install and try again."; exit 1)