package resource

import (
	"encoding/base64"
	"fmt"
	"intel/isecl/workload-service/repository/mock"
	"intel/isecl/workload-service/repository/postgres"
	"net/http"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

func setupServer(t *testing.T) *mux.Router {
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
	wlsDB := postgres.PostgresDatabase{DB: db.Debug()}
	wlsDB.Migrate()
	SetFlavorsEndpoints(r.PathPrefix("/wls/flavors").Subrouter(), wlsDB)
	SetImagesEndpoints(r.PathPrefix("/wls/images").Subrouter(), wlsDB)
	SetReportsEndpoints(r.PathPrefix("/wls/reports").Subrouter(), wlsDB)
	return r
}

func setupMockServer(t *testing.T) *mux.Router {
	r := mux.NewRouter()
	db := mock.DatabaseMock{}
	SetFlavorsEndpoints(r.PathPrefix("/wls/flavors").Subrouter(), db)
	SetImagesEndpoints(r.PathPrefix("/wls/images").Subrouter(), db)
	SetReportsEndpoints(r.PathPrefix("/wls/reports").Subrouter(), db)
	return r
}

func mockHVS(addr string) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/mtwilson/v2/reports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/samlassertion+xml")
		w.Write([]byte(Saml))
	}).Methods("POST")
	h := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	go h.ListenAndServe()
	return h
}

func mockKMS(addr string) *http.Server {
	r := mux.NewRouter()
	r.HandleFunc("/v1/login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
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
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	r := mux.NewRouter()
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

var Saml = `<saml2:Assertion xmlns:saml2="urn:oasis:names:tc:SAML:2.0:assertion" ID="MapAssertion" IssueInstant="2019-01-08T19:09:45.318Z" Version="2.0">
<saml2:Issuer>https://10.105.168.177:8443</saml2:Issuer>
<Signature xmlns="http://www.w3.org/2000/09/xmldsig#">
  <SignedInfo>
	<CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315#WithComments"/>
	<SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"/>
	<Reference URI="#MapAssertion">
	  <Transforms>
		<Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
	  </Transforms>
	  <DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"/>
	  <DigestValue>IqVUXEPGYQF1/5iJ3Yw24YXJ4FeXmQgMRZm+72dEAxo=</DigestValue>
	</Reference>
  </SignedInfo>
  <SignatureValue>YGwRf2gN9U0fiJ8V2OafKpzBTaCNoKcfnphhgdka8DEdw9d/vZWlVSBz2Ex7yB681cxS9GeMqVLh
9Tx61Sq2kkQkvqLOPWBByp7upueJVc5jiqYPNm4U7Dyk42CwemHi66DMAeZwHIMy3Fs8wgfjf1eV
oCXfkDqFcUQfkYRUFpazW/ynkonKA8DhztdX5m5HM3hjnNA7WF6t7cde9Ku8bJRjnDTla7c4BK5a
aPMDQPa4f6sjmayYOPLlKXto1ubEWKOtvpaVcGHOBIQ/n1+bNJHFOajZYwPE4XqJv1ayDuJqcj+7
9bIWyhbjd7OLC5s5EGm9ZT1+pNeoVp9dxPnOCQ==</SignatureValue>
  <KeyInfo>
	<X509Data>
	  <X509Certificate>MIIDYzCCAkugAwIBAgIEa5C0sDANBgkqhkiG9w0BAQsFADBiMQswCQYDVQQGEwJVUzELMAkGA1UE
CBMCQ0ExDzANBgNVBAcTBkZvbHNvbTEOMAwGA1UEChMFSW50ZWwxEjAQBgNVBAsTCU10IFdpbHNv
bjERMA8GA1UEAxMIbXR3aWxzb24wHhcNMTkwMTA1MDM1MjU4WhcNMjkwMTAyMDM1MjU4WjBiMQsw
CQYDVQQGEwJVUzELMAkGA1UECBMCQ0ExDzANBgNVBAcTBkZvbHNvbTEOMAwGA1UEChMFSW50ZWwx
EjAQBgNVBAsTCU10IFdpbHNvbjERMA8GA1UEAxMIbXR3aWxzb24wggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQCKU1xCwHXt1nQwpmB756XDIpY2oZOL3NB3hOXyERTE+wfeTVT4LC85lIEz
2ki9EEu3xgP7QxtLUIpzgFmSAx/fcF7yAqg0Akmnm8zbzHUU3qJce4aofxBOCzlXET/6XHXVhIDR
rm6MGSCiAvKpVcoqx7x9eBl9QA8RzkIxTZ9LBCKCxINDEwcSekHbQVROxfUa8dYuoh4k1Lr59Bnl
/8kry0Exv5uV3rAzXcBepi7V3DJkeq+eVc2sk3j1imLFf9I801BmvAm+A7fH77+Y81TutlwaMt8D
XxXvloiS+kSlTPgy2JpYiDVkU9UqVbvx1AomEO337QMnKGMgMYGRusX5AgMBAAGjITAfMB0GA1Ud
DgQWBBSLi8n0DamJuZQubjhGIxcXxIzY8jANBgkqhkiG9w0BAQsFAAOCAQEAJQfbBm0sNjRhXt6M
4QhX3MSW+A+squM8BUklhvSnRe2dLD0aWp2KPuT2TdgWKUh4gY5Xxygw4jON2HIdGM3Xv/HaV6NT
oZmTGCP6nz8TMbRzlVFpQgZNNUuBUZtyo7A6z8YJF1er7EMW1T8o+KNfa1RxYY9m52JH4HxfBQqL
H6GUP+ich2Xzkoe3EGUeMH1Tq8bYjSdWhhceusNBzRUp+pDgzlKuHyKRyJe+vZ27Axn5ug5a+3Ur
PbE1rN8IqNYeLJrDUgnQpLCDanf99iyYcPM4Ohu2xAptbpOccBTCVRHXO+/cQgLGm+sXZDUN9RSB
L62GvU1C49S1374L2l8nuw==</X509Certificate>
	</X509Data>
  </KeyInfo>
</Signature>
<saml2:Subject>
  <saml2:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified">Purley21</saml2:NameID>
  <saml2:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:sender-vouches">
	<saml2:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified">Intel Security Libraries</saml2:NameID>
	<saml2:SubjectConfirmationData NotBefore="2019-01-08T19:09:45.318Z" NotOnOrAfter="2019-01-09T19:09:45.318Z"/>
  </saml2:SubjectConfirmation>
</saml2:Subject>
<saml2:AttributeStatement>
  <saml2:Attribute Name="biosVersion">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">SE5C620.86B.00.01.0014.070920180847</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="hostName">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Purley21</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="tpmVersion">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">2.0</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="processorInfo">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">54 06 05 00 FF FB EB BF</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="vmmName">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Docker</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="hardwareUuid">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">0030A847-D4B7-E811-906E-00163566263E</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="vmmVersion">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">1.13.1</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="osName">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">RedHatEnterpriseServer</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="noOfSockets">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">2</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="tpmEnabled">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="biosName">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">Intel Corporation</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="osVersion">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">7.5</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="processorFlags">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">fpu vme de pse tsc msr pae mce cx8 apic sep mtrr pge mca cmov pat pse36 clflush dts acpi mmx fxsr sse sse2 ss ht tm pbe syscall nx pdpe1gb rdtscp lm constant_tsc art arch_perfmon pebs bts rep_good nopl xtopology nonstop_tsc aperfmperf eagerfpu pni pclmulqdq dtes64 monitor ds_cpl vmx smx est tm2 ssse3 sdbg fma cx16 xtpr pdcm pcid dca sse4_1 sse4_2 x2apic movbe popcnt tsc_deadline_timer aes xsave avx f16c rdrand lahf_lm abm 3dnowprefetch epb cat_l3 cdp_l3 intel_ppin intel_pt ssbd mba ibrs ibpb stibp tpr_shadow vnmi flexpriority ept vpid fsgsbase tsc_adjust bmi1 hle avx2 smep bmi2 erms invpcid rtm cqm mpx rdt_a avx512f avx512dq rdseed adx smap clflushopt clwb avx512cd avx512bw avx512vl xsaveopt xsavec xgetbv1 cqm_llc cqm_occup_llc cqm_mbm_total cqm_mbm_local dtherm ida arat pln pts hwp hwp_act_window hwp_epp hwp_pkg_req pku ospke spec_ctrl intel_stibp flush_l1d</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="txtEnabled">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="pcrBanks">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">[SHA1, SHA256]</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="FEATURE_TPM">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="FEATURE_TXT">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="FEATURE_CBNT">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">false</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="FEATURE_cbntProfile">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string"/>
  </saml2:Attribute>
  <saml2:Attribute Name="FEATURE_SUEFI">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">false</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="TRUST_PLATFORM">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="TRUST_OS">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="TRUST_HOST_UNIQUE">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="TRUST_ASSET_TAG">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">NA</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="TRUST_SOFTWARE">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="TRUST_OVERALL">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">true</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="Binding_Key_Certificate">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">-----BEGIN CERTIFICATE-----&#13;
MIIEoDCCA4igAwIBAgIJAK8VrKjS0GZvMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNVBAMTEG10d2ls&#13;
c29uLXBjYS1haWswHhcNMTkwMTA4MDkyMTQwWhcNMjkwMTA1MDkyMTQwWjAlMSMwIQYDVQQDDBpD&#13;
Tj1CaW5kaW5nX0tleV9DZXJ0aWZpY2F0ZTCCASEwDQYJKoZIhvcNAQEBBQADggEOADCCAQkCggEA&#13;
AQDWAzVmzRYnEEjg1orKesec++miHfMczKualq1sYza6uX/TJe/z+RaSeRGLFdOYrSD27LsbJZW0&#13;
DgF8D6P8D2cGLg9HvL2oANyd6Jsn0N1LFMOUx5xxc14p8UPu4855YxeiEfm5Dg0eVGj1BF8QuqQ5&#13;
+igZANSfdezsqc4JMh6rO9siOYjvnGKfjiA5RJcxh38uuku+iaDJUOJ1HhYNSeAl1vfTmzjSKc52&#13;
6a5fIyuHHJOXgdfoxfDCjBuIiO8aGo33uV+MPoZ5/IDbv0bP3Fsq2s1M3kPyETpZOAUsi6qol5/q&#13;
lxJ2j88+KKX0PHAKD4P+NvqCtljx1tEcOkYHbAIDAQABo4IB3DCCAdgwDgYDVR0PAQH/BAQDAgUg&#13;
MIGdBgdVBIEFAwIpBIGR/1RDR4AXACIACxxbGub15/KbDRgV9eb3hoIdlCF3eXS6yS5uIBvo2iNQ&#13;
AAQA/1WqAAAAAAViegwAAAACAAAAAQEABwA+AAw2AAAiAAuWghM0SM2fWaBqN6W8z2Cr0tExrj2T&#13;
OZ/F230qS1uncAAiAAv7tdCD86BTQJVU/S620UQ0kN2Vtar9XEplk+SExN+4ojCCARQGCFUEgQUD&#13;
AikBBIIBBgAUAAsBAIShMcoXHvgGMwXWFtn1Wlu6/lsSChE6wir/0xnuM0ivUaUDK58hrNMjY8LU&#13;
LzKweI32MFFx44Z2dV2Oy/fewiQCLi5+RhJMzwPXD4GA2aSHdC5GBVo+eZ8DcOtU2Zfa3+bCWWKl&#13;
0WtD9xCw2WpolZgDoAfMOQUDP2W5p3Tn9ldVZkRO3E0qjXbLEj19t6wZxu69X+XkA0tL/9W//CYD&#13;
E3n16gE+X/rv6mx2NlXlB+oxshkmnyp6PV0LYFbYxk3Trq92Vj62E5OoodIajUzdWJAHHAy6zQEE&#13;
UBpkPVX30RIFk8BUyac4voTbs+bAvCHUsqsCvzJM1C0SqR33w6v4AW4wDgYIVQSBBQMCKQIEAgAA&#13;
MA0GCSqGSIb3DQEBCwUAA4IBAQB9rYmAz9QUbYQ3WoffWz/7OLlygeof0O6lzIHQ2upi/3GChWNL&#13;
maEivV4vC1G0K1cQvfmcKpJs0y0jr/Dd+Fh4FCjyPAuMFzWunLqkv3IXUsxjOCXDa1zv5+z3W9+E&#13;
iUeNsyvgW6zMculGC62eUNxOoD47RPXkF4BeuBRoX1piEbFS+1y6kr4j9llyWXUH/AqEs+/c6nYn&#13;
5+4flowtJYEvJKL3H99pVBvj57/uyKNXUIjTYhLcfhGKijaWU/Mb4UQpVMTZLaFhLMLq1B1XI84T&#13;
nIdZuo6IEg4j+QDob2WmDMDWsMVAqPXySknkGuaO4s9Gi6kGCfFQtxw+n9icX48P&#13;
-----END CERTIFICATE-----</saml2:AttributeValue>
  </saml2:Attribute>
  <saml2:Attribute Name="AIK_Certificate">
	<saml2:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">-----BEGIN CERTIFICATE-----&#13;
MIICzzCCAbegAwIBAgIGAWgnf/BWMA0GCSqGSIb3DQEBCwUAMBsxGTAXBgNVBAMTEG10d2lsc29u&#13;
LXBjYS1haWswHhcNMTkwMTA3MDg0ODQyWhcNMjkwMTA2MDg0ODQyWjAAMIIBIjANBgkqhkiG9w0B&#13;
AQEFAAOCAQ8AMIIBCgKCAQEA07F4kytDEtufc261WH+pCweZsbQDqzU+hQv2NPNsPQnfY1ZN3uEX&#13;
H105d9U2C1bEU8GFsc4pXU7bKFois1yDb5brTYu7aQCUKAlESOjtIqxO+Yx/tZK/Hzkl/svT8Z9V&#13;
Qm0yhMO9n1leBn1bvVc+8lISxWF1FHLOTFAOjHvy3aUjoegvP79rfcng+N0FOXcDmX49TPCbjajd&#13;
m2DDC/T+1mQ+A7KE8Oax3GY0EqriI7cCla601H6JQKqJpzZiyIRiGQlKKDWd2X8LtQKu3pRQcw5h&#13;
KeT8Rb1tS4t2II5UZwDLVUwvRyIJ54UcaPws8yrrfmVsiZvHl4ITghEfA9o9GwIDAQABozQwMjAw&#13;
BgNVHREBAf8EJjAkgSIACyH9/Un9Zv0MFkv9/R97Wv39/f1X/TT9LP39L/08Qv1WMA0GCSqGSIb3&#13;
DQEBCwUAA4IBAQBdUZP1BrpAYM27E5GMAMHI/JpsC2vRfgOv6/GLz5Uh1EILlqt89EE9PEO0Pd96&#13;
tYqSzgv5lSCDPDM1pfBuOGC4kh+CsfH8b7zAsFeSZad4nxMugw/mE17btREOpoYWewb4cPVgJTd3&#13;
t0I0gLEBGFJ6KX4hRESXrOZA+2G96CCZlsWcy8T538gpn6ZmhQKP7k/GZd5yo+TKVE1/YXgUewTF&#13;
QtVe5iBZelgdVlU1ZJy1BAHTmnRO9XmVSRKzenDO3nmLSx2iKJm8DnG0n+Or1W3qzzqvBdA4XkQa&#13;
O88I2zpgmKrUOyQq+G4APRSXrOlVNV+7B8nj5tlFYcaZR6/9YDRj&#13;
-----END CERTIFICATE-----</saml2:AttributeValue>
  </saml2:Attribute>
</saml2:AttributeStatement>
</saml2:Assertion>`
