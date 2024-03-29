/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package resource

import (
	"encoding/base64"
	"fmt"
	"intel/isecl/lib/common/v4/middleware"
	"intel/isecl/workload-service/v4/constants"
	"intel/isecl/workload-service/v4/repository"
	"intel/isecl/workload-service/v4/repository/postgres"
	"net/http"
	"os"
	"testing"

	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

//The JWT signature verifier certificate at ./workload-service/mockJWTDir/jwtVerifier.pem and the
//corresponding bearer token needs to be changed after it expires every 20 years
const BearerToken = "eyJhbGciOiJSUzM4NCIsImtpZCI6IjRiNDA3MmYyNWQ1ZDk1ZWE2NjlmZWRhOWU4NGUzZjJiNWY5ZmM3YzQiLCJ0eXAiOiJKV1QifQ.eyJyb2xlcyI6W3sic2VydmljZSI6IkFBUyIsIm5hbWUiOiJBZG1pbmlzdHJhdG9yIn0seyJzZXJ2aWNlIjoiVEEiLCJuYW1lIjoiQWRtaW5pc3RyYXRvciJ9LHsic2VydmljZSI6IkFIIiwibmFtZSI6IkFkbWluaXN0cmF0b3IifSx7InNlcnZpY2UiOiJIVlMiLCJuYW1lIjoiQWRtaW5pc3RyYXRvciJ9LHsic2VydmljZSI6IktNUyIsIm5hbWUiOiJLZXlDUlVEIn0seyJzZXJ2aWNlIjoiV0xTIiwibmFtZSI6IkFkbWluaXN0cmF0b3IifV0sInBlcm1pc3Npb25zIjpbeyJzZXJ2aWNlIjoiQUFTIiwicnVsZXMiOlsiKjoqOioiXX0seyJzZXJ2aWNlIjoiQUgiLCJydWxlcyI6WyIqOio6KiJdfSx7InNlcnZpY2UiOiJIVlMiLCJydWxlcyI6WyIqOio6KiJdfSx7InNlcnZpY2UiOiJLTVMiLCJydWxlcyI6WyIqOio6KiJdfSx7InNlcnZpY2UiOiJUQSIsInJ1bGVzIjpbIio6KjoqIl19LHsic2VydmljZSI6IldMUyIsInJ1bGVzIjpbIio6KjoqIl19XSwiZXhwIjoyMjI3MjUwNDAzLCJpYXQiOjE1OTY1MzAzNzMsImlzcyI6IkFBUyBKV1QgSXNzdWVyIiwic3ViIjoiZ2xvYmFsX2FkbWluX3VzZXIifQ.mT0IlmD6ZzBKv98maup6EkKQ5qAgFuz0wZ7AjB_O5TukEpcznGZfuXelR8awyDZcuC8wdjvUEubive6ip1QB-_6KV2TFdc85Am8eWRk8eRei0Na3JIh7yEh9rk-Xjv9lcj4uwm-fdNe2vJ7mSxs07gsRB-ufw0YA5fX5Xs_VxCCp3sPgBvSJS5DarRJDLAnbWEPRbnyP0HXnfkwGlQAvHcyi8kYEflOlsLDsUwZC9fxQEJRz2qteSU-BVUYzzlt8nMjSu8X5EDGAI4DVYk1WecO9DxbVWYa2Zu2yUnIbFake6bulTGvD4ahhkHA4anLtC9tgf3hOoHGabl7lplja2XCtGBHU_h4mJcGg-aH4EfM3jXjfwJdhnN_lihbcI7LSQ9yQFDAigALW6xPKLSbpH__cbvFooKw7eRcX6AY1x_8hLhBpnvsivzE51rxchsMJ1QC07HdZQQ_RU5Dcg5Kc2rtRnanlY8G7nZ_XXVmU_EG-rW8dintqZztvSHmStnz9"

var cacheTime, _ = time.ParseDuration(constants.JWTCertsCacheTime)

func setupServer(t *testing.T) *mux.Router {
	log.Trace("resource/common_test:setupServer() Entering")
	defer log.Trace("resource/common_test:setupServer() Leaving")
	_, ci := os.LookupEnv("CI")
	var host string
	if ci {
		host = "postgres"
	} else {
		host = "localhost"
	}
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=5432 user=runner dbname=wls password=test sslmode=disable", host))
	if err != nil {
		t.Fatal("could not open DB")
	}
	r := mux.NewRouter()
	r.Use(middleware.NewTokenAuth("../mockJWTDir", "../mockJWTDir", mockRetrieveJWTSigningCerts, cacheTime))
	wlsDB := postgres.PostgresDatabase{DB: db.Debug()}
	wlsDB.Migrate()
	SetFlavorsEndpoints(r.PathPrefix("/wls/v1/flavors").Subrouter(), wlsDB)
	SetImagesEndpoints(r.PathPrefix("/wls/v1/images").Subrouter(), wlsDB)
	SetReportsEndpoints(r.PathPrefix("/wls/v1/reports").Subrouter(), wlsDB)
	return r
}

func mockRetrieveJWTSigningCerts() error {
	log.Trace("resource/common_test:mockRetrieveJWTSigningCerts() Entering")
	defer log.Trace("resource/common_test:mockRetrieveJWTSigningCerts() Leaving")
	return nil
}

func setupMockServer(db repository.WlsDatabase) *mux.Router {
	log.Trace("resource/common_test:setupMockServer() Entering")
	defer log.Trace("resource/common_test:setupMockServer() Leaving")
	r := mux.NewRouter()
	r.Use(middleware.NewTokenAuth("../mockJWTDir", "../mockJWTDir", mockRetrieveJWTSigningCerts, cacheTime))
	SetFlavorsEndpoints(r.PathPrefix("/wls/v1/flavors").Subrouter(), db)
	SetImagesEndpoints(r.PathPrefix("/wls/v1/images").Subrouter(), db)
	SetReportsEndpoints(r.PathPrefix("/wls/v1/reports").Subrouter(), db)
	return r
}

func mockHVS(addr string) *http.Server {
	log.Trace("resource/common_test:mockHVS() Entering")
	defer log.Trace("resource/common_test:mockHVS() Leaving")
	r := mux.NewRouter()
	r.HandleFunc("/mtwilson/v2/reports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/samlassertion+xml")
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Write([]byte(saml))
	}).Methods("POST")
	h := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	go h.ListenAndServe()
	return h
}

func mockKMS(addr string) *http.Server {
	log.Trace("resource/common_test:mockKMS() Entering")
	defer log.Trace("resource/common_test:mockKMS() Leaving")
	r := mux.NewRouter()
	r.HandleFunc("/v1/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Write([]byte(`{
			"authorization_token": "FkJpk11cfoQt+OVTH1Yg+xssxa6hyVm2riLi3RUHe4U=",
			"authorization_date": "2019-12-18T18:07:16-0800",
			"not_after": "2019-12-18T18:37:16-0800",
			"faults": []
		}`))
	}).Methods("POST")
	r.HandleFunc("/v1/keys/{keyId}/transfer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		enc, _ := base64.StdEncoding.DecodeString(`ibjvgE7lIdDqGrgf3CLY4xeOMdzU6K6c1dZO04U51Z7JomuaQCTgdtUbQUU5eJxnapV3lTO2ev3q
		pmnyCvR1fpwF7n/dQKRDVraLvuElABcJ33uQiVTxjBcCRIDmNRpBNjS0q6f7EuynUrbeqmEVFJWn
		v0U4smZd6s3x6krTP4BiOGttpDiR0TD5N9kbMJMBZvWvERkBMwRED/Nmt9JEdD0s3mHe5zV3G9WX
		ln40773Cczo9awtNfUVdVyDx6LejJcCgkt4XNdRZbK9cVdGK+w6Q1tASiVxRZmvJDVFA0Pa8F1I0
		I9Iri2+YRM6sGVg8ZkzcCmFd+CoTNy+cw/Y9AQ==`)
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.Write(enc)
	}).Methods("POST")
	h := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	go h.ListenAndServe()
	return h
}

func badKMS(addr string) *http.Server {
	log.Trace("resource/common_test:badKMS() Entering")
	defer log.Trace("resource/common_test:badKMS() Leaving")
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	})
	h := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	go h.ListenAndServe()
	return h
}

func badHVS(addr string) *http.Server {
	log.Trace("resource/common_test:badHVS() Entering")
	defer log.Trace("resource/common_test:badHVS() Leaving")
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	})
	h := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	go h.ListenAndServe()
	return h
}

var saml = `<?xml version="1.0" encoding="UTF-8"?>
<saml2:Assertion ID="MapAssertion" IssueInstant="2019-08-13T20:35:04.312Z" Version="2.0" 
    xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion">
    <saml2:Issuer>https://vs.server.com:8443</saml2:Issuer>
    <Signature xmlns="http://www.w3.org/2000/09/xmldsig#">
        <SignedInfo>
            <CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315#WithComments"/>
            <SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
            <Reference URI="#MapAssertion">
                <Transforms>
                    <Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
                </Transforms>
                <DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
                <DigestValue>nm/o1HX2yhqYwcAVfKYELusc8UdMXOP36XmM+QzjaTo=</DigestValue>
            </Reference>
        </SignedInfo>
        <SignatureValue>FQx+EN6sbjgTPYppa4zXFuerAFaXMrGjiEx7VUm1FBgRWs4eTTDw+hnMUIGy5maGZhuJMxTHCRPM
VnTsAFgSJwrsbT4xVqdR0Pia1GCbQ9pwwO9rubFcXkmbeoSqlZKGlgw0itC4sx/jfJSPRMwcXeEA
U/ikNufVcWfUhPE2+icmpy0NgVz8+WybVs+UDj22sMatD9u3E2rCziuDu3heKvOUfHhIKohoXEBz
Y6aWQ1Q6XMnh9YqBBpV/q+YHUDDABnRGZhrt1+YR4gaXppOKvRYen/VfLa1khaDJOzBPiBlSxzEa
fngWTFj+rq1nJf+IhWFaMacVB2wB3wE7puN2/5M11GT6p0Cy5P1mAKLA/Hf65EtejpyiGFP3YbQl
8hzFlfycLVDtyiwd1khSmVRieYf3Qz0nVcO8oAQMc2w3OtmPRvnFvYKwFaHR80j5Y2DRsWVbqnLF
guGouSQRVoa8UoGl+9jeYZGwE9LpqyHTJbT5yDOCETaBQvdUFRPmVSuI</SignatureValue>
        <KeyInfo>
            <X509Data>
                <X509Certificate>MIIEYzCCAsugAwIBAgIET0rGXTANBgkqhkiG9w0BAQsFADBiMQswCQYDVQQGEwJVUzELMAkGA1UE CBMCQ0ExDzANBgNVBAcTBkZvbHNvbTEOMAwGA1UEChMFSW50ZWwxEjAQBgNVBAsTCU10IFdpbHNv bjERMA8GA1UEAxMIbXR3aWxzb24wHhcNMTkwODA3MTY0OTU4WhcNMjkwODA0MTY0OTU4WjBiMQsw CQYDVQQGEwJVUzELMAkGA1UECBMCQ0ExDzANBgNVBAcTBkZvbHNvbTEOMAwGA1UEChMFSW50ZWwx EjAQBgNVBAsTCU10IFdpbHNvbjERMA8GA1UEAxMIbXR3aWxzb24wggGiMA0GCSqGSIb3DQEBAQUA A4IBjwAwggGKAoIBgQCGD6Wfd4s3bC46uhrVl0hLd/OqpLaAac59mldriEPAHgw8G0DEZaewjVFp ZQoBSELiNQCPp7HVV+0MIsPrIj5Dw6zMNESbDTuRqRQQM9j2D+F47Z61ngeLFjY0Ht/LQvaj1TPq sT6A1Xb624PaD/7yNz78cbrm4rkaaf7ROm1LhUDG1Fd7PAgaAvgxBBHVK+pPLAuTASX288CJ/19c uTv5Odu2V+HXI6lJZpbYxbY+o9cAO872shrBQEJJDa8IMXVHKi9L9xgQSRECiSB2NSb53PExOMuB g5xWA+F8Vic2REcAxvhE+1uvQRCGY/ZBuH6FcmFohjWgafmu8zJqSdL5STIArwYHF12ERhbwF15X H60hbyLBm7WGpSzWkphamFN3mn/qm41+WRA/Vp5uRU5ifeIovHt1lgXgPc6sk5a9H2pFeU/tyunq qgFHHNC1k45Oa5bL8HMBq8j6CfzNbPoPd0bvgXRYpa3dS54NuJctPcHiOmtCqwOSVVxwayJGfqUC AwEAAaMhMB8wHQYDVR0OBBYEFLIGahyJ3x/utXbtVqlolgB8huu9MA0GCSqGSIb3DQEBCwUAA4IB gQAHgktGAWGxj5aGOgxWk7GK9OcEgmmBJRiNfVjsjFwDCAyCP3gFNVDwg67OxbBIo77V5ikey76e lYRbYzsRUWLJ54QnwbPt42aOYTNuDgs97s8H+vEwlBj016cvo0HhslJn2X7EK9eYweZzBZ82KUpC YKhMGyeS6iAAd39iBakqjY0khpJlX+Ti6ITV4ZDilXrK2FWYUvl0ZU2ytaoh8r/s1pJa17VKDgNJ btMvbXae5EYoyVwr/DoYroPTvS01MHOoRmwOxGjRlr4cnTfXmEEZiZuGrvQcyPySdadZK0QHokL0 snXKg0u4YIU60oaTYn0jiQmCn4YACJWScBS9Mm/pO4urXkqj71VqJsHVZxRyUm1Bss+MaCn7JhlY BIsHDnaml3ZyX+KLnv/eTQYsaXeaUk0APdId3nQqiMuFqpZjRdOZrE2Kn7IES2DI/wbnbxnRcLLO AvUXoxs/yIf0UxEMbR77+Z3hHn4YbM3s1Uu2ZqCQmHIhWK1NsD8gYNsuLl8=</X509Certificate>
            </X509Data>
        </KeyInfo>
    </Signature>
    <saml2:Subject>
        <saml2:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified">O23RU15</saml2:NameID>
        <saml2:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:sender-vouches">
            <saml2:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified">Intel Security Libraries</saml2:NameID>
            <saml2:SubjectConfirmationData NotBefore="2019-08-13T20:35:04.312Z" NotOnOrAfter="2019-08-14T20:35:04.312Z"/>
        </saml2:SubjectConfirmation>
    </saml2:Subject>
    <saml2:AttributeStatement>
        <saml2:Attribute Name="biosVersion">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">SE5C620.86B.0X.01.0155.073020181001</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="hostName">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">O23RU15</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="tpmVersion">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">2.0</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="processorInfo">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">54 06 05 00 FF FB EB BF</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="vmmName">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">QEMU</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="hardwareUuid">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">00964993-89C1-E711-906E-00163566263E</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="errorCode">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">0</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="vmmVersion">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">2.10.0</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="osName">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">RedHatEnterpriseServer</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="noOfSockets">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">2</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="tpmEnabled">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="biosName">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">Intel Corporation</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="osVersion">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">7.6</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="processorFlags">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc aperfmperf eagerfpu pni pclmulqdq dtes64 monitor ds_cpl vmx smx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid dca sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch epb cat_l3 cdp_l3 intel_ppin intel_pt ssbd mba ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid fsgsbase tsc_adjust bmi1 hle avx2 smep bmi2 erms invpcid rtm cqm mpx rdt_a avx512f avx512dq rdseed adx smap clflushopt clwb avx512cd avx512bw avx512vl xsaveopt xsavec xgetbv1 cqm_llc cqm_occup_llc cqm_mbm_total cqm_mbm_local dtherm ida arat pln pts hwp hwp_act_window hwp_epp hwp_pkg_req pku ospke spec_ctrl intel_stibp flush_l1d</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="installedComponents">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">[wlagent, tagent]</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="tbootInstalled">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="txtEnabled">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="pcrBanks">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">[SHA1, SHA256]</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="isDockerEnv">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">false</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="FEATURE_TPM">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="FEATURE_TXT">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="TRUST_PLATFORM">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="TRUST_OS">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="TRUST_HOST_UNIQUE">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="TRUST_ASSET_TAG">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">NA</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="TRUST_SOFTWARE">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="TRUST_OVERALL">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">true</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="Binding_Key_Certificate">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">-----BEGIN CERTIFICATE-----&#xd;
MIIFITCCA4mgAwIBAgIJANyKbWsRVtYiMA0GCSqGSIb3DQEBDAUAMBsxGTAXBgNVBAMTEG10d2ls&#xd;
c29uLXBjYS1haWswHhcNMTkwODA3MTcyMzE2WhcNMjkwODA0MTcyMzE2WjAlMSMwIQYDVQQDDBpD&#xd;
Tj1CaW5kaW5nX0tleV9DZXJ0aWZpY2F0ZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEB&#xd;
AJrShQOtV3jknK5vYJNO8HqenzlaKQ26QMoT8McDPbvoNjVxRMQ5/RiDmuGNlnzRez6TMht3kkMu&#xd;
a6aLx4jYDoas8m/AheMFflCQK4jvKqP5UqoGZyHLlVh+v+oHi+gonNkFCYQ8whm6ckxt1XqxyQCV&#xd;
y/KDTNITJZOdwMvMkY5c0OjXRnbXoXi3iODMCVVQO979un6ogBmtnYVr/cIjjZRhJo6DkB16OBzy&#xd;
bDa47noj1H8LeRMttCZVxow39fx+8wcfLcjxcQP8HiY9M9aoK4jPeWN/4KzWAbbEbcaO17t1GrbC&#xd;
JG6W6RQCAsqcE3dLsuDGNTJ/TExmWpyek6yoR6MCAwEAAaOCAdwwggHYMA4GA1UdDwEB/wQEAwIF&#xd;
IDCBnQYHVQSBBQMCKQSBkf9UQ0eAFwAiAAtuQF7sjD4U0B5+h9U5/40rsQtC3IqGsWag4rXk471G&#xd;
zgAEAP9VqgAAAAAAFE6sAAAABQAAAAABAAcAKAAIMgAAIgALpg1+mlPAzBTW+rW2c4JgrvvIVhfv&#xd;
QpNMa09ob+flDY4AIgAL8Q5CCTUFUaLaGV1+JCzkWLhIh+gZIRH/fZ10rnf93dIwggEUBghVBIEF&#xd;
AwIpAQSCAQYAFAALAQB2+KW776Be8GYZIsO3UuRRXvs3I87bJngO0GXOj9EnTISmLJdcIeAlkkKR&#xd;
LvCuDLLlFh1QTDY6lQhaRcT4Q6lRoCM7k4fsGadLXGjT8U85tjNmGGicpAL5vKXeQUhzNVHrjCiq&#xd;
mN5hs5o4YDCDRlzjz6Pc9wqBEUFiyOjg60hnNgFS4s/INFK+rEgoPdVDNz7/dZiR4hNiD/m89ZyS&#xd;
wUFwUoZqTQdBjFfpVDRkBYN1hUUmMQVEFC6xolHgU3CmvwB3NMmoig41BoUv6d8UjmaZsnaer7WZ&#xd;
XJ/sTwLMfhrMgrj/Xp76lfr7VJFB09qaONdKUu1/4wf7I/1hQbmZn0DVMA4GCFUEgQUDAikCBAIA&#xd;
ADANBgkqhkiG9w0BAQwFAAOCAYEAQVSDj4B+4K4+SCjcR4C31jvxh5MqgBsUcZjB3UYayLr59/NQ&#xd;
1SuWwpntYIGqzcGWN8UssqQh8i5cU+mtnf5qNCpDK0FtZeWuSvTok4eOrPxT/jahQnFsYuArNgHJ&#xd;
2HNxanAMWcshCJ4wugEErSo5FiLSFEC4jE16BzDFwXpHfVpocbJTBuHiKu2r1ZWHzaOU5TIsaZn3&#xd;
N2EK7Sqj7rIy0Xi3/lwlFurs3rZOynnsns1yZgMELMVJyP6T+yqUkNSzWGdRki9kOOAh9xVZna/Z&#xd;
Kdztuma5tfljWa5sUnux61J0FinVlQfyDCJYMIT3XQr2Q1s54yxMyrFGVeMJKKh5Ixr6dUEVmF2/&#xd;
rVg9vNuLpHJCkubs32+YMNePLuif/wfoO1zAJuC0vmWCfB2ipjoytJvWdhMxxe5lpIZgT9L1ny53&#xd;
0pHbuDQRICFbUm8DOgnD/NcpCuLy24GJzI7RFDNnqsyJ1rtlUXeja2lzI0b/ZJnISR3S/JZlmxuh&#xd;
2BUWmgMd&#xd;
-----END CERTIFICATE-----</saml2:AttributeValue>
        </saml2:Attribute>
        <saml2:Attribute Name="AIK_Certificate">
            <saml2:AttributeValue xmlns:xsd="http://www.w3.org/2001/XMLSchema" 
                xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xsd:string">-----BEGIN CERTIFICATE-----&#xd; MIIDTjCCAbagAwIBAgIGAWxtGvaGMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNVBAMTEG10d2lsc29u&#xd; LXBjYS1haWswHhcNMTkwODA3MTcyMjU5WhcNMjkwODA2MTcyMjU5WjAAMIIBIjANBgkqhkiG9w0B&#xd; AQEFAAOCAQ8AMIIBCgKCAQEAnhh03DnQYS86N904/QA2VifuavSn3x+uwX1undxzFsp0DqgyjsNc&#xd; Bjt6+Yr9W3bmgEl1nun2ehhA5Q7AO7Jo5mompyn06OucTRvMTssTMwpcNyjfozCHejMK0aG7LVPr&#xd; 7Fk4KDE6ArSSeno/8QPbhW7eueHXJCsQ7rwDWet+sfaBCDRVtbVH5rrzqf2Jw3ekLBkISUuU+xd2&#xd; 0qZsAe/VBlFTFM5vf091OwOiGulpfJhU7d/UYkiiJY0rmG7jF6IZjvsUAWYfwiPdLtXabQI/xaTR&#xd; ekPdTBwFQoC7WMClK7JBwKae1EKyY3VDq63CChUCsgye/y6p/3EaI2NxBgBtpwIDAQABozMwMTAv&#xd; BgNVHREBAf8EJTAjgSEAC/0ANf39A/39VV0cYv39/TH9/V39/f39Ef1JWP0TcP0wDQYJKoZIhvcN&#xd; AQELBQADggGBACNYVC5JT1G2eSxs1Td5yoO4vuIACyqBVYrijDrgr7isRuJmYvimn8vtZEhk4MNP&#xd; jRuZTc0JannDKcySwaeUbN7d17iraDMStmi1i0cIPP5YhXNFszgp4QplFXoyfg5REpjsYV7kTxRo&#xd; AO6fuC2B+5h2kPi+uZwKnvXxgEX6zeX4Qr3h/kFYw1EbDgdHmPQLcs02BMRF8UPFoZAe8wusrTJM&#xd; c/IP+j1mW8h2rAm3YlN+dWkMdU2vtEXuve3zHC1ndRTFEOhAHXfwXM5nRl8nopqPsdP4scEagjZR&#xd; FleODT5JA7QpIwhnnNYTiorghKtK+jNGjrtFE0jXmQWTNDD5IjQ3JJ5luPlpzmM7TzKLMPyl2PcG&#xd; xnjVLebDuZe12aByG0+jgmbHJOe0wVGt2ezajd+zgxXZlgNm/isR/xJd0lNyWnxb+o6ZtqE4TA13&#xd; 5ihJR+qraQT24Vpb9AffL+s3GWDH255YGVEa2+X8q21Uayt5+nasYRL9BK20UI4NCcpEBQ==&#xd;
-----END CERTIFICATE-----</saml2:AttributeValue>
        </saml2:Attribute>
    </saml2:AttributeStatement>
</saml2:Assertion>`
