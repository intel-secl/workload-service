GITTAG := $(shell git describe --tags --abbrev=0 2> /dev/null)
GITCOMMIT := $(shell git describe --always)
GITCOMMITDATE := $(shell git log -1 --date=short --pretty=format:%cd)
VERSION := $(or ${GITTAG}, v0.0.0)
BUILDDATE := $(shell TZ=UTC date +%Y-%m-%dT%H:%M:%S%z)

.PHONY: workload-service installer docker all clean

workload-service:
	env GOOS=linux GOSUMDB=off GOPROXY=direct go build -ldflags "-X intel/isecl/workload-service/version.BuildDate=$(BUILDDATE) -X intel/isecl/workload-service/version.Version=$(VERSION) -X intel/isecl/workload-service/version.GitHash=$(GITCOMMIT)" -o out/workload-service

installer: workload-service
	mkdir -p out/wls
	cp dist/linux/workload-service.service out/wls/workload-service.service
	cp dist/linux/install.sh out/wls/install.sh && chmod +x out/wls/install.sh
	cp out/workload-service out/wls/workload-service
	makeself out/wls out/wls-$(VERSION).bin "Workload Service $(VERSION)" ./install.sh

docker: installer
	cp dist/docker/entrypoint.sh out/entrypoint.sh && chmod +x out/entrypoint.sh
	docker build --build-arg http_proxy=http://proxy-us.intel.com:911 --build-arg https_proxy=http://proxy-us.intel.com:911 -t isecl/workload-service:latest -f ./dist/docker/Dockerfile ./out
	docker save isecl/workload-service:latest > ./out/docker-wls-$(VERSION).tar 

all: clean installer

clean:
	rm -rf out/
