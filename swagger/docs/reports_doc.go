package docs

import (
        "intel/isecl/lib/common/v2/crypt"
        "intel/isecl/lib/verifier/v2"
	"intel/isecl/lib/common/v2/pkg/instance"
)

type Rule struct {
	verifier.EncryptionMatches
}

type Result struct {
        Rule     Rule    `json:"rule"`
        FlavorID string  `json:"flavor_id"`
        Faults   []verifier.Fault `json:"faults,omitempty"`
        Trusted  bool    `json:"trusted"`
}

type InstanceTrustReport struct {
        Manifest   instance.Manifest `json:"instance_manifest"`
        PolicyName string         `json:"policy_name"`
        Results    []Result       `json:"results"`
        Trusted    bool           `json:"trusted"`
}

type Report struct {
        ID string `json:"id,omitempty"`
        InstanceTrustReport
        crypt.SignedData
}

// ReportCreateInfo request payload
// swagger:parameters ReportCreateInfo
type ReportCreateInfo struct {
	// in:body
	Body Report
}

type ReportsResponse []Report

// ReportsResponse response payload
// swagger:response ReportsResponse
type SwaggReportsResponse struct {
	// in:body
	Body ReportsResponse
}


// swagger:operation POST /reports Reports createReport
// ---
//
// description: |
//   Creates an image trust report. A report contains the status of an associated image. 
//   The report schema provided in the request body contains an interface called Rule which works on 
//   any matching policy based on provided rule_name. Rule policy can be either image encryption policy 
//   or flavor integrity policy or integrity policy. 
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
//     "$ref": "#/definitions/Report"
// responses:
//   '201':
//     description: Successfully created the trust report for the image.
//     schema:
//       "$ref": "#/definitions/Report"
//
// x-sample-call-endpoint: https://workloadservice.com:5000/wls/reports
// x-sample-call-input: |
//  {
//   "id": "f52023eb-7991-47ba-91fc-c43bd9d80c29",
//   "instance_manifest": {
//       "instance_info": {
//       "instance_id": "7f803018-f56f-45bb-942a-88fb838ca231",
//       "host_hardware_uuid": "808b706f-5631-e511-906e-0012795d96dd",
//       "image_id": "12002400-d06b-4c9b-ae3b-ad462cceb674"
//       },
//       "image_encrypted": true
//   },
//   "policy_name": "Intel Container Policy",
//   "results": [
//      {
//         "rule": {
//            "rule_name": "EncryptionMatches",
//            "markers": [
//               "CONTAINER_IMAGE"
//            ],
//            "expected": {
//               "name": "encryption_required",
//               "value": true
//             }
//          },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//      },
//      {
//          "rule": {
//            "rule_name": "IntegrityMatches",
//            "markers": [
//               "CONTAINER_IMAGE"
//            ],
//            "expected": {
//            "name": "integrity_enforced",
//            "value": false
//            }
//          },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//       },
//       {
//          "rule": {
//            "rule_name": "FlavorIntegrityMatches",
//            "markers": [
//               "flavorIntegrity"
//            ],
//            "expected": {
//            "name": "flavor_trusted",
//            "value": true
//            }
//         },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//       }
//   ],
//   "trusted": true,
//   "data": "eyJpbnN0YW5jZV9tYW5pZmVzdCI6eyJpbnN0YW5jZV9pbmZvIjp7Imluc3RhbmNlX2lkIjoiN2Y4MDMwMTgtZjU2Zi00NWJiLTk0Mm
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
//   "hash_alg": "SHA-256",
//   "cert": "-----BEGIN CERTIFICATE-----
//   MIIFEDCCA3igAwIBAgIIbb/wqvbCU/swDQYJKoZIhvcNAQEMBQAwGzEZMBcGA1UEAxMQbXR3aWxzb24tcGNhLWFpazA
//   eFw0xOTA4MjMxMjQxMzlaFw0yOTA4MjAxMjQxMzlaMCUxIzAhBgNVBAMMGkNOPVNpZ25pbmdfS2V5X0NlcnRpZmljYX
//   RlMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtykLA0i8F/i8/XaxxiOQODQ8DBtmHFUmagQt0/Uanfe7g
//   ZIlDWXidc7JJ0c4BaO8TACTOU2RH3aI4qOdTM3KzCERyssuJPFpJyhAbMlQcW1GqO/ZnH1SinnrDnDuCaFE1H1oWldW
//   GIJFG964Chlu1tDBaSpHqhrSe9HFT2ne+k8SbWNCKQp99UJWwvE9ZIb8HI42I+duwkazPEKGcL64AKhx+/9su3qJ51k
//   PaIp3LnPlFWj8GTQ1hPCTCxMZcc6b/BdGjfQYuz7XFL8Vnwx53ILidMsAf/7WWT9dc99yYFIl4AtzcToAIV28gRBzRe
//   6AVQPa8tPtgixg34LIwDGZZQIDAQABo4IBzDCCAcgwDgYDVR0PAQH/BAQDAgbAMIGdBgdVBIEFAwIpBIGR/1RDR4AXA
//   CIAC6PFIigCakiZJBZJguUgZk6BUJXqvvNG5JPeoTyYzN3uAAQA/1WqAAAAAAA7OO4AAAAEAAAAAQAABAABFQQVCQAi
//   AAsx6pP2c46F7wuGDjaihI06cBKS3tSCGFm4ekaqR45d0gAiAAuh4F4oZZ1bUYKd4p5cle2322+fKD09GJ8/NSyl1FF
//   E3TCCARQGCFUEgQUDAikBBIIBBgAUAAsBAI/eR4gVGaVG3DGO4n8bhmeBHlVgeuUkWyDet3oA+GZqtITbVM4PsChznS
//   7l2hZx3HkevL2LWX3WycR0VSR1uylWVeHCJUJjwArA2i9LhItFZdoUlEvVMnfzaGdoa2m2VuHs7NUDW1Nd0rE9eAlHA
//   EdUzCIE2wNr7dloHZHs6S7F3KJrWH5ATGfEayiBYg+/hj4jXqxyW1VNGNI6KdKzZyzk+QlGQkFtdRpBTqUhrZBp4FYD
//   lKq81PuJ3ayVS1FY6vQ2Nz70ueGoF1K73qbEAFTcBuMHzpU8ZS5JUO9eEeWHRCn+BbA1KMB6576pVAXz//e3/sEDsCX
//   nlieYETzPbA8wDQYJKoZIhvcNAQEMBQADggGBAJuupBNdZwCZ6gDoZEOYr0SeM6O+xK0Q/Oe7ecPjTGXrdsgbul5DnI
//   prIav7fZ4e7q+utzh8l9uij3wAfOM5gT1ud1TJLuCyl70e3VSHcuGBCRcFC4xXGykT+UqfdnqCnkSnQbQUAIBr46zoy
//   Ap1P4/QdMBDJo/BjPQTdThFValrp95EfNTavFkXogryEKmWMANgHh+ZNyQfXunjjaaS99yFdoPGgoLFHbVQ1ehUP3m5
//   4X1la90+59TccPD01JHibA9Tp9hsDA9NVGf3tBokoEzPBgu9ilHUR7sZ+C6CZ3YjHhcghJoau03UzUQ/pyLGs+lV9U+
//   PpCr4i/w4OgOOOfh0u0vJrFf3x7BiFikwgcTJHspvdRzAkNAndSNH6PnJrpd+Q/s1mlEwytUumycenIRLwDmWGYCSGB
//   xcU6IfSaeUPI6JROuhFjNjb2N6CpiswU1pWAGKVLi3a5BcmPnJc7XQcG+vb9sAX9BVmwL1sRqC1Dz0IBjLz8eYzebLu
//   uMb8Q==
//    -----END CERTIFICATE-----",
//   "signature": "tcWuM6dk/a0XYFcbqSpDIe7BvN/EsX2CskB6xecryFhXS3HbbeB97K6GqI/TQnZZPC40KfQDUTVn7oSDH9AvnFIDQSsBUCqcfl0Q0CRdm9KE9brCT
//    zwxPTHdN4kcC8I2iMxQVzsqV/TL39QUGlwhtYLJOqEaJ5sDqbmnW9XVO6TaEkcagHoF3Je+iglVbMmL7cfpySlppU9+TO6tSYlLP/nX53mEhBgN4ANSQ0gKNQR3izG
//    Ca8mHgqGWnoahrR/dHHWADrKDF6SeMbmvo0e4/EOXaO89QyPrGwF8kuzI04HpXUWlOYfaubKkNHi3gckXFzNB84nhrp0SNz3RHxNEkA=="
//  }
// x-sample-call-output: |
//  {
//   "id": "f52023eb-7991-47ba-91fc-c43bd9d80c29",
//   "instance_manifest": {
//       "instance_info": {
//       "instance_id": "7f803018-f56f-45bb-942a-88fb838ca231",
//       "host_hardware_uuid": "808b706f-5631-e511-906e-0012795d96dd",
//       "image_id": "12002400-d06b-4c9b-ae3b-ad462cceb674"
//       },
//       "image_encrypted": true
//   },
//   "policy_name": "Intel Container Policy",
//   "results": [
//      {
//         "rule": {
//            "rule_name": "EncryptionMatches",
//            "markers": [
//               "CONTAINER_IMAGE"
//            ],
//            "expected": {
//               "name": "encryption_required",
//               "value": true
//             }
//          },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//      },
//      {
//          "rule": {
//            "rule_name": "IntegrityMatches",
//            "markers": [
//               "CONTAINER_IMAGE"
//            ],
//            "expected": {
//            "name": "integrity_enforced",
//            "value": false
//            }
//          },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//       },
//       {
//          "rule": {
//            "rule_name": "FlavorIntegrityMatches",
//            "markers": [
//               "flavorIntegrity"
//            ],
//            "expected": {
//            "name": "flavor_trusted",
//            "value": true
//            }
//         },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//       }
//   ],
//   "trusted": true,
//   "data": "eyJpbnN0YW5jZV9tYW5pZmVzdCI6eyJpbnN0YW5jZV9pbmZvIjp7Imluc3RhbmNlX2lkIjoiN2Y4MDMwMTgtZjU2Zi00NWJiLTk0Mm
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
//   "hash_alg": "SHA-256",
//   "cert": "-----BEGIN CERTIFICATE-----
//   MIIFEDCCA3igAwIBAgIIbb/wqvbCU/swDQYJKoZIhvcNAQEMBQAwGzEZMBcGA1UEAxMQbXR3aWxzb24tcGNhLWFpazA
//   eFw0xOTA4MjMxMjQxMzlaFw0yOTA4MjAxMjQxMzlaMCUxIzAhBgNVBAMMGkNOPVNpZ25pbmdfS2V5X0NlcnRpZmljYX
//   RlMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtykLA0i8F/i8/XaxxiOQODQ8DBtmHFUmagQt0/Uanfe7g
//   ZIlDWXidc7JJ0c4BaO8TACTOU2RH3aI4qOdTM3KzCERyssuJPFpJyhAbMlQcW1GqO/ZnH1SinnrDnDuCaFE1H1oWldW
//   GIJFG964Chlu1tDBaSpHqhrSe9HFT2ne+k8SbWNCKQp99UJWwvE9ZIb8HI42I+duwkazPEKGcL64AKhx+/9su3qJ51k
//   PaIp3LnPlFWj8GTQ1hPCTCxMZcc6b/BdGjfQYuz7XFL8Vnwx53ILidMsAf/7WWT9dc99yYFIl4AtzcToAIV28gRBzRe
//   6AVQPa8tPtgixg34LIwDGZZQIDAQABo4IBzDCCAcgwDgYDVR0PAQH/BAQDAgbAMIGdBgdVBIEFAwIpBIGR/1RDR4AXA
//   CIAC6PFIigCakiZJBZJguUgZk6BUJXqvvNG5JPeoTyYzN3uAAQA/1WqAAAAAAA7OO4AAAAEAAAAAQAABAABFQQVCQAi
//   AAsx6pP2c46F7wuGDjaihI06cBKS3tSCGFm4ekaqR45d0gAiAAuh4F4oZZ1bUYKd4p5cle2322+fKD09GJ8/NSyl1FF
//   E3TCCARQGCFUEgQUDAikBBIIBBgAUAAsBAI/eR4gVGaVG3DGO4n8bhmeBHlVgeuUkWyDet3oA+GZqtITbVM4PsChznS
//   7l2hZx3HkevL2LWX3WycR0VSR1uylWVeHCJUJjwArA2i9LhItFZdoUlEvVMnfzaGdoa2m2VuHs7NUDW1Nd0rE9eAlHA
//   EdUzCIE2wNr7dloHZHs6S7F3KJrWH5ATGfEayiBYg+/hj4jXqxyW1VNGNI6KdKzZyzk+QlGQkFtdRpBTqUhrZBp4FYD
//   lKq81PuJ3ayVS1FY6vQ2Nz70ueGoF1K73qbEAFTcBuMHzpU8ZS5JUO9eEeWHRCn+BbA1KMB6576pVAXz//e3/sEDsCX
//   nlieYETzPbA8wDQYJKoZIhvcNAQEMBQADggGBAJuupBNdZwCZ6gDoZEOYr0SeM6O+xK0Q/Oe7ecPjTGXrdsgbul5DnI
//   prIav7fZ4e7q+utzh8l9uij3wAfOM5gT1ud1TJLuCyl70e3VSHcuGBCRcFC4xXGykT+UqfdnqCnkSnQbQUAIBr46zoy
//   Ap1P4/QdMBDJo/BjPQTdThFValrp95EfNTavFkXogryEKmWMANgHh+ZNyQfXunjjaaS99yFdoPGgoLFHbVQ1ehUP3m5
//   4X1la90+59TccPD01JHibA9Tp9hsDA9NVGf3tBokoEzPBgu9ilHUR7sZ+C6CZ3YjHhcghJoau03UzUQ/pyLGs+lV9U+
//   PpCr4i/w4OgOOOfh0u0vJrFf3x7BiFikwgcTJHspvdRzAkNAndSNH6PnJrpd+Q/s1mlEwytUumycenIRLwDmWGYCSGB
//   xcU6IfSaeUPI6JROuhFjNjb2N6CpiswU1pWAGKVLi3a5BcmPnJc7XQcG+vb9sAX9BVmwL1sRqC1Dz0IBjLz8eYzebLu
//   uMb8Q==
//    -----END CERTIFICATE-----",
//   "signature": "tcWuM6dk/a0XYFcbqSpDIe7BvN/EsX2CskB6xecryFhXS3HbbeB97K6GqI/TQnZZPC40KfQDUTVn7oSDH9AvnFIDQSsBUCqcfl0Q0CRdm9KE9brCT
//    zwxPTHdN4kcC8I2iMxQVzsqV/TL39QUGlwhtYLJOqEaJ5sDqbmnW9XVO6TaEkcagHoF3Je+iglVbMmL7cfpySlppU9+TO6tSYlLP/nX53mEhBgN4ANSQ0gKNQR3izG
//    Ca8mHgqGWnoahrR/dHHWADrKDF6SeMbmvo0e4/EOXaO89QyPrGwF8kuzI04HpXUWlOYfaubKkNHi3gckXFzNB84nhrp0SNz3RHxNEkA=="
//  }

// ---

// swagger:operation GET /reports Reports queryReports
// ---
// description: |
//   Search(es) for the trust report(s) based on filter criteria in the workload service database.
//   Minimum one query parameter should be provided to retrieve the reports.
//   A valid bearer token should be provided to authorize this REST call.
//
// security:
//  - bearerAuth: []
// produces:
//  - application/json
// parameters:
// - name: filter 
//   description: |
//      Boolean value to indicate whether the response should be filtered to return no results instead of listing all reports. 
//      When the filter is true and no other query parameter is specified, error response will be returned. Default value is true.
//   in: query
//   type: boolean
// - name: instance_id
//   description: Unique ID of the VM.
//   in: query
//   type: string
//   format: uuid
// - name: report_id
//   description: Unique ID of the report.
//   in: query
//   type: string
//   format: uuid
// - name: hardware_uuid
//   description: Unique hardware UUID of the host.
//   in: query
//   type: string
//   format: uuid
// - name: from_date
//   description: Reports returned will be restricted to after this date. from_date should be given in date format yyyy-mm-ddTHH:mm:ss.
//   in: query
//   type: string
// - name: to_date
//   description: Reports returned will be restricted to before this date. to_date should be given in date format yyyy-mm-ddTHH:mm:ss.
//   in: query
//   type: string
// - name: latest_per_vm
//   description: |
//      By default this is set to TRUE, returning only the latest report for each VM.
//   in: query
//   type: boolean
// - name: num_of_days
//   description: |
//      Results returned will be restricted to between the current date and number of days prior. 
//      This option will override other date options.
//   in: query
//   type: integer
// responses:
//   '200':
//     description: Successfully retrieved the reports based on filter criteria.
//     schema:
//       "$ref": "#/definitions/ReportsResponse"
//
// x-sample-call-endpoint: https://workloadservice.com:5000/wls/reports?report_id=f52023eb-7991-47ba-91fc-c43bd9d80c29
// x-sample-call-output: |
//  {
//   "id": "f52023eb-7991-47ba-91fc-c43bd9d80c29",
//   "instance_manifest": {
//       "instance_info": {
//       "instance_id": "7f803018-f56f-45bb-942a-88fb838ca231",
//       "host_hardware_uuid": "808b706f-5631-e511-906e-0012795d96dd",
//       "image_id": "12002400-d06b-4c9b-ae3b-ad462cceb674"
//       },
//       "image_encrypted": true
//   },
//   "policy_name": "Intel Container Policy",
//   "results": [
//      {
//         "rule": {
//            "rule_name": "EncryptionMatches",
//            "markers": [
//               "CONTAINER_IMAGE"
//            ],
//            "expected": {
//               "name": "encryption_required",
//               "value": true
//             }
//          },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//      },
//      {
//          "rule": {
//            "rule_name": "IntegrityMatches",
//            "markers": [
//               "CONTAINER_IMAGE"
//            ],
//            "expected": {
//            "name": "integrity_enforced",
//            "value": false
//            }
//          },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//       },
//       {
//          "rule": {
//            "rule_name": "FlavorIntegrityMatches",
//            "markers": [
//               "flavorIntegrity"
//            ],
//            "expected": {
//            "name": "flavor_trusted",
//            "value": true
//            }
//         },
//          "flavor_id": "6921c95d-14a8-4ea8-949d-d338d88a447f",
//          "trusted": true
//       }
//   ],
//   "trusted": true,
//   "data": "eyJpbnN0YW5jZV9tYW5pZmVzdCI6eyJpbnN0YW5jZV9pbmZvIjp7Imluc3RhbmNlX2lkIjoiN2Y4MDMwMTgtZjU2Zi00NWJiLTk0Mm
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
//   "hash_alg": "SHA-256",
//   "cert": "-----BEGIN CERTIFICATE-----
//   MIIFEDCCA3igAwIBAgIIbb/wqvbCU/swDQYJKoZIhvcNAQEMBQAwGzEZMBcGA1UEAxMQbXR3aWxzb24tcGNhLWFpazA
//   eFw0xOTA4MjMxMjQxMzlaFw0yOTA4MjAxMjQxMzlaMCUxIzAhBgNVBAMMGkNOPVNpZ25pbmdfS2V5X0NlcnRpZmljYX
//   RlMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtykLA0i8F/i8/XaxxiOQODQ8DBtmHFUmagQt0/Uanfe7g
//   ZIlDWXidc7JJ0c4BaO8TACTOU2RH3aI4qOdTM3KzCERyssuJPFpJyhAbMlQcW1GqO/ZnH1SinnrDnDuCaFE1H1oWldW
//   GIJFG964Chlu1tDBaSpHqhrSe9HFT2ne+k8SbWNCKQp99UJWwvE9ZIb8HI42I+duwkazPEKGcL64AKhx+/9su3qJ51k
//   PaIp3LnPlFWj8GTQ1hPCTCxMZcc6b/BdGjfQYuz7XFL8Vnwx53ILidMsAf/7WWT9dc99yYFIl4AtzcToAIV28gRBzRe
//   6AVQPa8tPtgixg34LIwDGZZQIDAQABo4IBzDCCAcgwDgYDVR0PAQH/BAQDAgbAMIGdBgdVBIEFAwIpBIGR/1RDR4AXA
//   CIAC6PFIigCakiZJBZJguUgZk6BUJXqvvNG5JPeoTyYzN3uAAQA/1WqAAAAAAA7OO4AAAAEAAAAAQAABAABFQQVCQAi
//   AAsx6pP2c46F7wuGDjaihI06cBKS3tSCGFm4ekaqR45d0gAiAAuh4F4oZZ1bUYKd4p5cle2322+fKD09GJ8/NSyl1FF
//   E3TCCARQGCFUEgQUDAikBBIIBBgAUAAsBAI/eR4gVGaVG3DGO4n8bhmeBHlVgeuUkWyDet3oA+GZqtITbVM4PsChznS
//   7l2hZx3HkevL2LWX3WycR0VSR1uylWVeHCJUJjwArA2i9LhItFZdoUlEvVMnfzaGdoa2m2VuHs7NUDW1Nd0rE9eAlHA
//   EdUzCIE2wNr7dloHZHs6S7F3KJrWH5ATGfEayiBYg+/hj4jXqxyW1VNGNI6KdKzZyzk+QlGQkFtdRpBTqUhrZBp4FYD
//   lKq81PuJ3ayVS1FY6vQ2Nz70ueGoF1K73qbEAFTcBuMHzpU8ZS5JUO9eEeWHRCn+BbA1KMB6576pVAXz//e3/sEDsCX
//   nlieYETzPbA8wDQYJKoZIhvcNAQEMBQADggGBAJuupBNdZwCZ6gDoZEOYr0SeM6O+xK0Q/Oe7ecPjTGXrdsgbul5DnI
//   prIav7fZ4e7q+utzh8l9uij3wAfOM5gT1ud1TJLuCyl70e3VSHcuGBCRcFC4xXGykT+UqfdnqCnkSnQbQUAIBr46zoy
//   Ap1P4/QdMBDJo/BjPQTdThFValrp95EfNTavFkXogryEKmWMANgHh+ZNyQfXunjjaaS99yFdoPGgoLFHbVQ1ehUP3m5
//   4X1la90+59TccPD01JHibA9Tp9hsDA9NVGf3tBokoEzPBgu9ilHUR7sZ+C6CZ3YjHhcghJoau03UzUQ/pyLGs+lV9U+
//   PpCr4i/w4OgOOOfh0u0vJrFf3x7BiFikwgcTJHspvdRzAkNAndSNH6PnJrpd+Q/s1mlEwytUumycenIRLwDmWGYCSGB
//   xcU6IfSaeUPI6JROuhFjNjb2N6CpiswU1pWAGKVLi3a5BcmPnJc7XQcG+vb9sAX9BVmwL1sRqC1Dz0IBjLz8eYzebLu
//   uMb8Q==
//    -----END CERTIFICATE-----",
//   "signature": "tcWuM6dk/a0XYFcbqSpDIe7BvN/EsX2CskB6xecryFhXS3HbbeB97K6GqI/TQnZZPC40KfQDUTVn7oSDH9AvnFIDQSsBUCqcfl0Q0CRdm9KE9brCT
//    zwxPTHdN4kcC8I2iMxQVzsqV/TL39QUGlwhtYLJOqEaJ5sDqbmnW9XVO6TaEkcagHoF3Je+iglVbMmL7cfpySlppU9+TO6tSYlLP/nX53mEhBgN4ANSQ0gKNQR3izG
//    Ca8mHgqGWnoahrR/dHHWADrKDF6SeMbmvo0e4/EOXaO89QyPrGwF8kuzI04HpXUWlOYfaubKkNHi3gckXFzNB84nhrp0SNz3RHxNEkA=="
//  }

// ---

// swagger:operation DELETE /reports/{report_id} Reports deleteReportById
// ---
// description: |
//   Deletes the image trust report associated with a specified report id from the workload service 
//   database. A valid bearer token should be provided to authorize this REST call.
//
// security:
//  - bearerAuth: []
// produces:
//  - application/json
// parameters:
// - name: report_id
//   description: Unique ID of the report.
//   in: path
//   required: true
//   type: string
//   format: uuid
// responses:
//   '200':
//     description: Successfully deleted the image trust report associated with specified report id.
//
// x-sample-call-endpoint: |
//    https://workloadservice.com:5000/wls/reports/f52023eb-7991-47ba-91fc-c43bd9d80c29
// x-sample-call-output: |
//    204 No content
// ---
