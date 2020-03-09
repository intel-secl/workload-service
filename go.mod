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

replace intel/isecl/lib/flavor => github.com/intel-secl/flavor v2.0.0

replace intel/isecl/lib/common => github.com/intel-secl/common v2.0.0

replace intel/isecl/lib/verifier => github.com/intel-secl/verifier v2.0.0

replace intel/isecl/lib/kms-client => github.com/intel-secl/kms-client v2.0.0

replace intel/isecl/lib/clients => github.com/intel-secl/clients v2.0.0

replace intel/isecl/lib/mtwilson-client => github.com/intel-secl/mtwilson-client v2.0.0
