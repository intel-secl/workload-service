package config

// Config is a global structure of config values
var Port int
var UseTLS bool
var DatabasePass string

func init() {
	Port = 8444
	UseTLS = false
}
