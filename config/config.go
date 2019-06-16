package config

import "github.com/lillilli/logger"

// Config - service configuration
type Config struct {
	Log  logger.Params
	HTTP HTTPServer

	PrivateKey      string
	EthNodeURL      string
	ContractAddress string
}

// HTTPServer - http server configuration
type HTTPServer struct {
	Host string `default:"0.0.0.0"`
	Port int    `default:"8080"`
}
