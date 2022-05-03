package main

import (
	"github.com/sahithvibudhi/ns3-playground/config"
	"github.com/sahithvibudhi/ns3-playground/pkg/logger"
	"github.com/sahithvibudhi/ns3-playground/pkg/server"
)

func main() {
	logger.Setup()
	config.Setup()
	server.Start()
}
