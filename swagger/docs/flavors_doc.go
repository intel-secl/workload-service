/*
 * Copyright (C) 2020 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package docs

import (
	flvr "intel/isecl/lib/flavor/v2"
)

// FlavorCreateInfo request payload
// swagger:parameters FlavorCreateInfo
type FlavorCreateInfo struct {
	// in:body
	Body flvr.SignedImageFlavor
}

// FlavorResponse response payload
// swagger:response FlavorResponse
type FlavorResponse struct {
	// in:body
	Body flvr.ImageFlavor
}

type FlavorsResponse []flvr.ImageFlavor

// FlavorsResponse response payload
// swagger:response FlavorsResponse
type SwaggFlavorsResponse struct {
	// in:body
	Body FlavorsResponse
}


// swagger:operation POST /flavors Flavors createFlavor
// ---
//
// description: |
//   Creates a flavor for the encrypted image in the workload service database. 
//   Flavor can be created by providing the image flavor content obtained from the WPM after encrypting the image.
//   A valid bearer token should be provided to authorize this REST call.
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
//     "$ref": "#/definitions/SignedImageFlavor"
// responses:
//   '201':
//     description: Successfully created the flavor.
//     schema:
//       "$ref": "#/definitions/ImageFlavor"
//
// x-sample-call-endpoint: https://workloadservice.com:5000/wls/flavors
// x-sample-call-input: |
//    {
//       "flavor": {
//          "meta": {
//             "id": "d6129610-4c8f-4ac4-8823-df4e925688c4",
//             "description": {
//                "flavor_part": "CONTAINER_IMAGE",
//                "label": "label_image-test-4"
//             }
//          },
//          "encryption_required": true,
//          "encryption": {
//             "key_url": "https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer",
//             "digest": "3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="
//          }
//       },
//       "signature": "CStRpWgj0De7+xoX1uFSOacLAZeEcodUuvH62B4hVoiIEriVaHxrLJhBjnIuSPmIoZewCdTShw7GxmMQiMik
//       CrVhaUilYk066TckOcLW/E3K+7NAiZ5kuS96J6dVxgJ+9k7iKf7Z+6lnWUJz92VWLP4U35WK4MtV+MPTYn2Zj1p+/tTUuSqlk8
//       KCmpywzI1J1/XXjvqee3M9cGInnbOUGEFoLBAO1+w30yptoNxKEaB/9t3qEYywk8buT5GEMYUjJEj9PGGaW+lR37x0zcXggwMg
//       /RsijMV6rNKsjjC0fN1vGswzoaIJPD1RJkQ8X9l3AaM0qhLBQDrurWxKK4KSQSpI0BziGPkKi5vAeeRkVfU5JXNdPxdOkyXVeb
//       eMQR9bYntXtZl41qjOZ0zIOKAHNJiBLyMYausbTZHVCwDuA/HBAT8i7JAIesxexX89bL+khPebHWkHaifS4NejymbGzM+n62EH
//       uoeIo33qDMQ/U0FA3i6gRy0s/sFQVXR0xk8l"
//    }
// x-sample-call-output: |
//  {
//    "flavor": {
//        "meta": {
//            "id": "d6129610-4c8f-4ac4-8823-df4e925688c4",
//            "description": {
//                "flavor_part": "CONTAINER_IMAGE",
//                "label": "label_image-test-4"
//            }
//        },
//        "encryption_required": true,
//        "encryption": {
//            "key_url": "https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer",
//            "digest": "3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="
//        },
//        "integrity_enforced": false
//    }
//  }
// ---

// swagger:operation GET /flavors Flavors queryFlavors
// ---
// description: |
//   Search(es) for the flavor(s) based on the provided filter criteria in the workload service database.
//   A valid bearer token should be provided to authorize this REST call.
//
// security:
//  - bearerAuth: []
// produces:
//  - application/json
// parameters:
// - name: filter 
//   description: |
//      Boolean value to indicate whether the response should be filtered to return specific flavors instead of listing 
//      all flavors. When the filter is true and no other query parameter is specified, error response will be returned. 
//      Default value is true.
//   in: query
//   type: boolean
// - name: id
//   description: Unique ID of the flavor.
//   in: query
//   type: string
//   format: uuid
// - name: label
//   description: Label associated with the flavor.
//   in: query
//   type: string
// responses:
//   '200':
//     description: Successfully retrieved the flavors based on the provided filter criteria.
//     schema:
//       "$ref": "#/definitions/FlavorsResponse"
//
// x-sample-call-endpoint: https://workloadservice.com:5000/wls/flavors?label=label_image-test-4
// x-sample-call-output: |
//  [
//    {
//        "flavor": {
//            "meta": {
//                "id": "d6129610-4c8f-4ac4-8823-df4e925688c4",
//                "description": {
//                    "flavor_part": "CONTAINER_IMAGE",
//                    "label": "label_image-test-4"
//                }
//            },
//            "encryption_required": true,
//            "encryption": {
//                "key_url": "https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer",
//                "digest": "3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="
//            },
//            "integrity_enforced": false
//        }
//    }
//  ]
// ---

// swagger:operation DELETE /flavors/{flavor_id} Flavors deleteFlavorByID
// ---
// description: |
//   Deletes the flavor associated with a specified flavor id from the workload service 
//   database. A valid bearer token should be provided to authorize this REST call.
//
// security:
//  - bearerAuth: []
// produces:
//  - application/json
// parameters:
// - name: flavor_id
//   description: Unique ID of the flavor.
//   in: path
//   required: true
//   type: string
//   format: uuid
// responses:
//   '204':
//     description: Successfully deleted the flavor for the specified flavor id.
//
// x-sample-call-endpoint: |
//    https://workloadservice.com:5000/wls/flavors/d6129610-4c8f-4ac4-8823-df4e925688c4
// x-sample-call-output: |
//   204 No content
// ---

// swagger:operation GET /flavors/{flavor_id} Flavors getFlavorByIdOrLabel
// ---
// description: |
//   Retrieves the flavor associated with a specified flavor ID or flavor label from the workload service 
//   database. The path parameter can be either flavor ID or flavor Label. 
//   A valid bearer token should be provided to authorize this REST call.
//
// security:
//  - bearerAuth: []
// produces:
//  - application/json
// parameters:
// - name: flavor_id
//   description: Unique ID of the flavor.
//   in: path
//   required: true
//   type: string
// responses:
//   '200':
//     description: Successfully retrieved the flavor for the specified flavor id or flavor label.
//     schema:
//       "$ref": "#/definitions/ImageFlavor"
//
// x-sample-call-endpoint: |
//    https://workloadservice.com:5000/wls/flavors/d6129610-4c8f-4ac4-8823-df4e925688c4
// x-sample-call-output: |
//  {
//    "flavor": {
//        "meta": {
//            "id": "d6129610-4c8f-4ac4-8823-df4e925688c4",
//            "description": {
//                "flavor_part": "CONTAINER_IMAGE",
//                "label": "label_image-test-4"
//            }
//        },
//        "encryption_required": true,
//        "encryption": {
//            "key_url": "https://10.105.168.234:443/v1/keys/60a9fe49-612f-4b66-bf86-b75c7873f3b3/transfer",
//            "digest": "3JiqO+O4JaL2qQxpzRhTHrsFpDGIUDV8fTWsXnjHVKY="
//        },
//        "integrity_enforced": false
//    }
//  }
// ---
