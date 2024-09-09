package main

import (
	"log"

	"uniswap-v4-rpc/internal/config"
	"uniswap-v4-rpc/internal/ethereum"
	"uniswap-v4-rpc/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	CFG_TEST, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println(CFG_TEST.EthereumNodeURL)
	if err := ethereum.InitClient(CFG_TEST.EthereumNodeURL); err != nil {
		log.Fatalf("Failed to initialize Ethereum client: %v", err)
	}

	if err := ethereum.SetPrivateKey(CFG_TEST.PrivateKey); err != nil {
		log.Fatalf("Failed to set private key: %v", err)
	}

	if err := ethereum.InitContracts(CFG_TEST); err != nil {
		log.Fatalf("Failed to initialize contracts: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	routes.SetupRoutes(router)

	log.Printf("Server starting on %s", CFG_TEST.ServerAddress)
	log.Fatal(router.Run(CFG_TEST.ServerAddress))
}
