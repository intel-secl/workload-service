GITTAG := $(shell git describe --tags --abbrev=0 2> /dev/null)
GITCOMMIT := $(shell git describe --always)
GITCOMMITDATE := $(shell git log -1 --date=short --pretty=format:%cd)
VERSION := $(or ${GITTAG}, v0.0.0)
workload-service:
	env GOOS=linux go build -ldflags "-X version.Version=$(VERSION)-$(GITCOMMIT)" -o out/workload-service

installer: workload-service
	mkdir -p out/wls
	cp dist/linux/setup.sh out/wls/setup.sh && chmod +x out/wls/setup.sh
	cp out/workload-service out/wls/workload-service
	makeself --sha256 out/wls out/wls-$(VERSION).bin "Workload Service $(VERSION)" ./setup.sh 

docker: installer
	docker build -f ./dist/docker/Dockerfile ./out

all: docker

clean:
	rm -rf out/