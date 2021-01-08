module intel/isecl/workload-service/v3

require (
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/intel-secl/intel-secl/v3 v3.3.1
	github.com/jinzhu/gorm v1.9.12
	github.com/lib/pq v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/sirupsen/logrus v1.4.0
	github.com/stretchr/testify v1.3.0
	gopkg.in/yaml.v2 v2.3.0
	intel/isecl/lib/common/v3 v3.3.1
	intel/isecl/lib/flavor/v3 v3.3.1
	intel/isecl/lib/verifier/v3 v3.3.1
)

replace intel/isecl/lib/flavor/v3 => github.com/intel-secl/flavor/v3 v3.3.1

replace intel/isecl/lib/common/v3 => github.com/intel-secl/common/v3 v3.3.1

replace intel/isecl/lib/verifier/v3 => github.com/intel-secl/verifier/v3 v3.3.1

replace github.com/vmware/govmomi => github.com/arijit8972/govmomi fix-tpm-attestation-output

replace github.com/intel-secl/intel-secl/v3 => github.com/intel-secl/intel-secl/v3 v3.3.1
