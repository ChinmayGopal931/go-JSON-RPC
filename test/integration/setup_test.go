package integration

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"uniswap-v4-rpc/internal/config"
	"uniswap-v4-rpc/internal/ethereum"
	"uniswap-v4-rpc/internal/routes"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

var (
	testServer *httptest.Server
	router     *gin.Engine
)

func setup() {
	log.Println("Starting setup...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Println("Configuration loaded successfully")

	// Initialize Ethereum client
	err = ethereum.InitClient(cfg.EthereumNodeURL)
	if err != nil {
		log.Fatalf("Failed to initialize Ethereum client: %v", err)
	}
	log.Println("Ethereum client initialized successfully")

	// Initialize contract addresses
	err = ethereum.InitContracts(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize contract addresses: %v", err)
	}
	log.Println("Contract addresses initialized successfully")

	if err := ethereum.SetPrivateKey(cfg.PrivateKey); err != nil {
		log.Fatalf("Failed to set private key: %v", err)
	}

	// Set up the Gin router
	router = gin.Default()
	routes.SetupRoutes(router)

	// Create a test server
	testServer = httptest.NewServer(router)

	log.Println("Setup completed")
}
