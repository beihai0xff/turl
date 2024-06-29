COMMIT_ID				:= $(shell git rev-parse --short HEAD)
TAG_VERSION				:= $(shell git describe --abbrev=0 --tags)
BUILD_TIME				:= $(shell date)
ARCH					:= $(shell uname -m)

# fill the ldflags with the build info
ldflags					=  "-w -X 'github.com/beihai0xff/turl/cli.version=$(TAG_VERSION)' -X 'github.com/beihai0xff/turl/cli.gitHash=$(COMMIT_ID)' -X 'github.com/beihai0xff/turl/cli.buildTime=$(BUILD_TIME)'"
BUILD_PLATFORMS 		=  linux/amd64,linux/arm64
GO_VERSION 				=  1.22-bookworm

# different Linux(MacOS) distro use different arch name, so we unify them using the same name aarch64
# eg. on MacOS with Apple silicon arch name is arm64, we use aarch64 as the arch name
ifneq ($(ARCH),x86_64)
    ARCH = aarch64
endif

bootstrap:
	go mod download -x
	go generate -tags tools tools/tools.go

lint:
	golangci-lint run -v

gen/mock:
	mockery

gen/struct_tag:
	bash -x scripts/gen_configs_struct_tag.sh

gen/swagger:
	swag init --parseDependency --parseDepth 1 -g app/turl/server/http_controller.go -o docs/swagger

test: bootstrap gen/mock
	docker compose -f ./internal/tests/docker-compose.yaml up -d --wait
	go test -gcflags="all=-l" -race -coverprofile=coverage.out -v ./...
	docker compose -f ./internal/tests/docker-compose.yaml down


.PHONY: bootstrap lint gen/mock gen/struct_tag gen/swagger test
#
# build section
#

build: build/docker

# build binary file
build/binary: clean bootstrap
	go build -tags=jsoniter -ldflags=$(ldflags) -o ./build/dist/binary/turl cmd/turl/main.go

# docker: enable containerd for pulling and storing images
build/docker:
	DOCKER_BUILDKIT=1 docker buildx build \
		--ulimit nofile=1048576:1048576 \
		-f ./build/Dockerfile \
 		--build-arg BUILD_DATE="$(BUILD_TIME)" \
 		--build-arg BUILD_COMMIT="$(COMMIT_ID)" \
 		--build-arg BUILD_VERSION="$(TAG_VERSION)" \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--platform=$(BUILD_PLATFORMS) \
		--output type=docker \
		-t beihai0xff/turl:latest -t beihai0xff/turl:$(TAG_VERSION) .

build/docker_and_push:
	DOCKER_BUILDKIT=1 docker buildx build \
		--ulimit nofile=1048576:1048576 \
		-f ./build/Dockerfile \
		--build-arg BUILD_DATE="$(BUILD_TIME)" \
		--build-arg BUILD_COMMIT="$(COMMIT_ID)" \
		--build-arg BUILD_VERSION="$(TAG_VERSION)" \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--platform=$(BUILD_PLATFORMS) \
		--push \
		-t beihai0xff/turl:latest -t beihai0xff/turl:$(TAG_VERSION) .

.PHONY: build build/binary build/docker build/docker_and_push


.PHONY: clean
clean:
	rm -rf ./build/dist ./coverage.out internal/tests/mocks

#
# upload section
#


upload/docker:
	docker push

.PHONY: upload/docker



.PHONY: deploy
deploy:
	@echo "starting turl service containers..."
	docker compose -f ./internal/example/docker-compose.yaml \
		-p turl-service up -V --abort-on-container-exit
	@echo "turl service containers start successfully"

