module intel/isecl/workload-service

require (
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.3
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/lib/pq v1.2.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.4.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/yaml.v2 v2.2.2
	intel/isecl/lib/clients v1.0.0
	intel/isecl/lib/common v1.0.0-Beta
	intel/isecl/lib/flavor v1.0.0
	intel/isecl/lib/kms-client v1.0.0
	intel/isecl/lib/verifier v1.0.0
)

replace intel/isecl/lib/flavor => gitlab.devtools.intel.com/sst/isecl/lib/flavor.git v2.1/develop

replace intel/isecl/lib/common => gitlab.devtools.intel.com/sst/isecl/lib/common.git v2.1/develop

replace intel/isecl/lib/verifier => gitlab.devtools.intel.com/sst/isecl/lib/verifier.git v2.1/develop

replace intel/isecl/lib/kms-client => gitlab.devtools.intel.com/sst/isecl/lib/kms-client.git v2.1/develop

replace intel/isecl/lib/clients => gitlab.devtools.intel.com/sst/isecl/lib/clients.git v2.1/develop

replace intel/isecl/lib/mtwilson-client => gitlab.devtools.intel.com/sst/isecl/lib/mtwilson-client.git v2.1/develop
