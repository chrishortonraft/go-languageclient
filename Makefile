PROJECT_HOME=$(dir $(firstword) $(MAKEFILE_LIST))
IMAGE=go-languageclient
REGISTRY=ghcr.io/raft-tech
VERSION=dev
FULL_IMAGE=${IMAGE}:${VERSION}
# What docker command you use (docker, podman, etc.) b/c Makefiles don't respect shell aliases (like `alias docker=podman`)
DOCKER=docker

build: fmt vet
	go build -o ${IMAGE} main.go

docker: fmt vet
	$(DOCKER) buildx build -f ${PROJECT_HOME}/Dockerfile \
       ${PROJECT_HOME}/ \
	   --platform linux/amd64 \
       -t ${FULL_IMAGE} --load

docker-arm: fmt vet
	$(DOCKER) buildx build -f ${PROJECT_HOME}/Dockerfile \
       ${PROJECT_HOME}/ \
	   --platform linux/arm64 \
       -t ${FULL_IMAGE} --load

run-arm: docker-arm
	$(DOCKER) run -p 8080:8080 --rm -t ${FULL_IMAGE} --

run: docker
	$(DOCKER) run -t ${FULL_IMAGE}

fmt:
	go fmt ./...
	
vet:
	go vet ./...

test: fmt vet
	go clean -cache
	go test ./... -v

pull:
	$(DOCKER) pull ${FULL_IMAGE}

push:
	$(DOCKER) push ${FULL_IMAGE}