package config

import "os"

var Config TConfig

func Setup() {
	Config = TConfig{}
	Config.Parse()
}

type TServer struct {
	Port string
}

type TDocker struct {
	Image string
}

type TConfig struct {
	Server TServer
	Docker TDocker
}

func (c *TConfig) Parse() {
	server := TServer{}
	server.Port = os.Getenv("SERVER_PORT")
	c.Server = server
}
