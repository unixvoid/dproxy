GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX="sudo"
IMAGE_NAME="unixvoid/dproxy"

dproxy:
	$(GOC) dproxy.go

run:
	go run dproxy.go

docker:
	$(MAKE) stat
	$(DOCKER_PREFIX) docker build -t $(IMAGE_NAME) .

clean:
	rm -f dproxy

stat:
	$(CGOR) $(GOC) $(GOFLAGS) dproxy.go
