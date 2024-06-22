GIT_DESC				:= $(shell git describe --always --tags --long --match 'v[0-9]*' HEAD| sed -E 's/-/ /g')
COMMIT_ID				?= $(lastword $(GIT_DESC))
COMMIT_NUM				:= $(firstword $(word 3, $(GIT_DESC)) 0)
RC_ID					:= $(firstword $(word 2, $(GIT_DESC)) rc0)
MAIN_VERSION			:= $(subst v,,$(word 1, $(GIT_DESC)))
GIT_TAG					:= "${RC_ID}.${COMMIT_NUM}.git.${COMMIT_ID}"

BUILD_TIME				:= $(shell git show -s --format=%cd)
# fill the ldflags with the build info
ldflags					=  "-w -X "
BUILD_PLATFORMS 		=  linux/amd64,linux/arm64
GO_VERSION 				=  1.22-bookworm
ARCH					=  $(shell uname -m)
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
	bash -x scripts/gen_mock.sh

gen/struct_tag:
	bash -x scripts/gen_configs_struct_tag.sh

gen/swagger:
	swag init --parseDependency --parseDepth 1 -g app/turl/server/http_controller.go -o docs/swagger

test: bootstrap gen/mock
	docker compose -f ./tests/docker-compose.yaml up -d --wait
	go test -gcflags="all=-l" -race -coverprofile=coverage.out -v ./...


.PHONY: bootstrap lint gen/mock gen/struct_tag gen/swagger test
#
# build section
#

build: clean
	docker run --rm \
		-w /workspace \
 		-v=$(shell pwd):/workspace \
 		golang:$(GO_VERSION) bash -c "make build/binary && make build/rpms"

# build binary file
build/binary: clean bootstrap
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags=$(ldflags) -o ./build/dist/binary/x86_64/turl cmd/turl/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags=$(ldflags) -o ./build/dist/binary/aarch64/turl cmd/turl/main.go


# docker: enable containerd for pulling and storing images
build/docker:
	DOCKER_BUILDKIT=1 docker buildx build \
		--ulimit nofile=1048576:1048576 \
		-f ./build/Dockerfile \
 		--build-arg BUILD_DATE="$(BUILD_TIME)" \
 		--build-arg BUILD_COMMIT="$(COMMIT_ID)" \
 		--build-arg BUILD_VERSION="$(MAIN_VERSION)" \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--platform=$(BUILD_PLATFORMS) \
		--output type=docker \
		-t  .

build/docker_and_push:
	DOCKER_BUILDKIT=1 docker buildx build \
		--ulimit nofile=1048576:1048576 \
		-f ./build/Dockerfile \
 		--build-arg BUILD_DATE="$(BUILD_TIME)" \
 		--build-arg BUILD_COMMIT="$(COMMIT_ID)" \
 		--build-arg BUILD_VERSION="$(MAIN_VERSION)" \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--platform=$(BUILD_PLATFORMS) \
		--push \
		-t registry. .

.PHONY: build build/binary build/docker build/docker_and_push


.PHONY: clean
clean:
	rm -rf ./build/dist ./coverage.out

#
# upload section
#


upload/docker:
	docker push

.PHONY: upload/docker

