package main

import (
	"github.com/sahithvibudhi/ns3-playground/config"
	"github.com/sahithvibudhi/ns3-playground/pkg/server"
)

func main() {
	config.Setup()
	server.Start()
}
