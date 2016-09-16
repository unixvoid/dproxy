GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX="sudo"
IMAGE_NAME="unixvoid/dproxy"
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

dproxy:
	$(GOC) dproxy.go

run:
	go run dproxy.go

docker:
	mkdir stage.tmp/
	$(MAKE) stat
	mv dproxy stage.tmp/
	cp deps/Dockerfile stage.tmp/
	cp deps/rootfs.tar.gz stage.tmp/
	cp deps/run.sh stage.tmp/
	sed -i "s/<GIT_HASH>/$(GIT_HASH)/g" stage.tmp/Dockerfile
	cd stage.tmp/ && \
		$(DOCKER_PREFIX) docker build -t $(IMAGE_NAME) .

dockerrun:
	$(DOCKER_PREFIX) docker run \
		-d \
		-p 8053:8053 \
		-v $(shell pwd)/config.gcfg:/config.gcfg \
		-v $(shell pwd)/upstream/:/upstream/ \
		--name dproxy \
		$(IMAGE_NAME)
	$(DOCKER_PREFIX) docker logs -f dproxy

clean:
	rm -f dproxy
	rm -rf stage.tmp/

stat:
	$(CGOR) $(GOC) $(GOFLAGS) dproxy.go
