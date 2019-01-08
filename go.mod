module intel/isecl/workload-service

require (
	github.com/gorilla/mux v1.6.2
	github.com/jinzhu/gorm v1.9.2
	github.com/jinzhu/inflection v0.0.0-20180308033659-04140366298a // indirect
	github.com/lib/pq v1.0.0 // indirect
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
	github.com/stretchr/testify v1.2.2
	gopkg.in/yaml.v2 v2.2.2
	intel/isecl/lib/common v0.0.0
	intel/isecl/lib/flavor v0.0.0
	intel/isecl/lib/kms-client v0.0.0
	intel/isecl/lib/middleware v0.0.0
	intel/isecl/lib/verifier v0.0.0
)

replace intel/isecl/lib/flavor => gitlab.devtools.intel.com/sst/isecl/lib/flavor v0.0.0-20181206181952-1ec1e81fed41

replace intel/isecl/lib/middleware => gitlab.devtools.intel.com/sst/isecl/lib/middleware v0.0.0-20181204232340-f34e00324e43

replace intel/isecl/lib/common => gitlab.devtools.intel.com/sst/isecl/lib/common v0.0.0-20190102215830-f9c9e2da84e8

replace intel/isecl/lib/verifier => gitlab.devtools.intel.com/sst/isecl/lib/verifier v0.0.0-20181221212114-b1d5e4114406

replace intel/isecl/lib/kms-client => gitlab.devtools.intel.com/sst/isecl/lib/kms-client v0.0.0-20190108194113-346ab5daaac1

//replace intel/isecl/lib/kms-client => ../lib/kms-client
