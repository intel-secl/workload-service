module intel/isecl/workload-service/v4

require (
	github.com/google/uuid v1.2.0
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.7.3
	github.com/intel-secl/intel-secl/v4 v4.1.1
	github.com/jinzhu/gorm v1.9.16
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	intel/isecl/lib/common/v4 v4.1.1
	intel/isecl/lib/flavor/v4 v4.1.1
	intel/isecl/lib/verifier/v4 v4.1.1
)

replace (
	github.com/intel-secl/intel-secl/v4 => github.com/intel-innersource/applications.security.isecl.intel-secl/v4 v4.1.1/develop
	intel/isecl/lib/common/v4 => github.com/intel-innersource/libraries.security.isecl.common/v4 v4.1.1/develop
	intel/isecl/lib/flavor/v4 => github.com/intel-innersource/libraries.security.isecl.flavor/v4 v4.1.1/develop
	intel/isecl/lib/verifier/v4 => github.com/intel-innersource/libraries.security.isecl.verifier/v4 v4.1.1/develop
)
