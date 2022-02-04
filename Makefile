GITCOMMIT := $(shell git describe --always)
GITCOMMITDATE := $(shell git log -1 --date=short --pretty=format:%cd)
VERSION := "v4.0.2"
BUILDDATE := $(shell TZ=UTC date +%Y-%m-%dT%H:%M:%S%z)
PROXY_EXISTS := $(shell if [[ "${https_proxy}" || "${http_proxy}" ]]; then echo 1; else echo 0; fi)
MONOREPO_GITURL := "https://github.com/intel-innersource/applications.security.isecl.intel-secl"
MONOREPO_GITBRANCH := "v4.0.2/develop"

ifeq ($(PROXY_EXISTS),1)
	DOCKER_PROXY_FLAGS = --build-arg http_proxy=${http_proxy} --build-arg https_proxy=${https_proxy}
endif

.PHONY: workload-service installer wls-docker wls-oci-archive all clean

workload-service:
	env GOOS=linux GOSUMDB=off GOPROXY=direct go build -ldflags "-X intel/isecl/workload-service/v4/version.BuildDate=$(BUILDDATE) -X intel/isecl/workload-service/v4/version.Version=$(VERSION) -X intel/isecl/workload-service/v4/version.GitHash=$(GITCOMMIT)" -o out/workload-service

installer: workload-service
	mkdir -p out/wls
	cp dist/linux/workload-service.service out/wls/workload-service.service
	cp dist/linux/install.sh out/wls/install.sh && chmod +x out/wls/install.sh
	cp out/workload-service out/wls/workload-service
	git clone --depth 1 -b $(MONOREPO_GITBRANCH) $(MONOREPO_GITURL) tmp_monorepo
	cp -a tmp_monorepo/pkg/lib/common/upgrades/* out/wls/
	rm -rf tmp_monorepo
	cp -a upgrades/* out/wls/
	mv out/wls/build/* out/wls/
	chmod +x out/wls/*.sh

	makeself out/wls out/wls-$(VERSION).bin "Workload Service $(VERSION)" ./install.sh

oci-archive: workload-service
	cp dist/docker/entrypoint.sh out/entrypoint.sh && chmod +x out/entrypoint.sh
ifeq ($(PROXY_EXISTS),1)
	docker build -t isecl/workload-service:$(VERSION) --build-arg http_proxy=${http_proxy} --build-arg https_proxy=${https_proxy} -f dist/docker/Dockerfile .
else
	docker build -t isecl/workload-service:$(VERSION) -f dist/docker/Dockerfile .
endif
	skopeo copy docker-daemon:isecl/workload-service:$(VERSION) oci-archive:out/workload-service-$(VERSION)-$(GITCOMMIT).tar:$(VERSION)

swagger-get:
	wget https://github.com/go-swagger/go-swagger/releases/download/v0.21.0/swagger_linux_amd64 -O /usr/local/bin/swagger
	chmod +x /usr/local/bin/swagger
	wget https://repo1.maven.org/maven2/io/swagger/codegen/v3/swagger-codegen-cli/3.0.16/swagger-codegen-cli-3.0.16.jar -O /usr/local/bin/swagger-codegen-cli.jar

swagger-doc:
	mkdir -p out/swagger
	env GOOS=linux GOSUMDB=off GOPROXY=direct \
	/usr/local/bin/swagger generate spec -o ./out/swagger/openapi.yml --scan-models
	java -jar /usr/local/bin/swagger-codegen-cli.jar generate -i ./out/swagger/openapi.yml -o ./out/swagger/ -l html2 -t ./swagger/templates/

k8s: oci-archive
	cp -r dist/k8s out/k8s

swagger: swagger-get swagger-doc

all: clean installer

clean:
	rm -rf out/
