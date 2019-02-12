all: deps linux_build

export GOBIN := $(CURDIR)/bin

CURRENT_GIT_GROUP = online_judge
CURRENT_GIT_REPO  = JudgeServer
COMMONENVVAR      = GOOS=linux GOARCH=amd64
BUILDENVVAR       = CGO_ENABLED=0
BIN_JudgeServer = JudgeServer

ifeq ($(DOCKER_IMAGE_TAG),)
        DOCKER_IMAGE_TAG  ?= $(shell git describe --abbrev=0 --tags)
endif

DOCKER_IMAGE_NAME ?= harbor.platform.facethink.com/axer/team

deps:
	dep ensure -v

build:
	$(BUILDENVVAR) go build -o $(GOBIN)/$(BIN_JudgeServer)  -ldflags "-X main.BuildTime=`date '+%Y-%m-%d_%I:%M:%S%p'` -X main.BuildGitHash=`git rev-parse HEAD` -X main.BuildGitTag=`git describe --tags` -w -s" $(CURRENT_GIT_GROUP)/$(CURRENT_GIT_REPO)

linux_build:
	$(COMMONENVVAR) $(BUILDENVVAR) make build

#编译Docker
docker: image clean

image:
	@echo ">> build docker team"
	docker build --rm --no-cache -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) $(CURDIR)

clean:
	for imageId in `docker images | grep none | awk '{ print $$3 }'`; \
	do \
		docker rmi -f $$imageId; \
	done

.PHONY: deps, build, docker