module intel/isecl/workload-service/v2

require (
	github.com/gorilla/handlers v1.4.0
	github.com/gorilla/mux v1.7.3
	github.com/jinzhu/gorm v1.9.15
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/lib/pq v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.4.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/yaml.v2 v2.2.2
	intel/isecl/lib/clients/v2 v2.2.0
	intel/isecl/lib/common/v2 v2.2.0
	intel/isecl/lib/flavor/v2 v2.2.0
	intel/isecl/lib/kms-client/v2 v2.2.0
	intel/isecl/lib/verifier/v2 v2.2.0
)

replace intel/isecl/lib/flavor/v2 => github.com/intel-secl/flavor/v2 v2.2.0

replace intel/isecl/lib/common/v2 => github.com/intel-secl/common/v2 v2.2.0

replace intel/isecl/lib/verifier/v2 => github.com/intel-secl/verifier/v2 v2.2.0

replace intel/isecl/lib/kms-client/v2 => github.com/intel-secl/kms-client/v2 v2.2.0

replace intel/isecl/lib/clients/v2 => github.com/intel-secl/clients/v2 v2.2.0
