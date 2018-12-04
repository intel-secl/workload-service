package main

// Config is a global structure of config values
var Config struct {
	Port int
}

func init() {
	Config.Port = 8444
}
