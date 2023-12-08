.PHONY: clean test security build run swag

APP_NAME = c4stage
BUILD_DIR = $(PWD)/build

# values to pass for BinVersion, GitCommitLog, GitStatus, BuildTime and BuildGoVersion"
# Version=`git describe --tags`  # git tag 1.0.1  # require tag tagged before
Version=1.0.0
BuildTime=`date +%FT%T%z`
BuildGoVersion=`go version`

LDFLAGS=-ldflags "-w -s \
-X 'ppi.de/launchpad/shared.version=${Version}' \
-X 'ppi.de/launchpad/shared.buildTime=${BuildTime}' \
-X 'ppi.de/launchpad/shared.buildGoVersion=${BuildGoVersion}' \
"

clean:
	rm -rf ./build

linter:
	golangci-lint run

arch-go.install:
	go install github.com/fdaines/arch-go@v1.0.2

arch-go.check:
	arch-go --verbose

test:
	go test --short -v -timeout 60s -coverprofile=cover.out -cover ./...
	go tool cover -func=cover.out

build: clean
	CGO_ENABLED=0 go build ${LDFLAGS} -o $(BUILD_DIR)/$(APP_NAME) .

docker.neo4j:
	docker-compose -f docker/compose.yaml up
