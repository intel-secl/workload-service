/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package mock

import (
	"encoding/json"
	commLog "intel/isecl/lib/common/v3/log"
	"intel/isecl/lib/common/v3/pkg/instance"
	"intel/isecl/lib/flavor/v3"
	flavorUtil "intel/isecl/lib/flavor/v3/util"
	"intel/isecl/lib/verifier/v3"
	"intel/isecl/workload-service/v3/model"
)

var log = commLog.GetDefaultLogger()

var f model.Flavor
var signedFlavor flavor.SignedImageFlavor
var f2, _ = flavor.GetImageFlavor("Cirros-enc", true, "http://localhost:1337/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer", "1160f92d07a3e9bf2633c49bfc2654428c517ee5a648d715bf984c83f266a4fd")
var flavorBytes, _ = json.Marshal(f2)
var signedFlavorString, err = flavorUtil.GetSignedFlavor(string(flavorBytes), "../repository/mock/flavor-signing-key.pem")
var i = model.Image{}
var r = model.Report{
	ID: "ffff021e-9669-4e53-9224-8880fb4e4080",
	InstanceTrustReport: verifier.InstanceTrustReport{
		Manifest: instance.Manifest{
			InstanceInfo: instance.Info{
				InstanceID:       "0000021e-9669-4e53-9224-8880fb4e4080",
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
	log.Trace("repository/mock/init:init() Entering")
	defer log.Trace("repository/mock/init:init() Leaving")

	f = model.Flavor(*f2)
	log.Info("repository/mock/init:init() Unmarshal the signed flavor")
	json.Unmarshal([]byte(signedFlavorString), &signedFlavor)
	f2.Image.Meta.ID = "ecee021e-9669-4e53-9224-8880fb4e4080"
	i.ID = "dddd021e-9669-4e53-9224-8880fb4e4080"
	i.FlavorIDs = []string{f.Image.Meta.ID}
	r.ID = "eeee021e-9669-4e53-9224-8880fb4e4080"
}
