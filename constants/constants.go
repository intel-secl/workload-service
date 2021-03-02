/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package constants

import "time"

const (
	ServiceName               = "WLS"
	ExplicitServiceName       = "Workload Service"
	ApiVersion                = "v1"
	ConfigDir                 = "/etc/workload-service/"
	ConfigFile                = ConfigDir + "config.yml"
	LogDir                    = "/var/log/workload-service/"
	LogFile                   = LogDir + "wls.log"
	SecurityLogFile           = LogDir + "wls-security.log"
	TrustedJWTSigningCertsDir = ConfigDir + "certs/trustedjwt/"
	TrustedCaCertsDir         = ConfigDir + "certs/trustedca/"
	DefaultTLSCertPath        = ConfigDir + "tls-cert.pem"
	DefaultTLSKeyPath         = ConfigDir + "tls.key"
	CertApproverGroupName     = "CertApprover"
	DefaultWlsTlsCn           = "WLS TLS Certificate"
	DefaultWlsTlsSan          = "127.0.0.1,localhost"
	DefaultKeyAlgorithm       = "rsa"
	DefaultKeyAlgorithmLength = 3072
	DefaultSSLCertFilePath    = ConfigDir + "wlsdbsslcert.pem"
	ReportCreationGroupName   = "ReportsCreate"
	BearerToken               = "BEARER_TOKEN"
	CmsBaseUrlEnv             = "CMS_BASE_URL"
	WlsTLsCertCnEnv           = "WLS_TLS_CERT_CN"
	WlsCertSANList            = "SAN_LIST"
	DefaultKeyCacheSeconds    = 300
	KeyCacheSeconds           = "KEY_CACHE_SECONDS"
	CmsTlsCertDigestEnv       = "CMS_TLS_CERT_SHA384"
	LogEntryMaxlengthEnv      = "LOG_ENTRY_MAXLENGTH"
	JWTCertsCacheTime         = "1m"
	HttpLogFile               = "/var/log/workload-service/http.log"
	SamlCaCertFilePath        = TrustedCaCertsDir + "SamlCaCert.pem"
	DefaultReadTimeout        = 30 * time.Second
	DefaultReadHeaderTimeout  = 10 * time.Second
	DefaultWriteTimeout       = 10 * time.Second
	DefaultIdleTimeout        = 10 * time.Second
	DefaultMaxHeaderBytes     = 1 << 20
	DefaultWLSListenerPort    = 5000
	DBTypePostgres            = "postgres"
	DefaultLogEntryMaxlength  = 300
	WLSRuntimeUser            = "wls"
	WLSRuntimeGroup           = "wls"
	WLSConsoleEnableEnv       = "WLS_ENABLE_CONSOLE_LOG"
)

//Resource endpoints
const (
	KeyEndpoint   = "resource/keys"
	ImageEndpoint = "resource/images"
)

//Roles and permissions
const (
	FlavorsRetrieve = "flavors:retrieve"
	FlavorsSearch   = "flavors:search"
	FlavorsCreate   = "flavors:create"
	FlavorsDelete   = "flavors:delete"

	ImageFlavorsRetrieve = "image_flavors:retrieve"
	ImageFlavorsSearch   = "image_flavors:search"
	ImageFlavorsStore    = "image_flavors:store"
	ImageFlavorsDelete   = "image_flavors:delete"

	ImagesRetrieve = "images:retrieve"
	ImagesSearch   = "images:search"
	ImagesCreate   = "images:create"
	ImagesDelete   = "images:delete"

	ReportsSearch = "reports:search"
	ReportsCreate = "reports:create"
	ReportsDelete = "reports:delete"

	KeysCreate = "keys:create"
)

// State represents whether or not a daemon is running or not
type State bool

const (
	// Stopped is the default nil value, indicating not running
	Stopped State = false
	// Running means the daemon is active
	Running State = true
)
