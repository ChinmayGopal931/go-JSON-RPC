package main

import (
	"log"

	"uniswap-v4-rpc/internal/config"
	"uniswap-v4-rpc/internal/ethereum"
	"uniswap-v4-rpc/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println(cfg.EthereumNodeURL)
	if err := ethereum.InitClient(cfg.EthereumNodeURL); err != nil {
		log.Fatalf("Failed to initialize Ethereum client: %v", err)
	}

	if err := ethereum.SetPrivateKey(cfg.PrivateKey); err != nil {
		log.Fatalf("Failed to set private key: %v", err)
	}

	if err := ethereum.InitContracts(cfg); err != nil {
		log.Fatalf("Failed to initialize contracts: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	routes.SetupRoutes(router)

	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Fatal(router.Run(cfg.ServerAddress))
}
