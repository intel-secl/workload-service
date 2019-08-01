module intel/isecl/workload-service

require (
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.3
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/lib/pq v1.1.1 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	gopkg.in/yaml.v2 v2.2.2
	intel/isecl/lib/clients v0.0.0
	intel/isecl/lib/common v1.0.0-Beta
	intel/isecl/lib/flavor v0.0.0
	intel/isecl/lib/kms-client v0.0.0
	intel/isecl/lib/verifier v0.0.0
)

replace intel/isecl/lib/flavor => gitlab.devtools.intel.com/sst/isecl/lib/flavor.git v0.0.0-20190724130420-55de0ba53dc1

replace intel/isecl/lib/common => gitlab.devtools.intel.com/sst/isecl/lib/common.git v0.0.0-20190729104031-73fba1fea2b9

replace intel/isecl/lib/verifier => gitlab.devtools.intel.com/sst/isecl/lib/verifier.git v0.0.0-20190724131213-64d659e7973d

replace intel/isecl/lib/kms-client => gitlab.devtools.intel.com/sst/isecl/lib/kms-client.git v0.0.0-20190515175351-9007ad0d5416

replace intel/isecl/cms => gitlab.devtools.intel.com/sst/isecl/certificate-management-service.git v0.0.0-20190718144616-b20fd70163eb

replace intel/isecl/lib/clients => gitlab.devtools.intel.com/sst/isecl/lib/clients.git v0.0.0-20190801010949-eded2c2c8405
