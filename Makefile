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
	docker_opts := --platform="linux/amd64" --privileged=true
	k6_env = env PATH="$$(cygpath -u "$(IBM_DB_HOME)/bin"):$$PATH"
endif
ifeq ($(detected_OS),Darwin)
	include make.osx.env
	export $(shell sed 's/=.*//' make.osx.env)
	docker_opts := --platform="linux/amd64" --privileged=true
	# Pass the library path explicitly because macOS strips it when launching /bin/sh.
	k6_env = env DYLD_LIBRARY_PATH="$(DYLD_LIBRARY_PATH)"
	test_flags = -exec 'env "DYLD_LIBRARY_PATH=$(DYLD_LIBRARY_PATH)"'
endif
ifeq ($(detected_OS),Linux)
    include make.linux.env
	export $(shell sed 's/=.*//' make.linux.env)
	docker_opts := --privileged=true
endif

K6_VERSION ?= v2.4.0
XK6_SQL_VERSION ?= v1.2.4
XK6_ENCODING_VERSION ?= v0.0.0-20260702145940-258b2345beb3

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
	xk6${ext} -V || (echo "Please install xk6 or make sure it is available in your PATH"; exit 1)

test: setup-db2 *.go testdata/*.js
	go test $(test_flags) -count 1 ./...

container-test: setup-container test

build: k6${ext}

setup-container:
	-docker rm -f db2test
	-docker run  --rm -d --name db2test ${docker_opts} -p 50000:50000 -e LICENSE=accept -e DB2INST1_PASSWORD=password123 -e DBNAME=SAMPLE icr.io/db2_community/db2:latest
	$(call check_timeout, until docker exec db2test su - -c "db2 connect to SAMPLE" db2inst1)

setup-db2:
	go run ./helpers/db2_cli_setup.go

k6${ext}: xk6${ext} setup-db2 *.go go.mod go.sum
	xk6${ext} build -v --output k6${ext} --k6-version $(K6_VERSION) \
		--with github.com/grafana/xk6-sql@$(XK6_SQL_VERSION) \
		--with github.com/oleiade/xk6-encoding@$(XK6_ENCODING_VERSION) \
		--with github.com/iambaim/xk6-sql-driver-db2=.

ifeq ($(detected_OS),Windows)
.PHONY: k6
k6: k6.exe
endif

smoke-test: k6${ext}
	$(MAKE) check-module-loading

check-module-loading:
	K6_AUTO_EXTENSION_RESOLUTION=false $(k6_env) ./k6${ext} inspect examples/example.js

example: k6${ext} setup-container
	$(k6_env) ./k6${ext} run examples/example.js

clean:
	rm -f k6${ext}

.PHONY: test all build container-test smoke-test check-module-loading example setup-db2 setup-container clean
