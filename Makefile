GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX="sudo"
IMAGE_NAME="unixvoid/dproxy"
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

dproxy:
	$(GOC) dproxy.go

dependencies:
	go get github.com/miekg/dns
	go get github.com/unixvoid/glogger
	go get gopkg.in/gcfg.v1
	go get gopkg.in/redis.v3

run:
	go run \
		dproxy/dproxy.go \
		dproxy/checkdomain.go \
		dproxy/parseupstream.go \
		dproxy/resolve.go

docker:
	mkdir stage.tmp/
	$(MAKE) stat
	cp bin/dproxy* stage.tmp/dproxy
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

aci:
	$(MAKE) stat
	mkdir -p stage.tmp/dproxy-layout/rootfs/
	tar -zxf deps/rootfs.tar.gz -C stage.tmp/dproxy-layout/rootfs/
	cp bin/dproxy* stage.tmp/dproxy-layout/rootfs/dproxy
	chmod +x deps/run.sh
	cp deps/run.sh stage.tmp/dproxy-layout/rootfs/
	sed -i "s/\$$GIT_HASH/$(GIT_HASH)/g" stage.tmp/dproxy-layout/rootfs/run.sh
	cp config.gcfg stage.tmp/dproxy-layout/rootfs/
	cp deps/manifest.json stage.tmp/dproxy-layout/manifest
	cd stage.tmp/ && \
		actool build dproxy-layout dproxy.aci && \
		mv dproxy.aci ../
	@echo "dproxy.aci built"

testaci:
	deps/testrkt.sh

travisaci:
	wget https://github.com/appc/spec/releases/download/v0.8.7/appc-v0.8.7.tar.gz
	tar -zxf appc-v0.8.7.tar.gz
	$(MAKE) stat
	mkdir -p stage.tmp/dproxy-layout/rootfs/
	tar -zxf deps/rootfs.tar.gz -C stage.tmp/dproxy-layout/rootfs/
	cp bin/dproxy* stage.tmp/dproxy-layout/rootfs/dproxy
	chmod +x deps/run.sh
	cp deps/run.sh stage.tmp/dproxy-layout/rootfs/
	sed -i "s/\$$GIT_HASH/$(GIT_HASH)/g" stage.tmp/dproxy-layout/rootfs/run.sh
	cp config.gcfg stage.tmp/dproxy-layout/rootfs/
	cp deps/manifest.json stage.tmp/dproxy-layout/manifest
	cd stage.tmp/ && \
		../appc-v0.8.7/actool build dproxy-layout dproxy.aci && \
		mv dproxy.aci ../
	@echo "dproxy.aci built"

clean:
	rm -f dproxy.aci
	rm -rf bin/
	rm -rf stage.tmp/

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/dproxy-$(GIT_HASH)-linux-amd64 dproxy/*.go
