package setup

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"
	aasClient "intel/isecl/lib/clients/aas"
	aasTypes "intel/isecl/lib/common/types/aas"
	cos "intel/isecl/lib/common/os"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/lib/common/crypt"
	"intel/isecl/workload-service/config"
	consts "intel/isecl/workload-service/constants"
	"io/ioutil"
        "net/http"
)

// AASConnection is a setup task for setting roles in AAS
type AASConnection struct{}

// Run will run the HVS Connection setup task, but will skip if Validate() returns no errors
func (aas AASConnection) Run(c csetup.Context) error {

	fmt.Println("Setting up roles in AAS ...")
	var aasURL string
	var aasBearerToken string
	var err error
	if aasURL, err = c.GetenvString(config.AAS_API_URL, "AAS Server URL"); err != nil {
		return err
	}
        if strings.HasSuffix(aasURL, "/") {
                config.Configuration.AAS_API_URL = aasURL
        } else {
                config.Configuration.AAS_API_URL = aasURL + "/"
        }

	config.Save()
	if aasBearerToken, err = c.GetenvString(consts.BearerToken, "AAS Bearer Token"); err != nil {
		return err
	}

	ac := &aasClient.Client{
		BaseURL:  aasURL,
		JWTToken: []byte(aasBearerToken),
	}

	roles := [3]string{consts.FlavorImageRetrievalGroupName, consts.ReportCreationGroupName, consts.AdministratorGroupName}

	var role_ids []string
	for _, role := range roles {
		roleCreate := aasTypes.RoleCreate{
			Name:    role,
			Service: consts.ServiceName,
		}
		roleCreateResponse, err := ac.CreateRole(roleCreate)
		if err != nil {
			if(strings.Contains(err.Error(), "same role exists")) {
				continue
			}
			fmt.Printf("%v", err)
			return err
		}

		role_ids = append(role_ids, roleCreateResponse.ID)
	}
	
	//Fetch JWT Certificate from AAS	
	err = fnGetJwtCerts()
	if err != nil{
		return err
	}

	return nil
}

// Validate checks whether or not the KMS Connection setup task was completed successfully
func (aas AASConnection) Validate(c csetup.Context) error {
	return nil
}

func fnGetJwtCerts() error {

        url := config.Configuration.AAS_API_URL + "noauth/jwt-certificates"
	req, _ := http.NewRequest("GET", url, nil)
        req.Header.Add("accept", "application/x-pem-file")
        rootCaCertPems, err := cos.GetDirFileContents(consts.TrustedCaCertsDir, "*.pem" )
        if err != nil {
                return err
        }

        // Get the SystemCertPool, continue with an empty pool on error
        rootCAs, _ := x509.SystemCertPool()
        if rootCAs == nil {
                rootCAs = x509.NewCertPool()
        }
        for _, rootCACert := range rootCaCertPems{
                if ok := rootCAs.AppendCertsFromPEM(rootCACert); !ok {
                        return err
                }
        }
        httpClient := &http.Client{
                                Transport: &http.Transport{
                                        TLSClientConfig: &tls.Config{
                                                InsecureSkipVerify: false,
                                                RootCAs: rootCAs,
                                                },
                                        },
                                }

        res, err := httpClient.Do(req)
        if err != nil {
                return fmt.Errorf("Could not retrieve jwt certificate")
        }
        defer res.Body.Close()
        body, _ := ioutil.ReadAll(res.Body)
	err = crypt.SavePemCertWithShortSha1FileName(body, consts.TrustedJWTSigningCertsDir)
	if err != nil {
		fmt.Println("Could not store Certificate")
		return fmt.Errorf("Certificate setup: %v", err)
	}

        return nil
}

