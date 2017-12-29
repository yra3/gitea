#Makefile related to docker

DOCKER_IMAGE ?= gitea/gitea
DOCKER_TAG ?= latest
DOCKER_REF := $(DOCKER_IMAGE):$(DOCKER_TAG)

DOCKER_TEST_CONTAINER ?= gitea-test-env

.PHONY: docker
docker:
	docker run -ti --rm -v $(CURDIR):/srv/app/src/code.gitea.io/gitea -w /srv/app/src/code.gitea.io/gitea -e TAGS="bindata $(TAGS)" webhippie/golang:edge make clean generate build
	docker build --disable-content-trust=false -t $(DOCKER_REF) .


.PHONY: docker-env-setup
docker-env-setup:
	@if [ -z $$(docker container ls --filter name=gitea-test-env -q) ]; then \
  	echo "Starting docker container in background for tests" ; \
  	docker run -d --rm --name $(DOCKER_TEST_CONTAINER) -v $(CURDIR):/go/src/code.gitea.io/gitea -w /go/src/code.gitea.io/gitea golang sleep infinity ; \
	fi

.PHONY: docker-env-clean
docker-env-clean:
	docker stop $(DOCKER_TEST_CONTAINER)

DOCKER_CMD ?= make test
.PHONY: docker-test
docker-test: docker-env-setup
	docker exec -ti $(DOCKER_TEST_CONTAINER) $(DOCKER_CMD)
