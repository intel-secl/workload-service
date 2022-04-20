module intel/isecl/workload-service/v4

require (
	github.com/google/uuid v1.2.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/intel-secl/intel-secl/v4 v4.2.0-Beta
	github.com/jinzhu/gorm v1.9.16
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	intel/isecl/lib/common/v4 v4.2.0-Beta
	intel/isecl/lib/flavor/v4 v4.2.0-Beta
	intel/isecl/lib/verifier/v4 v4.2.0-Beta
)

replace (
	intel/isecl/lib/common/v4 => github.com/intel-secl/common/v4 v4.2.0-Beta
	intel/isecl/lib/flavor/v4 => github.com/intel-secl/flavor/v4 v4.2.0-Beta
	intel/isecl/lib/verifier/v4 => github.com/intel-secl/verifier/v4 v4.2.0-Beta
)
