module intel/isecl/workload-service/v3

require (
	github.com/google/uuid v1.1.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/intel-secl/intel-secl/v3 v3.6.0
	github.com/jinzhu/gorm v1.9.16
	github.com/lib/pq v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	intel/isecl/lib/common/v3 v3.6.0
	intel/isecl/lib/flavor/v3 v3.6.0
	intel/isecl/lib/verifier/v3 v3.6.0
)

replace (
	github.com/intel-secl/intel-secl/v3 => github.com/intel-secl/intel-secl/v3 v3.6.0
	github.com/vmware/govmomi => github.com/arijit8972/govmomi fix-tpm-attestation-output
	intel/isecl/lib/common/v3 => github.com/intel-secl/common/v3 v3.6.0
	intel/isecl/lib/flavor/v3 => github.com/intel-secl/flavor/v3 v3.6.0
	intel/isecl/lib/verifier/v3 => github.com/intel-secl/verifier/v3 v3.6.0
)
