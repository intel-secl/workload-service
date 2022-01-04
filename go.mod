module intel/isecl/workload-service/v4

require (
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/intel-secl/intel-secl/v4 v4.2.0
	github.com/jinzhu/gorm v1.9.16
	github.com/lib/pq v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	intel/isecl/lib/common/v4 v4.2.0
	intel/isecl/lib/flavor/v4 v4.2.0
	intel/isecl/lib/verifier/v4 v4.2.0
)

replace (
	intel/isecl/lib/common/v4 => gitlab.devtools.intel.com/sst/isecl/lib/common.git/v4 v4.2/develop
	intel/isecl/lib/flavor/v4 => gitlab.devtools.intel.com/sst/isecl/lib/flavor.git/v4 v4.2/develop
	intel/isecl/lib/verifier/v4 => gitlab.devtools.intel.com/sst/isecl/lib/verifier.git/v4 v4.2/develop
	github.com/intel-secl/intel-secl/v4 => gitlab.devtools.intel.com/sst/isecl/intel-secl.git/v4 v4.2/develop
)
