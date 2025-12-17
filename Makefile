ifeq ($(OS),Windows_NT)
    detected_OS := Windows
	ext := .exe
else
    detected_OS := $(shell uname)
	ext :=
endif

ifeq ($(detected_OS),Windows)
    include make.win.env
	export $(shell sed 's/=.*//' make.win.env)
	docker_opts := --network host  --platform="linux/amd64" --privileged=true
endif
ifeq ($(detected_OS),Darwin)
	include make.osx.env
	export $(shell sed 's/=.*//' make.osx.env)
	docker_opts := --platform="linux/amd64" --privileged=true
endif
ifeq ($(detected_OS),Linux)
    include make.linux.env
	export $(shell sed 's/=.*//' make.linux.env)
	docker_opts := --privileged=true
endif

define check_timeout
    timer=0; \
    $(1); do \
    timer=$$(expr $$timer + 1); \
    if [ "$$timer" = 600 ]; then \
        exit -1; \
    fi; \
    sleep 1; \
    done
endef

all: test build example

xk6${ext}:
	xk6${ext} -V

test: setup-db2 *.go testdata/*.js
	go test -count 1 ./...

container-test: setup-container test

build: k6

setup-container:
	-docker rm -f db2test
	-docker run  --rm -d --name db2test ${docker_opts} -p 50000:50000 -e LICENSE=accept -e DB2INST1_PASSWORD=password123 -e DBNAME=SAMPLE icr.io/db2_community/db2:latest
	$(call check_timeout, until docker exec db2test su - -c "db2 connect to SAMPLE" db2inst1)

setup-db2:
	curl -L https://raw.githubusercontent.com/ibmdb/go_ibm_db/refs/heads/master/installer/setup.go -o setup.go
	-go run setup.go
	rm -f setup.go

k6: xk6${ext} setup-db2 *.go go.mod go.sum
	xk6${ext} build -v --with github.com/grafana/xk6-sql@latest --with github.com/oleiade/xk6-encoding@latest --with github.com/iambaim/xk6-sql-driver-db2=.

example: k6 setup-container
	./k6${ext} run examples/example.js

clean:
	rm -f k6${ext}

.PHONY: test all example setup-db2 setup-container clean
