package mock

import (
	"intel/isecl/lib/common/pkg/vm"
	"intel/isecl/lib/flavor"
	"intel/isecl/lib/verifier"
	"intel/isecl/workload-service/model"
)

var f model.Flavor
var f2, _ = flavor.GetImageFlavor("Cirros-enc", true, "http://localhost:1337/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
var i = model.Image{}
var r = model.Report{
	ID: "ffff021e-9669-4e53-9224-8880fb4e4080",
	VMTrustReport: verifier.VMTrustReport{
		Manifest: vm.Manifest{
			VmInfo: vm.Info{
				VmID:             "0000021e-9669-4e53-9224-8880fb4e4080",
				HostHardwareUUID: "0000021e-9669-4e53-9224-8880fb4e4080",
				ImageID:          "0000021e-9669-4e53-9224-8880fb4e4080",
			},
			ImageEncrypted: true,
		},
		PolicyName: "Intel VM Policy",
		Results:    nil,
		Trusted:    true,
	},
}

func init() {
	f = model.Flavor(*f2)
	f2.Image.Meta.ID = "ecee021e-9669-4e53-9224-8880fb4e4080"
	i.ID = "dddd021e-9669-4e53-9224-8880fb4e4080"
	i.FlavorIDs = []string{f.Image.Meta.ID}
	r.ID = "eeee021e-9669-4e53-9224-8880fb4e4080"
}
