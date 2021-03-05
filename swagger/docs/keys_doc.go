/*
 * Copyright (C) 2020 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package docs

import "github.com/intel-secl/intel-secl/v3/pkg/model/wls"

// KeyRequest request payload
// swagger:parameters KeyRequest
type KeyRequest struct {
	// in:body
	Body wls.RequestKey
}

// KeyResponse response payload
// swagger:response KeyResponse
type KeyResponse struct {
	// in:body
	Body wls.ReturnKey
}

// swagger:operation POST /keys Keys CreateKey
// ---
//
// description: |
//   Gets and returns the wrapped key from KBS for given key url, if saml report from hvs for given host with hardware uuid is trusted.
//
// security:
//  - bearerAuth: []
// consumes:
//  - application/json
// produces:
//  - application/json
// parameters:
// - name: request body
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/RequestKey"
// responses:
//   '201':
//     description: Successfully return wrapped key from KBS
//     schema:
//       "$ref": "#/definitions/ReturnKey"
//
// x-sample-call-endpoint: https://workloadservice.com:5000/wls/v1/keys
// x-sample-call-input: |
//  {
//      "hardware_uuid": "ecee021e-9669-4e53-9224-8880fb4e4080"
//      "key_url": "http://kbs.server.com:9443/v1/keys/73755fda-c910-46be-821f-e8ddeab189e9/transfer"
//  }
// x-sample-call-output: |
//  {
//   "key": "eyJpbnN0YW5jZV9tYW5pZmVzdCI6eyJpbnN0YW5jZV9pbmZvIjp7Imluc3RhbmNlX2lkIjoiN2Y4MDMwMTgtZjU2Zi00NWJiLTk0Mm
//    EtODhmYjgzOGNhMjMxIiwiaG9zdF9oYXJkd2FyZV91dWlkIjoiODA4YjcwNmYtNTYzMS1lNTExLTkwNmUtMDAxMjc5NWQ5NmRkIiwiaW1hZ2VfaW
//    QiOiIxMjAwMjQwMC1kMDZiLTRjOWItYWUzYi1hZDQ2MmNjZWI2NzQifSwiaW1hZ2VfZW5jcnlwdGVkIjp0cnVlfSwicG9saWN5X25hbWUiOiJJbn
//    RlbCBDb250YWluZXIgUG9saWN5IiwicmVzdWx0cyI6W3sicnVsZSI6eyJydWxlX25hbWUiOiJFbmNyeXB0aW9uTWF0Y2hlcyIsIm1hcmtlcnMiOl
//    siQ09OVEFJTkVSX0lNQUdFIl0sImV4cGVjdGVkIjp7Im5hbWUiOiJlbmNyeXB0aW9uX3JlcXVpcmVkIiwidmFsdWUiOnRydWV9fSwiZmxhdm9yX2
//    lkIjoiNjkyMWM5NWQtMTRhOC00ZWE4LTk0OWQtZDMzOGQ4OGE0NDdmIiwidHJ1c3RlZCI6dHJ1ZX0seyJydWxlIjp7InJ1bGVfbmFtZSI6IkludG
//    Vncml0eU1hdGNoZXMiLCJtYXJrZXJzIjpbIkNPTlRBSU5FUl9JTUFHRSJdLCJleHBlY3RlZCI6eyJuYW1lIjoiaW50ZWdyaXR5X2VuZm9yY2VkIiw
//    idmFsdWUiOmZhbHNlfX0sImZsYXZvcl9pZCI6IjY5MjFjOTVkLTE0YTgtNGVhOC05NDlkLWQzMzhkODhhNDQ3ZiIsInRydXN0ZWQiOnRydWV9LHsi
//    cnVsZSI6eyJydWxlX25hbWUiOiJGbGF2b3JJbnRlZ3JpdHlNYXRjaGVzIiwibWFya2VycyI6WyJmbGF2b3JJbnRlZ3JpdHkiXSwiZXhwZWN0ZWQiO
//    nsibmFtZSI6ImZsYXZvcl90cnVzdGVkIiwidmFsdWUiOnRydWV9fSwiZmxhdm9yX2lkIjoiNjkyMWM5NWQtMTRhOC00ZWE4LTk0OWQtZDMzOGQ4OG
//    E0NDdmIiwidHJ1c3RlZCI6dHJ1ZX1dLCJ0cnVzdGVkIjp0cnVlfQ==",
//  }
