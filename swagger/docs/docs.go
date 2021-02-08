// Workload Service
//
// Workload Service resources are used to manage images, flavors and reports.
// Workload Service handles the mapping of the image ID to the appropriate key ID in the form of Flavors.
// When the encrypted image is used to launch new VM or container, WLA will request the decryption key from the Workload Service.
// Then Workload Service will initiate the key transfer request to the Key Broker.
//
//  License: Copyright (C) 2020 Intel Corporation. SPDX-License-Identifier: BSD-3-Clause
//
//  Version: 2.2
//  Host: workloadservice.com:5000
//  BasePath: /wls
//
//  Schemes: https
//
//  SecurityDefinitions:
//   bearerAuth:
//     type: apiKey
//     in: header
//     name: Authorization
//     description: Enter your bearer token in the format **Bearer &lt;token&gt;**
//
// swagger:meta
package docs

// swagger:operation GET /version Version getVersion
// ---
// description: Retrieves the version of workload service.
//
// produces:
// - text/plain
// responses:
//   "200":
//     description: Successfully retrieved the version of workload service.
//     schema:
//       type: string
//       example: v2.2
//
// x-sample-call-endpoint: https://workloadservice.com:5000/wls/v1/version
// x-sample-call-output: v2.2
// ---
