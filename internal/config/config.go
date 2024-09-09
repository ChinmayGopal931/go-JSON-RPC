package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	EthereumNodeURL   string `mapstructure:"ethereum_node_url"`
	ServerHost        string `mapstructure:"server_host"`
	ServerPort        int    `mapstructure:"server_port"`
	PrivateKey        string `mapstructure:"private_key"`
	ServerAddress     string
	SwapRouterAddress string `mapstructure:"swap_router_address"`
	LPRouterAddress   string `mapstructure:"lp_router_address"`
	ManagerAddress    string `mapstructure:"manager_address"`
	HookAddress       string `mapstructure:"hook_address"`
	Token0_address    string `mapstructure:"token0_address"`
	Token1_address    string `mapstructure:"token1_address"`
}

func Load() (*Config, error) {
	// Get the current working directory
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Set the path to the config file
	configPath := filepath.Join(workDir, "config.yaml")

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// Try to read the config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found at %s: %w", configPath, err)
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	// Compute ServerAddress
	cfg.ServerAddress = fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)

	log.Printf("Config loaded successfully. Ethereum Node URL: %s", cfg.EthereumNodeURL)

	return &cfg, nil
}
